const async = require('async');
const protoLoader = require('@grpc/proto-loader');
const grpc = require('grpc');
const _ = require('lodash');
const ejs = require('ethereumjs-util');
const web3 = require('web3');
const BN = require('bn.js');
const secp256k1 = require('secp256k1');
const MerkleTree = require('merkletreejs');
const sha3 = require('js-sha3');
const UInt64BE = require('int64-buffer').Uint64BE;
const ethLib = require('eth-lib');

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
        if (l != 5) {
            throw new Error(`Got ${l} definitions, expecting 4`);
        }

        let GetBalanceFtor = definitions.GetBalance;
        let GetOutputsFtor = definitions.GetOutputs;
        let GetBlockFtor   = definitions.GetBlock;
        let SendFtor       = definitions.Send;
        let BlockHeightFtor= definitions.BlockHeight;

        if (_.isEmpty(GetBalanceFtor)) {
            throw new Error('GetBalance definition is missing');
        }
        if (_.isEmpty(GetOutputsFtor)) {
            throw new Error('GetOutputs definition is missing');
        }
        if (_.isEmpty(GetBlockFtor)) {
            throw new Error('GetBlock definition is missing');
        }
        if (_.isEmpty(SendFtor)) {
            throw new Error('Send definition is missing');
        }
        if (_.isEmpty(BlockHeightFtor)) {
            throw new Error('BlockHeight definition is missing');
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
    constructor(client, web3, contract, address, key, publicKey) {
        this.client = client;
        this.web3 = web3;
        this.contract = contract;
        this.address = address;
        this.privateKey = Buffer.from(key);
        this.publicKey = Buffer.from(publicKey);
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
                gasEstimate: 4 * results[0],
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
            const signedTransaction = self.signTransaction(tx);
            try {
                const rpcTx = signedTransaction.toRpc();
                self.client.Send({transaction: rpcTx}, (sendErr, sendRes) => {
                    if (sendErr != null) {
                        return cb(sendErr);
                    }
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
        const signatureHash = result.SignatureHash();
        const signed = ejs.ecsign(signatureHash, this.privateKey);
        const v = Buffer.from([signed.v]);
        const sig = Buffer.concat([signed.r, signed.s, v]);
        result.sig0 = sig;
        result.sig1 = sig;
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
                    return result;
                }
                sorter = sorter.concat([{amount: tx.output0.amount, index: index, outIdx: 0}]);
            }
            if (addr1 === this.address) {
                if (tx.output1.amount.cmp(amount) == 0) {
                    result.input0 = new Input(tx.BlockNum, tx.TxIdx, 1);
                    return result;
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
            return result;
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

            return result;
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

            return result;
        }

        throw new Error('no suitable UTXOs found');
    }

    Deposit(amount, cb) {
        let self = this;
        let depositFn = function (callback) {
            self.prepareEthCall( (err, result) => {
                if (err) {
                    return callback(e);
                }
                let params = {
                    nonce:    self.web3.utils.toHex(result.nonce),
                    chainId:  15,
                    to:       self.contract.options.address,
                    value:    amount,
                    gasPrice: self.web3.utils.toHex(result.gasPrice),
                    gas:      self.web3.utils.toHex(result.gasEstimate), // just to be safe
                    from:     self.address
                };
                self.contract.methods.deposit(self.address).send(params, (error, receipt) => {
                    callback(error, receipt);
                });
            });
        };
        async.retry({times: 5}, depositFn, (error, receipt) => {
            return cb(error, receipt);
        });
    }

    Exit(transaction, cb) {
        let self = this;
        const output0Address = self.web3.utils.bytesToHex(transaction.output0.newOwner);
        const output1Address = self.web3.utils.bytesToHex(transaction.output1.newOwner);
        let outputIdx;
        if (self.address === output0Address) {
            outputIdx = 0;
        }
        if (self.address === output1Address) {
            outputIdx = 1;
        }
        self.GetPlasmaBlock(transaction.BlockNum, (err, block) => {
            if (err != null) {
                return cb(err);
            }

            let elements = new Array(Math.pow(2, 16));
            const zeroTx = doHash(Buffer.from(new Uint8Array(32)));
            for (let i = 0; i < Math.pow(2, 16); i++) {
                if (i < block.transactions.length) {
                    const tx = new Transaction(block.transactions[i]);
                    elements[i] = doHash(tx.RLPEncode());
                }
                else {
                    elements[i] = zeroTx;
                }
            }
            const merkleTree = new MerkleTree(elements, doHash);
            const root = merkleTree.getRoot();
            let proof = merkleTree.getProof(elements[transaction.TxIdx]);
            let proofLength = 0;
            let proofBuffer = Buffer.from('');
            for (let i = 0; i < proof.length; i++) {
                proofLength += proof[i].data.length;
                proofBuffer = Buffer.concat([proofBuffer, proof[i].data], proofLength);
            }

            self.prepareEthCall( (err, result) => {
                if (err != null) {
                    return cb(err, null);
                }

                let nonce = result.nonce;
                // Retrying as sometimes the call fails with invalid RPC response error
                let exitFn = function (callback) {
                    let params = {
                        nonce:    self.web3.utils.toHex(nonce),
                        chainId:  15,
                        to:       self.contract.options.address,
                        gasPrice: self.web3.utils.toHex(result.gasPrice),
                        gas:      self.web3.utils.toHex(5 * result.gasEstimate),
                        from:     self.address
                    };
                    self.contract.methods.startExit(
                        transaction.BlockNum, // uint64
                        transaction.TxIdx, // uint32
                        outputIdx, // uint8
                        transaction.RLPEncode(),
                        proofBuffer
                    ).send(params, (error, receipt) => {
                        callback(error, receipt);
                    });
                };
                async.retry({times: 3}, exitFn, (error, receipt) => {
                    return cb(error, receipt);
                });
            });
        });
    }

    ExitDeposit(depositNonce, committedFee, cb) {
        let self = this;
        let exitDepositFn = function (callback) {
            self.prepareEthCall( (err, result) => {
                if (err != null) {
                    return cb(err);
                }

                let nonce = result.nonce;
                const encodedNonce = self.web3.eth.abi.encodeParameter('uint256', depositNonce);
                const encodedFee = self.web3.eth.abi.encodeParameter('uint256', committedFee);
                const value = web3.utils.toBN(web3.utils.toWei(committedFee, 'wei'));

                let params = {
                    nonce: self.web3.utils.toHex(nonce),
                    chainId: 15,
                    to: self.contract.options.address,
                    value: value,
                    gasPrice: self.web3.utils.toHex(result.gasPrice),
                    gas: self.web3.utils.toHex(5 * result.gasEstimate),
                    from: self.address
                };
                self.contract.methods.startDepositExit(encodedNonce, encodedFee)
                    .send(params, (error, exitResult) => {
                        callback(error, exitResult);
                    });
            });
        };
        async.retry({times: 5}, exitDepositFn, (error, exitResult) => {
            return cb(error, exitResult);
        });
    }
}

function doHash(input) {
    const result = web3.utils.sha3(input);
    return Buffer.from(web3.utils.hexToBytes(result));
}

class Input {
    zero() {
        this.blockNum = 0;
        this.txIdx    = 0;
        this.outIdx   = 0;
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
            if (_.isNumber(other.blockNum)) {
                this.blockNum = other.blockNum;
            } else {
                this.blockNum = parseInt(other.blockNum, 16);
            }
            this.txIdx    = other.txIdx;
            this.outIdx   = other.outIdx;
            return;
        }
        if (args.length == 3) {
            if (_.isNumber(args[0])) {
                this.blockNum = args[0];
            }
            else {
                this.blockNum = parseInt(args[0], 16);
            }

            this.txIdx    = args[1];
            this.outIdx   = args[2];
            return;
        }
        throw new Error("Invalid number of arguments for constructing Input instance");
    }

    Hash() {
        const blockNumBuf = new UInt64BE(this.blockNum).toBuffer(); // 64 bits
        const txIdxStr    = this.txIdx.toString(16).padStart(8, 0); // 16 bits
        const txIdxBuf    = Buffer.from(txIdxStr, 'hex');
        const outIdxBuf   = ejs.toBuffer(this.outIdx);
        const buf = Buffer.concat([blockNumBuf, txIdxBuf, outIdxBuf], blockNumBuf.length + txIdxBuf.length + outIdxBuf.length);
        const digest = doHash(buf);
        return digest;
    }

    toRpc() {
        return {
            blockNum: this.blockNum.toString(),
            txIdx: this.txIdx,
            outIdx: this.outIdx
        };
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
        const amountBuf   = this.amount.toBuffer('be', this.amount.byteLength());
        const buf = Buffer.concat([newOwnerBuf, amountBuf], newOwnerBuf.length + amountBuf.length);
        const digest = doHash(buf);
        return digest;
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

    toArray() {
        return [
            this.input0.blockNum,
            this.input0.txIdx,
            this.input0.outIdx,
            this.sig0,
            this.input1.blockNum,
            this.input1.txIdx,
            this.input1.outIdx,
            this.sig1,
            this.output0.newOwner,
            this.output0.amount,
            this.output1.newOwner,
            this.output1.amount,
            0, //this.fee,
            this.RootSig
        ];
    }

    RLPEncode() {
        const encoded = ejs.rlp.encode(this.toArray());
        return encoded;
    }

    SignatureHash() {
        const input0Buf  = this.input0.Hash();
        const input1Buf  = this.input1.Hash();
        const output0Buf = this.output0.Hash();
        const output1Buf = this.output1.Hash();
        let feeBuf = Buffer.from('');
        if (this.fee.byteLength() > 0) {
            feeBuf = this.fee.toBuffer('be', this.fee.byteLength());
        }

        let size = input0Buf.length + input1Buf.length + output0Buf.length + output1Buf.length + feeBuf.length;
        const buf = Buffer.concat([input0Buf, input1Buf, output0Buf, output1Buf, feeBuf], size);
        const digest = doHash(buf);
        const hash = ejs.hashPersonalMessage(digest);
        return hash;
    }
}

function toBN(input) {
    if (_.isEmpty(input)) {
        return web3.utils.toBN("0");
    }
    if (BN.isBN(input)) {
        return input;
    }
    if (_.has(input, 'hex')) {
        return web3.utils.toBN(input.hex);
    }
    return BN.BN(input);
}

module.exports = {
    Account: Account,
    PlasmaClient: PlasmaClient,
    Transaction: Transaction,
    toBN: toBN
};