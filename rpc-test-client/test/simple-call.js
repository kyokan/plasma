
const plasma = require('../index');
const _ = require('lodash');
const BigNumber = require('bignumber.js');


it('should load client', (done) => {
    var client;
    try {
        client = new plasma.PlasmaClient(__dirname + './../../rpc/proto/root.proto', 'localhost:8643');
    } catch (err) {
        return done(err)
    }
    // Get balance for non-existent account (should return zero.)
    client.GetBalance('0x0000000000000000000000000000000000000000', function(err, result) {
        if (err !== null) {
            console.log('Error getting the balance');
            return done(err);
        }
        let chars = "";
        let buffer = result.balance.values;
        for (let i = 0; i < buffer.length; i++) {
            let v = buffer[i];
            chars = chars.concat(v.toString(16).padStart(2, 0));
        }
        if (_.isEmpty(chars)) {
            chars = chars.concat("0");
        }
        let balance = new BigNumber(chars, 16);
        let zero = new BigNumber("0");
        if (!balance.eq(zero)) {
            return done(new Error(`Expected zero balance, got ${balance}`));
        }
        return done();
    });
});