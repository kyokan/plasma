const async = require('async');
const protoLoader = require('@grpc/proto-loader');
const grpc = require('grpc');
const _ = require('lodash');
const ejs = require('ethereumjs-util');
const etx = require('ethereumjs-tx');
const web3 = require('web3');
const BN = require('bn.js');
const secp256k1 = require('secp256k1');

class PlasmaClient {
    constructor(protoFile, url) {
        let packageDefinition = protoLoader.loadSync(
            protoFile,
            {
                keepCase: true,
                longs: String,
                enums: String,
                defaults: true,
                oneofs: true
            });
        let pb = grpc.loadPackageDefinition(packageDefinition).pb;
        this.client = new pb.Root(url, grpc.credentials.createInsecure());

        let definitions = this.client.$method_definitions;
        if (_.isEmpty(definitions)) {
            throw new Error('Failed to read the definitions');
        }
        let l = Object.keys(definitions).length;
        if (l != 4) {
            throw new Error(`Got ${l} definitions, expecting 4`);
        }

        let GetBalanceFtor = definitions.GetBalance;
        let GetUTXOsFtor   = definitions.GetUTXOs;
        let GetBlockFtor   = definitions.GetBlock;
        let SendFtor       = definitions.Send;

        if (_.isEmpty(GetBalanceFtor)) {
            throw new Error('GetBalance definition is missing');
        }
        if (_.isEmpty(GetUTXOsFtor)) {
            throw new Error('GetUTXOs definition is missing');
        }
        if (_.isEmpty(GetBlockFtor)) {
            throw new Error('GetBlock definition is missing');
        }
        if (_.isEmpty(SendFtor)) {
            throw new Error('Send definition is missing');
        }
    }

    GetBalance(address, cb) {
        this.client.getBalance({address: address}, cb);
    }

    GetUTXOs(address, cb) {
        this.client.GetUTXOs({address: address, spendable: true}, cb);
    }

    GetBlock(number, cb) {
        this.client.getBlock({number: number}, cb);
    }

    Send(params, cb) {
        this.client.send(params, cb);
    }
}

class Account {
    constructor(client, web3, contract, address, key) {
        this.client = client;
        this.web3 = web3;
        this.contract = contract;
        this.address = address;
        this.privateKey = ejs.toBuffer(`0x${key}`);
    }

    GetPlasmaBalance(cb) {
        this.client.GetBalance(this.web3.utils.hexToBytes(this.address), (err, result) => {
            if (err !== null) {
                return cb(err);
            }
            const balance = toBN(result.balance);
            return cb(null, balance);
        });
    }

    GetPlasmaUTXOs(cb) {
        this.client.GetUTXOs(this.web3.utils.hexToBytes(this.address), (err, utxos) => {
            cb(err, utxos);
        });
    }

    GetPlasmaBlock(number, cb) {
        this.client.GetBlock(number, cb);
    }

    GetBlock(number, cb) {
        this.contract.methods.getBlock(number).call(cb);
    }

    prepareEthCall(cb) {
        let self = this;
        async.parallel([
            (asyncCB) => {
                self.web3.eth.estimateGas({ from: self.address }, asyncCB);
            },
            (asyncCB) => {
                self.web3.eth.getGasPrice(asyncCB);
            },
            (asyncCB) => {
                self.web3.eth.getTransactionCount(self.address, asyncCB)
            }
        ], (err, results) => {
            if (err != null) {
                return cb(err);
            }
            return cb(null, {
                gasEstimate: 2 * results[0],
                gasPrice: results[1],
                nonce: results[2]
            });
        });
    }

    Send(to, amount, cb) {
        let self = this;
        this.client.GetUTXOs(this.web3.utils.hexToBytes(this.address), (err, result) => {
            if (err != null) {
                return cb(err);
            }
            const utxos = result.transactions;
            let tx = self.transactionForAmount(to, amount, utxos);
            try {
                const rpcTx = tx.toRpc();
                // console.log(`Transaction RPC: ${JSON.stringify(rpcTx, null, 2)}`);
                this.client.Send({transaction: rpcTx}, (sendErr, sendRes) => {
                    if (sendErr != null) {
                        return cb(sendErr);
                    }
                    // TODO: verify root's signature
                    return cb(null, new Transaction(sendRes.transaction));
                });
            }
            catch (e) {
                return cb(e);
            }
        });
    }

    signTransaction(tx) {
        let result = new Transaction(tx);
        // This will throw if key is invalid
        const signature = secp256k1.sign(result.SignatureHash(), this.privateKey);
        result.sig0 = Buffer.from(signature.signature);
        result.sig1 = Buffer.from(signature.signature);
        // console.log(`Signed transaction ${result.toString()}`);
        return result;
    }

