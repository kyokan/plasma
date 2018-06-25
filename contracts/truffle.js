module.exports = {
  networks: {
    ropsten: {
      host: '45.55.29.108',
      port: 8545,
      network_id: 3,
      from: '0x0108d76118d97b88aa40167064cb242fa391effa',
      gas: 4000000
    },
    development: {
      host: '127.0.0.1',
      port: 9545,
      network_id: '*',
      gas: 4000000
    }
  }
};
