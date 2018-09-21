const async = require('async');
const protoLoader = require('@grpc/proto-loader');
const grpc = require('grpc');
const _ = require('lodash');
const ejs = require('ethereumjs-util');
const web3 = require('web3');

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
        this.client.getUTXOs({address: address}, cb);
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

    }

    GetPlasmaBalance(cb) {
        this.client.GetBalance(this.web3.utils.hexToBytes(this.address), (err, result) => {
            if (err !== null) {
                return cb(err);
            }
            const balance = toBN(result.balance.values);
            return cb(null, balance);
        });
    }

    GetBlock(number, cb) {
        this.contract.methods.getBlock(number).call(cb);
    }

    Deposit(amount, cb) {
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
                return cb(err, null);
            }

            let tx = new Transaction(
                null, // input0
                null, // input1
                {NewOwner: self.address, Amount: amount}, //output0
                null, // output1
                0); // fee
            const rlpEncoded = tx.RLPEncode();
            let nonce = results[2];
            // Retrying as sometimes the call fails with invalid RPC response error
            let depositFn = function (callback) {
                let params = {
                    nonce:    self.web3.utils.toHex(nonce),
                    chainId:  15,
                    to:       self.contract.options.address,
                    value:    amount,
                    gasPrice: self.web3.utils.toHex(results[1]),
                    gas:      self.web3.utils.toHex(2 * results[0]), // just to be safe
                    from:     self.address
                };
                self.contract.methods.deposit(rlpEncoded).send(params, (error, receipt) => {
                    if (error != null) {
                        if (error.message !== 'Invalid JSON RPC response') {
                            nonce++;
                        }
                    }
                    callback(error, receipt);
                });
            };
            async.retry({times: 3}, depositFn, (error, receipt) => {
                return cb(error, receipt);
            });
        });
    }
}

class Transaction {
    constructor(input0, input1, output0, output1, fee) {
        if (!_.isEmpty(input0)) {
            this.BlkNum0 = input0.BlkNum;
            this.TxIdx0  = input0.TxIdx;
            this.OutIdx0 = input0.OutIdx;
            this.Sig0    = input0.Sig;
        } else {
            this.BlkNum0 = '';
            this.TxIdx0  = '';
            this.OutIdx0 = '';
            this.Sig0    = '';
        }
        if (!_.isEmpty(input1)) {
            this.BlkNum1 = input1.BlkNum;
            this.TxIdx1  = input1.TxIdx;
            this.OutIdx1 = input1.OutIdx;
            this.Sig1    = input1.Sig;
        } else {
            this.BlkNum1 = '';
            this.TxIdx1  = '';
            this.OutIdx1 = '';
            this.Sig1    = '';
        }
        if (!_.isEmpty(output0)) {
            this.NewOwner0 = output0.NewOwner;
            this.Amount0   = output0.Amount;
        } else {
            this.NewOwner0 = Buffer.from(new Uint8Array(20));
            this.Amount0   = '';
        }
        if (!_.isEmpty(output1)) {
            this.NewOwner1 = output1.NewOwner;
            this.Amount1   = output1.Amount;
        }else {
            this.NewOwner1 = Buffer.from(new Uint8Array(20));
            this.Amount1   = '';
        }
        this.Fee = '';
    }

    RLPEncode() {
        const a = [
                this.BlkNum0, this.TxIdx0, this.OutIdx0, this.Sig0,
                this.BlkNum1, this.TxIdx1, this.OutIdx1, this.Sig1,
                this.NewOwner0, this.Amount0,
                this.NewOwner1, this.Amount1,
                this.Fee
            ];
        return ejs.rlp.encode(a);
    }
}

function toBN(buffer) {
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

module.exports = {
    Account: Account,
    PlasmaClient: PlasmaClient,
    Transaction: Transaction,
    toBN: toBN
};