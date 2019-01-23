const W3 = require("web3");
const _ = require('lodash');
const plasma = require('../index');

const plasmaJson = require('../../plasma-mvp-rootchain/build/contracts/PlasmaMVP');
const contractItf = plasmaJson['abi'];
const web3 = new W3(`http://localhost:8545`);

// These should be the same across runs as ganache-cli is ran deterministic, bin/start
const ganacheAccountsInfo = require('../../test/test-accounts');
let accounts = [];

const contractAddress = '0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da';
let contract;
let client;

let state = {};

function init() {
    if (!_.isEmpty(state)) {
        return state;
    }
    contract = new web3.eth.Contract(contractItf, contractAddress);
    client = new plasma.PlasmaClient(__dirname + './../../rpc/proto/root.proto', 'localhost:6545');
    _.each(ganacheAccountsInfo.private_keys, function(key, address) {
        const info = ganacheAccountsInfo.addresses[address];
        accounts = accounts.concat(new plasma.Account(client, web3, contract, address, info.secretKey, info.publicKey));
    });

    state = {
        Contract: contract,
        Accounts: accounts,
        Client: client,
        Keys: ganacheAccountsInfo.addresses, // used to lookup the keys for signing exits
        Web3: web3
    };
    return state;
}

module.exports = init();
