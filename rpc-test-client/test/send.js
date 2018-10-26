const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;
const async = require('async');

let accounts = b4.Accounts;
let web3     = b4.Web3;

describe('Send', () => {
    it(`sends 1 nano from ${accounts[0].address} to ${accounts[1].address}`, function (done) {
        let account = accounts[0];
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'nano'));
        const sendWrapper = function(cb) {
            account.Send('0xf17f52151ebef6c7334fad080c5704d77216b732', amount, (err, transaction) => {
                // ignoring errors to ensure that all calls are made
                return cb(null, transaction);
            });
        };
        let fns = [];
        const concurrency = 64;
        for (let i = 0; i < concurrency; i++) {
            fns.push(sendWrapper)
        }
        async.parallel(fns, function (err, results) {
            let count = 0;
            let transaction;
            for (let i = 0; i < results.length; i++) {
                if (!_.isEmpty(results[i])) {
                    count++;
                    transaction = results[i];
                }
            }
            expect(count).to.equal(1);
            expect(transaction.output0.amount.eq(amount)).to.equal(true);
            done();
        });
    });

    it.skip(`sends 1 ether from ${accounts[0].address} to ${accounts[1].address}`, function (done) {
        let account = accounts[0];
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'ether'));
        account.Send('0xf17f52151ebef6c7334fad080c5704d77216b732', amount, (err, transaction) => {
            expect(err).to.equal(null);
            expect(transaction.output0.amount.eq(amount)).to.equal(true);
            done();
        });
    });

});
