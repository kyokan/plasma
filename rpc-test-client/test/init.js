const W3 = require("web3");
const _ = require('lodash');
const plasma = require('../index');

const plasmaJson = require('../../contracts/build/contracts/Plasma');
const contractItf = plasmaJson['abi'];
const web3 = new W3(`http://localhost:9545`);

// These should be the same across runs as ganache-cli is ran deterministic, bin/start
const ganacheAccountsInfo = require('../../test/test-accounts');
let accounts = [];

const contractAddress = '0xf25186b5081ff5ce73482ad761db0eb0d25abfbf';
let contract;
let client;

let state = {};

function init() {
    if (!_.isEmpty(state)) {
        return state;
    }
    contract = new web3.eth.Contract(contractItf, contractAddress);
    client = new plasma.PlasmaClient(__dirname + './../../rpc/proto/root.proto', 'localhost:8643');
    _.each(ganacheAccountsInfo.private_keys, function(key, address) {
        accounts = accounts.concat(new plasma.Account(client, web3, contract, address, key));
    });

    state = {
        Contract: contract,
        Accounts: accounts,
        Client: client,
        Web3: web3
    };
    return state;
}

module.exports = init();
