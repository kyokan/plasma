const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;
const plasma = require('../index');

let accounts = b4.Accounts;

describe('Exits', () => {
    it(`exits from ${accounts[2].address}`, function (done) {
        let account = accounts[2];
        account.GetPlasmaUTXOs((err, utxos) => {
            expect(err).to.equal(null);
            expect(_.isEmpty(utxos)).to.equal(false);
            expect(_.isEmpty(utxos.transactions)).to.equal(false);
            let utxo = utxos.transactions[Math.floor(Math.random()*utxos.transactions.length)];
            account.Exit(new plasma.Transaction(utxo), (exitErr, exitResult) => {
                expect(exitErr).to.equal(null);
                console.log(exitResult);
                done();
            });
        });
    });

});