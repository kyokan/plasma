var CBOR = artifacts.require("./libraries/CBOR.sol");
var SafeMath = artifacts.require("./libraries/SafeMath.sol");

module.exports = function(deployer) {
  // TODO: Might need to link this to the Plasma contract.
  deployer.deploy(CBOR);
  deployer.deploy(SafeMath);
};
