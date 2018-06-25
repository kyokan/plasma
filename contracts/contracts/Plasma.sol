pragma solidity 0.4.24;

import './libraries/ByteUtils.sol';
import './PriorityQueue.sol';
import './libraries/RLP.sol';
import './libraries/SafeMath.sol';

contract Plasma {
    using SafeMath for uint256;
    using RLP for bytes;
    using RLP for RLP.RLPItem;
    using RLP for RLP.Iterator;

    event Deposit(address sender, uint value);
    event SubmitBlock(address sender, bytes32 root);
    event ExitStarted(address sender, uint exitId);
    event ChallengeSuccess(address sender, uint exitId);
    event ChallengeFailure(address sender, uint exitId);
    event FinalizeExit(address sender, uint exitId);
    event DebugBytes32(address sender, bytes32 item);
    event DebugBytes(address sender, bytes item);
    event DebugAddress(address sender, address item);
    event DebugUint(address sender, uint item);
    event DebugBool(address sender, bool item);

    address public authority;
    mapping(uint256 => ChildBlock) public childChain;
    mapping(uint256 => Exit) public exits;
    uint256 public currentChildBlock;
    PriorityQueue public exitQueue;
    uint256 public lastExitId;
    uint256 public lastFinalizedTime;

    struct ChildBlock {
        bytes32 root;
        uint256 created_at;
    }

    struct Exit {
        address owner;
        uint256 amount;
        uint256 blocknum;
        uint256 txindex;
        uint256 oindex;
        uint256 started_at;
    }

    constructor() {
        authority = msg.sender;
        currentChildBlock = 1;
        lastFinalizedTime = block.timestamp;
        exitQueue = new PriorityQueue();
    }

    function submitBlock(bytes32 root) public {
        require(msg.sender == authority);
        childChain[currentChildBlock] = ChildBlock({
            root: root,
            created_at: block.timestamp
        });
        currentChildBlock = currentChildBlock.add(1);

        SubmitBlock(msg.sender, root);
    }

    function getBlock(uint256 blocknum)
        public
        view
        returns (bytes32, uint256)
    {
        ChildBlock memory blk = childChain[blocknum];
        return (blk.root, blk.created_at);
    }

    function deposit(bytes txBytes) public payable {
        RLP.RLPItem memory txItem = txBytes.toRLPItem();
        RLP.RLPItem[] memory txList = txItem.toList();

        uint newOwnerIdx = 6;
        uint amountIdx = 7;
        require(msg.sender == txList[newOwnerIdx].toAddress());
        require(msg.value == txList[amountIdx].toUint());

        bytes32 root = createSimpleMerkleRoot(txBytes);

        childChain[currentChildBlock] = ChildBlock({
            root: root,
            created_at: block.timestamp
        });

        currentChildBlock = currentChildBlock.add(1);

        Deposit(msg.sender, msg.value);
    }

    function createSimpleMerkleRoot(bytes txBytes) returns (bytes32) {
        // TODO: We may want a different null value.
        bytes32 zeroHash = keccak256(hex"0000000000000000000000000000000000000000000000000000000000000000");
        bytes32 root = keccak256(txBytes);
        
        for (uint i = 0; i < 15; i++) {
            root = keccak256(root, zeroHash);
        }

        return root;
    }

    function startExit(
        uint256 blocknum,
        uint256 txindex,
        uint256 oindex,
        bytes txBytes,
        bytes proof
    ) public
    {
        RLP.RLPItem memory txItem = txBytes.toRLPItem();
        RLP.RLPItem[] memory txList = txItem.toList();

        uint baseIndex = 6 + (oindex * 2);

        require(msg.sender == txList[baseIndex].toAddress());

        uint amount = txList[baseIndex + 1].toUint();
        // Simplify contract by only allowing exits > 0
        require(amount > 0);

        bool exists = checkProof(blocknum, txindex, txBytes, proof);

        require(exists);

        // TODO: check that the sigs given to the utxo owner from the input owner
        // are legit from the side chain.

        uint256 priority = calcPriority(blocknum, txindex, oindex);
        lastExitId = priority; // For convenience and debugging.
        exitQueue.add(priority);
        
        exits[priority] = Exit({
            owner: msg.sender,
            amount: amount,
            // These are necessary for challenges.
            blocknum: blocknum,
            txindex: txindex,
            oindex: oindex,
            started_at: block.timestamp
        });

        ExitStarted(msg.sender, priority);
    }

    function getExit(uint256 exitId)
        public
        view
        returns (address, uint256, uint256, uint256, uint256, uint256)
    {
        Exit memory exit = exits[exitId];

        return (exit.owner, exit.amount, exit.blocknum, exit.txindex, exit.oindex, exit.started_at);
    }

    function challengeExit(
        uint256 exitId,
        uint256 blocknum,
        uint256 txindex,
        bytes txBytes,
        bytes proof
    ) public
    {
        Exit memory currExit = exits[exitId];
        RLP.RLPItem memory txItem = txBytes.toRLPItem();
        RLP.RLPItem[] memory txList = txItem.toList();

        bool firstInput = txList[0].toUint() == currExit.blocknum && txList[1].toUint() == currExit.txindex && txList[2].toUint() == currExit.oindex;
        bool secondInput = txList[3].toUint() == currExit.blocknum && txList[4].toUint() == currExit.txindex && txList[5].toUint() == currExit.oindex;

        if(!firstInput && !secondInput) {
            ChallengeFailure(msg.sender, exitId);
            return;
        }

        bool exists = checkProof(blocknum, txindex, txBytes, proof);

        if (exists) {
            require(currExit.amount > 0);

            uint256 burn;
            if (currExit.owner.balance < currExit.amount) {
                burn = currExit.owner.balance;
            } else {
                burn = currExit.amount;
            }

            currExit.owner.send(-burn);

            exits[exitId] = Exit({
                owner: address(0),
                amount: 0,
                blocknum: 0,
                txindex: 0,
                oindex: 0,
                started_at: 0
            });

            exitQueue.remove(exitId);

            ChallengeSuccess(msg.sender, exitId);
        } else {
            ChallengeFailure(msg.sender, exitId);
        }
    }

    // TODO: move into merkle file.
    function checkProof(
        uint256 blocknum,
        uint256 txindex,
        bytes txBytes,
        bytes proof
    ) returns (bool)
    {
        // TODO: might need to adjust depth
        require(proof.length == 15 * 32);

        bytes32 root = childChain[blocknum].root;

        bytes32 otherRoot = keccak256(txBytes);

        // Offset for bytes assembly starts at 32
        uint j = 32;

        for(uint i = 0; i < 15; i++) {
            bytes32 sibling;
            assembly {
                sibling := mload(add(proof, j))
            }
            j += 32;

            if (txindex % 2 == 0) {
                otherRoot = keccak256(otherRoot, sibling);
            } else {
                otherRoot = keccak256(sibling, otherRoot);
            }
            
            txindex = txindex / 2;
        }

        return otherRoot == root;
    }

    // TODO: passively finalize.
    // If root node doesn't finalize, and validators finalize,
    // validators have to pay.
    // Finalizing is an expensive operation if the queue is large.
    function finalize() {
        if (!shouldFinalize()) {
            return;
        }

        lastFinalizedTime = block.timestamp;
        uint256 exitId = exitQueue.pop();
        while(exitId != SafeMath.max()) {
            Exit memory currExit = exits[exitId];

            if (
                isFinalizableTime(currExit.started_at) &&
                currExit.owner != address(0) &&
                currExit.amount > 0
            ) {
                currExit.owner.send(currExit.amount);
                
                exits[exitId] = Exit({
                    owner: address(0),
                    amount: 0,
                    blocknum: 0,
                    txindex: 0,
                    oindex: 0,
                    started_at: 0
                });
                FinalizeExit(msg.sender, exitId);
            }

            exitId = exitQueue.pop();
        }
    }

    // Periodically monitor if we should finalize
    function shouldFinalize() constant returns (bool) {
        // Not used for testing
        // return block.timestamp > lastFinalizedTime + 2 days;
        return true;
    }

    function isFinalizableTime(uint256 timestamp) constant returns (bool) {
        // Not used for testing
        // return block.timestamp > timestamp + 14 days;
        return true;
    }

    function calcPriority(
        uint256 blocknum,
        uint256 txindex,
        uint256 oindex
    ) constant returns (uint256) {
        // For now always allow the earliest block to be in the front
        // of the queue.  Don't care about 7 day cliff.
        return blocknum * 1000000000 + txindex * 10000 + oindex;
    }
}
