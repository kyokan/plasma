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
        this.client.getOutputs({address: address, spendable: true}, cb);
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
            const utxos = result.confirmedTransactions;
            let tx = self.transactionForAmount(to, amount, self.web3.utils.toBN(20000), utxos);
            const signedTransaction = self.signTransaction(tx);
            try {
                const rpcTx = signedTransaction.toRpc();
                self.client.Send({confirmed: rpcTx}, (sendErr, sendRes) => {
                    if (sendErr != null) {
                        return cb(sendErr);
                    }
                    return cb(null, new Transaction(sendRes.confirmed.transaction));
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
        result.sig0 = this.signInput(result.input0);
        result.sig1 = this.signInput(result.input1);
        const signatureHash = result.SignatureHash();
        const signed = ejs.ecsign(signatureHash, this.privateKey);
        const v = Buffer.from([signed.v]);
        const sig = Buffer.concat([signed.r, signed.s, v]);
        result.signatures.push(sig);
        result.signatures.push(sig);

        return result;
    }

    signInput(input) {
        // This will throw if key is invalid
        const signatureHash = input.SignatureHash();
        const signed = ejs.ecsign(signatureHash, this.privateKey);
        const v = Buffer.from([signed.v]);
        const sig = Buffer.concat([signed.r, signed.s, v]);
        return sig;
    }

    transactionForAmount(to, transfer, fee, utxos) {
        if (!_.isArray(utxos) || utxos.length == 0) {
            throw new Error('Invalid input');
        }
        let amount = transfer.add(fee);
        let result = new Transaction();
        // Construct first output
        result.output0 = new Output(to, transfer, toBN());
        let transactions = new Array(utxos.length);
        let sorter = [];
        for (let index = 0; index < utxos.length; index++) {
            const tx = new Transaction(utxos[index]);
            transactions[index] = tx;
            const addr0 = ejs.bufferToHex(tx.output0.newOwner);
            const addr1 = ejs.bufferToHex(tx.output1.newOwner);
            if (addr0 === this.address) {
                if (tx.output0.amount.cmp(amount) == 0) {
                    result.input0 = new Input(tx.blockNum, tx.txIdx, 0);
                    result.fee = fee;
                    return result;
                }
                sorter = sorter.concat([{amount: tx.output0.amount, index: index, outIdx: 0}]);
            }
            if (addr1 === this.address) {
                if (tx.output1.amount.cmp(amount) == 0) {
                    result.input0 = new Input(tx.blockNum, tx.txIdx, 1);
                    result.fee = fee;
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
            let outputAmount = inputAmount.sub(amount); // outputAmount is inputAmount - transfer - fee
            result.output1 = new Output(this.address, outputAmount, toBN());
            result.input0 = new Input(inputTx.blockNum, inputTx.txIdx, min.outIdx, this.address, inputAmount, toBN());
            result.fee = fee;
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
            result.input0 = new Input(inputTx0.BlockNum, inputTx0.TxIdx, this.address, sorter[leftBound].amount, toBN());
            result.input1 = new Input(inputTx1.BlockNum, inputTx1.TxIdx, this.address, sorter[rightBound].amount, toBN());

            return result;
        }
        if (lhs >= 0 && rhs >= 0) {
            const inputTx0 = transactions[sorter[lhs].index];
            const inputTx1 = transactions[sorter[lhs].index];
            result.input0 = new Input(inputTx0.BlockNum, inputTx0.TxIdx, sorter[lhs].outIdx, this.address, sorter[leftBound].amount, toBN());
            result.input1 = new Input(inputTx1.BlockNum, inputTx1.TxIdx, sorter[rhs].outIdx, this.address, sorter[leftBound].amount, toBN());
            let inputAmount = inputTx0.output0.amount;
            if (sorter[lhs].outIdx == 1) {
                inputAmount = inputTx0.output1.amount;
            }
            if (sorter[rhs].outIdx == 0) {
                inputAmount = inputAmount.add(inputTx1.output0.amount);
            } else {
                inputAmount = inputAmount.add(inputTx1.output1.amount);
            }
            result.output1 = new Output(this.address, inputAmount.sub(amount), toBN());

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

    ExitTransaction(transaction, block, txRlp, confirmationSignatures, committedFee, cb) {
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

        let leaves = [];
        for (let i = 0; i < block.confirmedTransactions.length; i++) {
            const elem = new Transaction(block.confirmedTransactions[i]);
            leaves.push(ejs.sha256(elem.RLPEncode()));
        }
        let merkle = generateMerkleRootAndProof(leaves, parseInt(transaction.txIdx.hex));

        const blkNum = self.web3.eth.abi.encodeParameter('uint256', parseInt(transaction.blockNum.hex, 16));
        const txIdx  = self.web3.eth.abi.encodeParameter('uint256', parseInt(transaction.txIdx.hex, 16));
        const outIdx = self.web3.eth.abi.encodeParameter('uint256', outputIdx);
        let pos = [blkNum,  txIdx, outIdx];

        const txBytes = self.web3.eth.abi.encodeParameter('bytes', txRlp);
        const proof = self.web3.eth.abi.encodeParameter('bytes', Buffer.from(merkle[1]));
        const encodedFee = self.web3.eth.abi.encodeParameter('uint256', committedFee);
        const value = web3.utils.toBN(web3.utils.toWei(committedFee, 'wei'));

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
                    from:     self.address,
                    value:    value
                };
                self.contract.methods.startTransactionExit(
                    pos,
                    txBytes,
                    proof,
                    self.web3.eth.abi.encodeParameter('bytes', confirmationSignatures),
                    encodedFee
                ).send(params, (error, receipt) => {
                    callback(error, receipt);
                });
            };
            async.retry({times: 3}, exitFn, (error, receipt) => {
                return cb(error, receipt);
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
        this.blockNum     = 0;
        this.txIdx        = 0;
        this.outIdx       = 0;
        this.owner        = Buffer.from(new Uint8Array(20));
        this.amount       = Buffer.from(new Uint8Array(0));
        this.depositNonce = Buffer.from(new Uint8Array(0));
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
            this.owner    = other.owner;
            this.amount   = other.amount;
            this.depositNonce = other.depositNonce;
            return;
        }
        if (args.length == 6) {
            this.blockNum     = args[0];
            this.txIdx        = args[1];
            this.outIdx       = args[2];
            this.owner        = args[3];
            this.amount       = toBN(args[4]);
            this.depositNonce = toBN(args[5]);
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
        const res = {
            blockNum: toRPC(this.blockNum),
            txIdx: toRPC(this.txIdx),
            outIdx: toRPC(this.outIdx),
            owner: ejs.toBuffer(this.owner),
            depositNonce: toRPC(this.depositNonce)
        };
        return res;
    }

    toArray() {
        const res = [
            toBuffer(this.blockNum),
            toBuffer(this.txIdx),
            toBuffer(this.outIdx),
            toBuffer(this.depositNonce),
            toBuffer(this.owner, 20),];
        return res;
    }

    RLPEncode() {
        const encoded = ejs.rlp.encode(this.toArray());
        return encoded;
    }

    SignatureHash() {
        const rlp = this.RLPEncode();

        const digest = doHash(rlp);
        //const hash = ejs.hashPersonalMessage(digest);
        return digest;
    }

    toString() {
        return this;
    }

}

class Output {
    zero() {
        this.newOwner     = Buffer.from(new Uint8Array(20));
        this.amount       = Buffer.from(new Uint8Array(0));
        this.depositNonce = Buffer.from(new Uint8Array(0));
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
            this.newOwner       = other.newOwner;
            this.amount         = toBN(other.amount);
            this.depositNonce   = toBN(other.depositNonce);

            return;
        }
        if (args.length == 3) {
            this.newOwner     = args[0];
            this.amount       = args[1];
            this.depositNonce = args[2];
            return;
        }
        throw new Error("Invalid number of arguments for constructing Output instance");
    }

    Hash() {
        const newOwnerBuf = ejs.toBuffer(this.newOwner);
        const amountBuf   = this.amount.toBuffer('be', this.amount.byteLength());
        const nonceBuf    = this.depositNonce.toBuffer('be', this.depositNonce.byteLength());
        const buf = Buffer.concat([newOwnerBuf, amountBuf, nonceBuf], newOwnerBuf.length + amountBuf.length + nonceBuf.length);
        const digest = doHash(buf);
        return digest;
    }

    toRpc() {
        return {
            newOwner: ejs.toBuffer(this.newOwner),
            amount: toRPC(this.amount),
            depositNonce: toRPC(this.depositNonce)
        };
    }

    toString() {
        return {
            newOwner: this.newOwner,
            amount: this.amount.toString(10),
            depositNonce: this.depositNonce.toString(10),
        };
    }
}


class Transaction {
    zero() {
        this.input0 = new Input();
        this.input1 = new Input();
        this.output0 = new Output();
        this.output1 = new Output();

        this.blockNum = "0";
        this.txIdx = 0;
        this.sig0 = Buffer.from("");
        this.sig1 = Buffer.from("");
        this.fee = Buffer.from("");
        this.signatures = [];
    }

    constructor(other) {
        this.zero();
        if (_.isEmpty(other)) {
            return;
        }
        let tx = other;
        if (!_.isEmpty(other.transaction )) {
            tx = other.transaction;
        }
        this.input0 = new Input(tx.input0);
        this.input1 = new Input(tx.input1);
        this.output0 = new Output(tx.output0);
        this.output1 = new Output(tx.output1);

        this.blockNum = tx.blockNum;
        this.txIdx = tx.txIdx;
        this.sig0 = tx.sig0;
        this.sig1 = tx.sig1;
        this.fee = toBN(tx.fee);
        this.signatures = other.signatures;
    }

    toRpc() {
        let res = {
            transaction: {
                blockNum: toRPC(this.blockNum),
                txIdx: toRPC(this.txIdx),
                sig0: this.sig0,
                sig1: this.sig1,
                fee: toRPC(this.fee),
                input0: this.input0.toRpc(),
                input1: this.input1.toRpc(),
                output0: this.output0.toRpc(),
                output1: this.output1.toRpc(),
            },
            signatures: this.signatures
        };
        return res;
    }

    toString() {
        return JSON.stringify({
            blockNum: this.blockNum,
            txIdx: this.txIdx,
            sig0: ejs.bufferToHex(this.sig0),
            sig1: ejs.bufferToHex(this.sig1),
            fee: this.fee.toString(10),
            input0: this.input0.toString(),
            input1: this.input1.toString(),
            output0: this.output0.toString(),
            output1: this.output1.toString()
        });
    }

    toArray() {
        return [
            toBuffer(this.input0.blockNum),
            toBuffer(this.input0.txIdx),
            toBuffer(this.input0.outIdx),
            toBuffer(this.input0.depositNonce),
            toBuffer(this.input0.owner, 20),
            toBuffer(this.sig0, 65),

            toBuffer(this.input1.blockNum),
            toBuffer(this.input1.txIdx),
            toBuffer(this.input1.outIdx),
            toBuffer(this.input1.depositNonce),
            toBuffer(this.input1.owner, 20),
            toBuffer(this.sig1, 65),

            toBuffer(this.output0.newOwner, 20),
            toBuffer(this.output0.amount),
            toBuffer(this.output1.newOwner, 20),
            toBuffer(this.output1.amount),
            toBuffer(this.fee),
        ];
    }

    RLPEncode(isExit) {
        let rlpInput = [];
        if (isExit === true) {
            rlpInput = [
                this.toArray(),
                this.signatures
            ];
        }
        else {
            rlpInput = this.toArray();
        }

        const encoded = ejs.rlp.encode(rlpInput);
        return encoded;
    }

    SignatureHash() {
        const rlp = this.RLPEncode();

        const digest = doHash(rlp);
        //const hash = ejs.hashPersonalMessage(digest);
        return digest;
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

function toBuffer(input, bufferLength) {
    let i = input;
    if (_.has(i, 'hex')) {
        i = input.hex;
    }
    let length = 32;
    if (_.isNumber(bufferLength)) {
        length = bufferLength;
    }
    let res = null;
    if (_.has(i, 'toBuffer')) {
        res = i.toBuffer('be', length);
    } else {
        const buf = ejs.toBuffer(i);
        const diff = length - buf.length;
        res = Buffer.concat([Buffer.from(new Uint8Array(diff)), buf]);
    }
    return res;

}

function toRPC(input) {
    if (_.has(input, 'hex')) {
        return input;
    }
    return { hex: '0x' + input.toString(16)}
}

let tendermintSHA256 = function(input1, input2) {
    const buf = Buffer.concat([
        Buffer.from([input1.length]),
        input1,
        Buffer.from([input2.length]),
        input2,
    ]);
    return ejs.sha256(buf);
};

// Simple Tree: https://tendermint.com/docs/spec/blockchain/encoding.html#merkle-trees
let generateMerkleRootAndProof = function(leaves, index) {
    if (leaves.length == 0) { // If there are no leaves, then we can't generate anything
        return [Buffer.from(""), Buffer.from("")];
    } else if (leaves.length == 1) { // If there's only 1 leaf, return it with and empty proof
        return [leaves[0], Buffer.from("")];
    } else {
        let pivot = Math.floor((leaves.length + 1) / 2);

        let left, right;
        let proof = Buffer.from("");

        // If the index will be in the left subtree (index < pivot), then we
        // need to generate the proof using the intermediary hash from the right
        // side. Otherwise, do the reverse.
        if (index < pivot) {
            // recursively call the function on the leaves that will be in the
            // left and right sub trees.
            left = generateMerkleRootAndProof(leaves.slice(0, pivot), index);
            right = generateMerkleRootAndProof(leaves.slice(pivot, leaves.length), -1);

            // add current level's right intermediary hash to the proof
            if (index >= 0) {
                proof = Buffer.concat([left[1], right[0]]);
            }
        } else {
            // recursively call the function on the leaves that will be in the
            // left and right sub trees.
            // since the index will be in the right sub tree, we need to update
            // it's value.
            left = generateMerkleRootAndProof(leaves.slice(0, pivot), -1);
            right = generateMerkleRootAndProof(leaves.slice(pivot, leaves.length), index - pivot);

            // add current level's left intermediary hash to the proof
            if (index >= 0) {
                proof = Buffer.concat([right[1], left[0]]);
            }
        }
        return [tendermintSHA256(left[0], right[0]), toHex(proof)];
    }
};


module.exports = {
    Account: Account,
    PlasmaClient: PlasmaClient,
    Transaction: Transaction,
    toBN: toBN
};