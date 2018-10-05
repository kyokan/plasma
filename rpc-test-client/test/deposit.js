const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;

let accounts = b4.Accounts;
let web3     = b4.Web3;

describe('Deposit (long test)', () => {
    it('deposits funds from first account and checks balance', function (done) {
        let account = accounts[0];
        const amount = web3.utils.toBN(web3.utils.toWei('3', 'ether'));
        account.GetPlasmaBalance((err, initialBalance) => {
            expect(err).to.equal(null);
            if (err != null) {
                return done(err);
            }
            account.Deposit(amount, function (err) {
                expect(err).to.equal(null);
                if (err != null) {
                    return done(err);
                }

                let timeout;
                let delayedBalanceFn = function() {
                    if (timeout != null && _.has(timeout, 'unref')) {
                        timeout.unref();
                    }
                    account.GetPlasmaBalance((err, balance) => {
                        expect(err).to.equal(null);
                        if (err != null) {
                            return done(err);
                        }
                        let expected = initialBalance.add(amount);
                        expect(balance.eq(expected)).to.equal(true);
                        done();
                    });
                };
                timeout = setTimeout(delayedBalanceFn, 10000);
            });
        });

    }).timeout(11000);

});
