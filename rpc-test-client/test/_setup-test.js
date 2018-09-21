const _      = require('lodash');
const plasma = require('../index');
const b4     = require('./init');
const expect = require('chai').expect;


let accounts = b4.Accounts;
let contract = b4.Contract;
let web3     = b4.Web3;

describe('Setup', () => {

    it('gets the contract genesis block', async() => {
        let account = accounts[0];
        account.GetPlasmaBlock(1, (err, result) => {
            expect(err).to.equal(null);
        });
    });
    // This test is invoked from the ../bin/start script
    // Make sure you update that script if you change test's description
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
