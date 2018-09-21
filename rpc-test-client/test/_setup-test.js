const W3 = require("web3");
const plasma = require('../index');
const _ = require('lodash');
const expect = require('chai').expect;

const plasmaJson = require('../../contracts/build/contracts/Plasma');
const contractItf = plasmaJson['abi'];
const web3 = new W3(`http://localhost:9545`);

// These should be the same across runs as ganache-cli is ran deterministic, bin/start
const accountAddress  = '0x627306090abab3a6e1400e9345bc60c78a8bef57';
const accountKey      = 'c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3';
const contractAddress = '0xf25186b5081ff5ce73482ad761db0eb0d25abfbf';
let contract;
let accounts;
let client;

before(async () => {
    contract = new web3.eth.Contract(contractItf, contractAddress);
    client = new plasma.PlasmaClient(__dirname + './../../rpc/proto/root.proto', 'localhost:8643');
});

describe('Setup', () => {
    it('checks contract address', async () => {
        accounts = _.mapValues(await web3.eth.getAccounts(), _.method('toLowerCase'));
        expect(accounts[0]).to.equal(accountAddress);
    });

    it('gets the contract genesis block', async() => {
        let account = new plasma.Account(client, web3, contract, accountAddress, accountKey);
        account.GetBlock(1, (err, result) => {
            expect(err).to.equal(null);
        });
    });

    it('should load client', (done) => {
        var client;
        try {
            client = new plasma.PlasmaClient(__dirname + './../../rpc/proto/root.proto', 'localhost:8643');
        } catch (err) {
            return done(err)
        }
        // Get balance for non-existent account (should return zero.)
        client.GetBalance(web3.utils.hexToBytes('0x0000000000000000000000000000000000000000'), function(err, result) {
            if (err !== null) {
                console.log('Error getting the balance');
                return done(err);
            }
            let balance = plasma.toBN(result.balance.values);
            let zero = web3.utils.toBN("0");
            if (!balance.eq(zero)) {
                return done(new Error(`Expected zero balance, got ${balance}`));
            }
            return done();
        });
    });


});
