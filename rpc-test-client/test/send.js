const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;
const async = require('async');
const ejs = require('ethereumjs-util');

let accounts = b4.Accounts;
let web3     = b4.Web3;
const repeat = 1;

describe('Send', () => {
    it(`sends 1 nano from ${accounts[1].address} to ${accounts[2].address} ${repeat} times`, function (done) {
        let from = accounts[1];
        let to = accounts[2];
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'nano'));
        const sendWrapper = function(id, next) {
            from.Send(to.address, amount, (err, transaction) => {
                if (err != null) {
                    console.log(`Error sending transaction: ${err}`);
                }
                // ignoring errors to ensure that all calls are made
                return next(null, transaction);
            });
        };
        async.timesLimit(repeat, 1, sendWrapper, function (err, results) {
            expect(results.length).to.equal(repeat);
            for (let i = 0; i < results.length; i++) {
                expect(results[i].transaction.output0.amount.eq(amount)).to.equal(true);
            }
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
