pragma solidity ^0.4.17;

contract Plasma {
    event Deposit(address sender, uint value);

    function deposit() public payable {
        Deposit(msg.sender, msg.value);
    }
}