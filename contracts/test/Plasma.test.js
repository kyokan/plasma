const Plasma = artifacts.require('Plasma');

contract('Plasma', async (accounts) => {
  let deployed;

  beforeEach(async () => {
    deployed = await Plasma.deployed();
  });

  describe('#deposit', () => {
    it('should accept deposits', async () => {
      const depositTx = '0xf838808080808080808094627306090abab3a6e1400e9345bc60c78a8bef57830186a09400000000000000000000000000000000000000008080';

      await deployed.deposit(depositTx, {
        from: accounts[0],
        value: 100000
      });
    });
  });
});