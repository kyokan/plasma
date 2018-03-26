pragma solidity ^0.4.17;

import './libraries/SafeMath.sol';

contract Plasma {
    using SafeMath for uint256;

    event Deposit(address sender, uint value);
    event SubmitBlock(address sender, bytes32 root);
    event ExitStarted(address sender, uint blocknum, uint txindex, uint oindex);

    address public authority;
    mapping(uint256 => childBlock) public childChain;
    mapping(uint256 => exit) public exits;
    uint256 public currentChildBlock;
    uint256[] public exitsIndexes;

    struct childBlock {
        bytes32 root;
        uint256 created_at;
    }

    struct exit {
        address owner;
        uint256 amount;
    }

    function Plasma() {
        authority = msg.sender;
        currentChildBlock = 1;
    }

    function startExit(
        uint256 blocknum,
        uint256 txindex,
        uint256 oindex,
        bytes txBytes,
        bytes proof
    ) public
    {
        // TODO: verify sender is owner of oindex.
        // TODO: verify sigs of owners of inputs signing off on transaction.
        // TODO: verify conf sigs of owners of inputs signing off on exit.
        // TODO: use oindex for prioritizing exits.
        // TODO: use priority queue later
        uint256 length = exitsIndexes.length++;
        exitsIndexes[length] = length;
        
        exits[length] = exit({
            owner: msg.sender,
            amount: msg.value //TODO: get from bytes
        });

        ExitStarted(msg.sender, blocknum, txindex, oindex);
    }

    function submitBlock(bytes32 root) public {
        require(msg.sender == authority);
        childChain[currentChildBlock] = childBlock({
            root: root,
            created_at: block.timestamp
        });
        currentChildBlock = currentChildBlock.add(1);

        SubmitBlock(msg.sender, root);
    }

    function deposit(bytes txBytes) public payable {
        // TODO: Make sure sender and amount matches data in txBytes.
        bytes32 root = createSimpleMerkleRoot(txBytes);

        childChain[currentChildBlock] = childBlock({
            root: root,
            created_at: block.timestamp
        });

        currentChildBlock = currentChildBlock.add(1);

        Deposit(msg.sender, msg.value);
    }

    function createSimpleMerkleRoot(bytes txBytes) returns (bytes32) {
        bytes32 zeroBytes;
        
        bytes32 root = keccak256(keccak256(txBytes), new bytes(130));
        for (uint i = 0; i < 16; i++) {
            root = keccak256(root, zeroBytes);
            zeroBytes = keccak256(zeroBytes, zeroBytes);
        }

        return root;
    }
}