    transactionForAmount(to, amount, utxos) {
        if (!_.isArray(utxos) || utxos.length == 0) {
            throw new Error('Invalid input');
        }
        let result = new Transaction();
        // Construct first output
        result.output0 = new Output(to, amount);
        let transactions = new Array(utxos.length);
        let sorter = [];
        for (let index = 0; index < utxos.length; index++) {
            const tx = new Transaction(utxos[index]);
            transactions[index] = tx;
            const addr0 = ejs.bufferToHex(tx.output0.newOwner);
            const addr1 = ejs.bufferToHex(tx.output1.newOwner);
            if (addr0 === this.address) {
                if (tx.output0.amount.cmp(amount) == 0) {
                    result.input0 = new Input(tx.BlockNum, tx.TxIdx, 0);
                    // console.log(`Using input ${tx.toString()}, output 0`);
                    return this.signTransaction(result);
                }
                sorter = sorter.concat([{amount: tx.output0.amount, index: index, outIdx: 0}]);
            }
            if (addr1 === this.address) {
                if (tx.output1.amount.cmp(amount) == 0) {
                    result.input0 = new Input(tx.BlockNum, tx.TxIdx, 1);
                    // console.log(`Using input ${tx.toString()}, output 1`);
                    return this.signTransaction(result);
                }
                sorter = sorter.concat([{amount: tx.output1.amount, index: index, outIdx: 1}]);
            }
        }
        sorter.sort((lhs, rhs) => {
            return lhs.amount.cmp(rhs.amount);
        });

        const min = sorter[0];
        if (min.amount.cmp(amount) == 1) { // min amount is greater than amount, no need to search
            const inputTx = transactions[min.index];
            let inputAmount = inputTx.output0.amount;
            if (min.outIdx == 1) {
                inputAmount = inputTx.output1.amount;
            }
            result.output1 = new Output(this.address, inputAmount.sub(amount));
            result.input0 = new Input(inputTx.BlockNum, inputTx.TxIdx, min.outIdx);
            // console.log(`Using input ${inputTx.toString()}, output ${min.outIdx}`);
            return this.signTransaction(result);
        }
        let leftBound = 0;
        let rightBound = sorter.length - 1;
        let lhs = -1;
        let rhs = -1;
        for (; leftBound < rightBound; ) {
            const sum = sorter[leftBound].amount.add(sorter[rightBound].amount);
            const compare = sum.cmp(amount);
            if (compare == 0) { // sum == amount
                break;
            }
            if (compare == -1) { // sum < amount
                leftBound++;
                continue;
            }
            // keep track of last sum greater than amount
            lhs = leftBound;
            rhs = rightBound;
            rightBound--; // sum > amount
        }
        if (leftBound < rightBound) { // found a pair that sums to amount
            const inputTx0 = transactions[sorter[leftBound].index];
            const inputTx1 = transactions[sorter[rightBound].index];
            result.input0 = new Input(inputTx0.BlockNum, inputTx0.TxIdx, sorter[leftBound].outIdx);
            result.input1 = new Input(inputTx1.BlockNum, inputTx1.TxIdx, sorter[rightBound].outIdx);

            // console.log(`Using input ${inputTx0.toString()}, output ${sorter[leftBound].outIdx}`);
            // console.log(`Using input ${inputTx1.toString()}, output ${sorter[rightBound].outIdx}`);

            return this.signTransaction(result);
        }
        if (lhs >= 0 && rhs >= 0) {
            const inputTx0 = transactions[sorter[lhs].index];
            const inputTx1 = transactions[sorter[lhs].index];
            result.input0 = new Input(inputTx0.BlockNum, inputTx0.TxIdx, sorter[lhs].outIdx);
            result.input1 = new Input(inputTx1.BlockNum, inputTx1.TxIdx, sorter[rhs].outIdx);
            let inputAmount = inputTx0.output0.amount;
            if (sorter[lhs].outIdx == 1) {
                inputAmount = inputTx0.output1.amount;
            }
            if (sorter[rhs].outIdx == 0) {
                inputAmount = inputAmount.add(inputTx1.output0.amount);
            } else {
                inputAmount = inputAmount.add(inputTx1.output1.amount);
            }
            result.output1 = new Output(this.address, inputAmount.sub(amount));

            // console.log(`Using input ${inputTx0.toString()}, output ${sorter[lhs].outIdx}`);
            // console.log(`Using input ${inputTx1.toString()}, output ${sorter[rhs].outIdx}`);

            return this.signTransaction(result);
        }

        throw new Error('no suitable UTXOs found');
    }

