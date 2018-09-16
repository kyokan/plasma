
let protoLoader = require('@grpc/proto-loader');
let grpc = require('grpc');
let _ = require('lodash');

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

module.exports = {
    PlasmaClient: PlasmaClient
}