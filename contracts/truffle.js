const PrivateKeyProvider = require('truffle-privatekey-provider');

module.exports = {
  networks: {
    ropsten: {
      provider: new PrivateKeyProvider(process.env.PRIVATE_KEY, process.env.HOST),
      network_id: 3,
      gas: 6000000,
      gasPrice: '5000000000'
    },
    rinkeby: {
      provider: new PrivateKeyProvider(process.env.PRIVATE_KEY, process.env.HOST),
      network_id: 4,
      gas: 6000000,
      gasPrice: '5000000000'
    },
    local: {
      host: '127.0.0.1',
      port: 8545,
      network_id: 15,
      gas: 4000000
    },
    development: {
      host: '127.0.0.1',
      port: 9545,
      network_id: '*',
      gas: 6700000
    }
  }
};