    Deposit(amount, cb) {
        let self = this;
        this.prepareEthCall( (err, result) => {
            if (err != null) {
                return cb(err, null);
            }
            let tx = new Transaction({
                output0: new Output(self.address, amount),
                });
            const rlpEncoded = tx.RLPEncode();
            let nonce = result.nonce;
            // Retrying as sometimes the call fails with invalid RPC response error
            let depositFn = function (callback) {
                let params = {
                    nonce:    self.web3.utils.toHex(nonce),
                    chainId:  15,
                    to:       self.contract.options.address,
                    value:    amount,
                    gasPrice: self.web3.utils.toHex(result.gasPrice),
                    gas:      self.web3.utils.toHex(result.gasEstimate), // just to be safe
                    from:     self.address
                };
                self.contract.methods.deposit(rlpEncoded).send(params, (error, receipt) => {
                    if (error != null && !error.message.startsWith('Error: Invalid JSON RPC response')) {
                        nonce++;
                    }
                    callback(error, receipt);
                });
            };
            async.retry({times: 3}, depositFn, (error, receipt) => {
                return cb(error, receipt);
            });
        });
    }

    Send(to, amount, cb) {
        let self = this;

    }
}

const transaction_rpc_format = {
    "input0": {
        "blockNum": "0",
        "txIdx": 0,
        "outIdx": 0
    },
    "sig0": {
        "type": "Buffer",
        "data": []
    },
    "input1": {
        "blockNum": "0",
        "txIdx": 0,
        "outIdx": 0
    },
    "sig1": {
        "type": "Buffer",
        "data": []
    },
    "output0": {
        "newOwner": {
            "type": "Buffer",
            "data": [98, 115, 6, 9, 10, 186, 179, 166, 225, 64, 14, 147, 69, 188, 96, 199, 138, 139, 239, 87
            ]
        },
        "amount": {
            "values": {
                "type": "Buffer",
                "data": [13, 224, 182, 179, 167, 100, 0, 0
                ]
            }
        }
    },
    "output1": {
        "newOwner": {
            "type": "Buffer",
            "data": [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
            ]
        },
        "amount": {
            "values": {
                "type": "Buffer",
                "data": []
            }
        }
    },
    "fee": {
        "values": {
            "type": "Buffer",
            "data": []
        }
    },
    "BlockNum": "2",
    "TxIdx": 0,
    "RootSig": {
        "type": "Buffer",
        "data": [157, 128, 191, 87, 94, 141, 243, 153, 49, 4, 199, 92, 215, 150, 160, 77, 11, 42, 191, 208, 85,
            237, 63, 238, 251, 81, 253, 181, 230, 50, 219, 101, 51, 220, 32, 238, 167, 76, 151, 136, 23, 254, 209,
            116, 188, 222, 128, 58, 51, 188, 212, 185, 230, 233, 213, 10, 110, 118, 43, 22, 179, 44, 97, 232, 1
        ]
    }
};
transaction_rpc_format;

function hash(input) {
    const hash = ejs.sha256(input);
    return hash;
}

class Input {
    zero() {
        this.blockNum = "0";
        this.txIdx    =  0;
        this.outIdx   =  0;
    }

    constructor (...args) {
        this.zero();
        if (args.length == 0) {
            return;
        }
        if (args.length == 1) {
            const other = args[0];
            if (_.isEmpty(other)) {
                return;
            }
            this.blockNum = other.blockNum;
            this.txIdx    = other.txIdx;
            this.outIdx   = other.outIdx;
            return;
        }
        if (args.length == 3) {
            this.blockNum = args[0];
            this.txIdx    = args[1];
            this.outIdx   = args[2];
            return;
        }
        throw new Error("Invalid number of arguments for constructing Input instance");
    }

    Hash() {
        const blockNumBuf = ejs.toBuffer(this.blockNum);
        const txIdxBuf    = ejs.toBuffer(this.txIdx);
        const outIdxBuf   = ejs.toBuffer(this.outIdx);
        const buf = Buffer.concat([blockNumBuf, txIdxBuf, outIdxBuf], blockNumBuf.length + txIdxBuf.length + outIdxBuf.length);
        return hash(buf);
    }

    toRpc() {
        return this;
    }

    toString() {
        return this;
    }

}

class Output {
    zero() {
        this.newOwner = Buffer.from(new Uint8Array(20));
        this.amount   = Buffer.from(new Uint8Array(0));
    }

