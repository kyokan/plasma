const _  = require('lodash');
const b4 = require('./init');
const expect = require('chai').expect;
const plasma = require('../index');
const web3 = require('web3');
const ejs = require('ethereumjs-util');

let accounts = b4.Accounts;
let keyLookup = b4.Keys;

describe('Exits', () => {
    it(`exits from ${accounts[2].address}`, function (done) {
        let account = accounts[2];
        account.GetPlasmaUTXOs((err, utxos) => {
            expect(err).to.equal(null);
            expect(_.isEmpty(utxos)).to.equal(false);
            expect(_.isEmpty(utxos.confirmedTransactions)).to.equal(false);
            let utxo = utxos.confirmedTransactions[Math.floor(Math.random()*utxos.confirmedTransactions.length)];
            let transaction = new plasma.Transaction(utxo);

            let input0Address = web3.utils.bytesToHex(transaction.input0.owner);
            let input0Key = keyLookup[input0Address];
            let input1Address = web3.utils.bytesToHex(transaction.input1.owner);
            let input1Key = keyLookup[input1Address];

            account.GetPlasmaBlock(parseInt(transaction.blockNum.hex, 16), function (error, block) {
                expect(error).to.equal(null);
                const rlpTx = transaction.RLPEncode(true);
                const confirmationHash = ejs.sha256(Buffer.concat([ejs.sha256(rlpTx), block.block.header.merkleRoot]));

                let key0Buffer = Buffer.from(input0Key.secretKey);
                let signed0 = ejs.ecsign(confirmationHash, key0Buffer);
                let v0 = Buffer.from([signed0.v]);
                let confirmationSignatures = Buffer.concat([signed0.r, signed0.s, v0]);
                if (!_.isEmpty(input1Key)) {
                    let key1Buffer = Buffer.from(input1Key.secretKey);
                    let signed1 = ejs.ecsign(confirmationHash, key1Buffer);
                    let v1 = Buffer.from([signed1.v]);
                    confirmationSignatures = Buffer.concat([
                        confirmationSignatures,
                        signed1.r, signed1.s, v1
                    ]);
                }

                const commitedFee = '10000';
                console.log(`RLP Data is ${web3.utils.bytesToHex(rlpTx)}`);
                account.ExitTransaction(transaction, block, rlpTx, confirmationSignatures, commitedFee, (exitErr, exitResult) => {
                    expect(exitErr).to.equal(null);
                    console.log(exitResult);
                    done();
                });
            });
        });
    });

});