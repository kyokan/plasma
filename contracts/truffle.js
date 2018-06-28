module.exports = {
  networks: {
    ropsten: {
      host: '45.55.29.108',
      port: 8545,
      network_id: 3,
      from: '0x0108d76118d97b88aa40167064cb242fa391effa',
      gas: 4000000
    },
    rinkeby: {
        host: '127.0.0.1',
        port: 8545,
        network_id: 4,
        from: '0xe8a001b12568de55223021bf31a9010b925ef012',
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