    constructor(...args) {
        this.zero();
        if (args.length == 0) {
            return;
        }
        if (args.length == 1) {
            const other = args[0];
            if (_.isEmpty(other)) {
                return;
            }
            this.newOwner = other.newOwner;
            this.amount   = toBN(other.amount);

            return;
        }
        if (args.length == 2) {
            this.newOwner = args[0];
            this.amount   = args[1];
            return;
        }
        throw new Error("Invalid number of arguments for constructing Output instance");
    }

    Hash() {
        const newOwnerBuf = ejs.toBuffer(this.newOwner);
        const amountBuf   = ejs.toBuffer(this.amount);
        const buf = Buffer.concat([newOwnerBuf, amountBuf], newOwnerBuf.length + amountBuf.length);
        return hash(buf);
    }

    toRpc() {
        return {
            newOwner: ejs.toBuffer(this.newOwner),
            amount: {
                values: this.amount.toBuffer('be', this.amount.byteLength())
            }
        };
    }

    toString() {
        return {
            newOwner: this.newOwner,
            amount: this.amount.toString(10)
        };
    }
}


function hash(input) {
    const hash = ejs.sha256(input);
    return hash;
}

class Transaction {
    zero() {
        this.input0 = new Input();
        this.input1 = new Input();
        this.output0 = new Output();
        this.output1 = new Output();

        this.BlockNum = "0";
        this.TxIdx = 0;
        this.sig0 = Buffer.from("");
        this.sig1 = Buffer.from("");
        this.RootSig = Buffer.from("");
        this.fee = Buffer.from("");
    }

    constructor(other) {
        this.zero();
        if (_.isEmpty(other)) {
            return;
        }
        this.input0 = new Input(other.input0);
        this.input1 = new Input(other.input1);
        this.output0 = new Output(other.output0);
        this.output1 = new Output(other.output1);

        this.BlockNum = other.BlockNum;
        this.TxIdx = other.TxIdx;
        this.sig0 = other.sig0;
        this.sig1 = other.sig1;
        this.fee = toBN(other.fee);
        this.RootSig = other.RootSig;
    }

    toRpc() {
        return {
            BlockNum: this.BlockNum,
            TxIdx: this.TxIdx,
            sig0: this.sig0,
            sig1: this.sig1,
            fee: this.fee,
            RootSig: this.RootSig,
            input0: this.input0.toRpc(),
            input1: this.input1.toRpc(),
            output0: this.output0.toRpc(),
            output1: this.output1.toRpc()
        };
    }

    toString() {
        return JSON.stringify({
            BlockNum: this.BlockNum,
            TxIdx: this.TxIdx,
            sig0: ejs.bufferToHex(this.sig0),
            sig1: ejs.bufferToHex(this.sig1),
            fee: this.fee.toString(10),
            RootSig: ejs.bufferToHex(this.RootSig),
            input0: this.input0.toString(),
            input1: this.input1.toString(),
            output0: this.output0.toString(),
            output1: this.output1.toString()
        });
    }

    RLPEncode() {
        const a = [
                this.input0.blkNum, this.input0.txIdx, this.input0.outIdx, this.sig0,
                this.input1.blkNum, this.input1.txIdx, this.input1.outIdx, this.sig1,
                this.output0.newOwner, this.output0.amount,
                this.output1.newOwner, this.output1.amount,
                this.fee, this.RootSig
            ];
        return ejs.rlp.encode(a);
    }

    SignatureHash() {
        const input0Buf  = this.input0.Hash();
        const input1Buf  = this.input1.Hash();
        const output0Buf = this.output0.Hash();
        const output1Buf = this.output1.Hash();
        const feeBuf     = ejs.toBuffer(this.fee);
        const size = input0Buf.length + input1Buf.length + output0Buf.length + output1Buf.length + feeBuf.length;
        const buf = Buffer.concat([input0Buf, input1Buf, output0Buf, output1Buf, feeBuf], size);
        return ejs.sha256(buf);
    }
}

function toBN(input) {
    if (_.isEmpty(input)) {
        return web3.utils.toBN("0");
    }
    if (BN.isBN(input)) {
        return input;
    }
    if (_.has(input, 'values')) {
        const buffer = input.values;
        if (buffer.length == 0) {
            return web3.utils.toBN("0");
        }
        let chars = [];
        for (let i = 0; i < buffer.length; i++) {
            let v = buffer[i];
            chars = chars.concat(v.toString(16).padStart(2, 0));
        }
        const s = web3.utils.bytesToHex(buffer);
        return web3.utils.toBN(s);
    }
    return BN.BN(input);
}

module.exports = {
    Account: Account,
    PlasmaClient: PlasmaClient,
    Transaction: Transaction,
    toBN: toBN
};