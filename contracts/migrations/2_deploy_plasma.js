var Plasma = artifacts.require("./Plasma.sol");
var PriorityQueue = artifacts.require("./PriorityQueue.sol");

module.exports = function(deployer) {
  deployer.deploy(PriorityQueue);
  // I think Plasma will reference its own Priority Queue impl and that's okay.
  deployer.deploy(Plasma);
};
