const W3 = require("web3");
const _ = require('lodash');
const expect = require('chai').expect;
const plasma = require('../index');

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

describe('Send', () => {
    it('sends 1 nano from 0x627306090abab3a6e1400e9345bc60c78a8bef57 to 0xf17f52151ebef6c7334fad080c5704d77216b732', function () {
        let account = new plasma.Account(client, web3, contract, accountAddress, accountKey);
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'nano'));
        account.Send('0xf17f52151ebef6c7334fad080c5704d77216b732', amount, (err, transaction) => {
            expect(err).to.equal(null);
            expect(transaction.output0.amount.eq(amount)).to.equal(true);
        });
    });

    it('sends 1 ether from 0x627306090abab3a6e1400e9345bc60c78a8bef57 to 0xf17f52151ebef6c7334fad080c5704d77216b732', function () {
        let account = new plasma.Account(client, web3, contract, accountAddress, accountKey);
        const amount = web3.utils.toBN(web3.utils.toWei('1', 'ether'));
        account.Send('0xf17f52151ebef6c7334fad080c5704d77216b732', amount, (err, transaction) => {
            expect(err).to.equal(null);
            expect(transaction.output0.amount.eq(amount)).to.equal(true);
        });
    });

});
