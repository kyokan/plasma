const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;
const async = require('async');
const ejs = require('ethereumjs-util');

let accounts = b4.Accounts;
let web3     = b4.Web3;

describe('Send', () => {
    it(`sends 1 nano from ${accounts[1].address} to ${accounts[2].address}`, function (done) {
        let from = accounts[1];
        let to = accounts[2];
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'nano'));
        const sendWrapper = function(cb) {
            from.Send(to.address, amount, (err, transaction) => {
                // ignoring errors to ensure that all calls are made
                return cb(null, transaction);
            });
        };
        let fns = [];
        const concurrency = 1;
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

    it.skip(`sends 1 ether from ${accounts[1].address} to ${accounts[2].address}`, function (done) {
        const from = accounts[1];
        const to = accounts[2];
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'ether'));
        from.Send(to.address, amount, (err, transaction) => {
            expect(err).to.equal(null);
            expect(transaction.output0.amount.eq(amount)).to.equal(true);
            done();
        });
    });

});
