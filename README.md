# Plasma

This project is a golang implementation and extension of the [Minimum Viable Plasma](https://ethresear.ch/t/minimal-viable-plasma/426) specification. Our goals are the following:

1. Extend the Plasma MVP.
2. Build a pluggable architecture that enables developers to build decentralized applications on top of a Plasma chain.

## Architecture

This implementation uses a proof-of-authority model. The authority owns root nodes that facilitate child chain transactions and create blocks on the Plasma Contract. Users on the child chain ensure security by running validator nodes, which monitor the validity of each plasma block, and perform exits if problems are detected.

The following are three main parts of the system:

### Root Nodes:

1. Processing deposits and exits via the Plasma smart contract it owns.
2. Processing transactions and packaging them into blocks.
3. Broadcasting blocks to validator nodes.
4. Reporting new blocks to the Plasma smart contract.

### Validator Nodes:

1. Checking the validity of every block emitted by a root node.
1. Checking the validity of block headers on the Plasma contract.
2. Exiting the Plasma chain if malfeasance is detected.

### Plasma Contract:

1. A smart contract on the Ethereum root chain.
2. Supports deposits, block submission, exits, and challenges.

## Block Submission

Blocks are submitted under the following conditions:

1. When a deposit transaction is received.
2. When 65,535 transactions are in the mempool.
3. Every 500 ms.

Every hour, the root node puts the last hour's worth of transactions into a Merkle tree and sends the Merkle root to the Plasma contract.

## Prerequisites

1. [Golang](https://golang.org/doc/install): This is primarily a golang development environment.
2. [Glide](https://github.com/Masterminds/glide#install): We use glide for our dependency management.
3. [Truffle](http://truffleframework.com/docs/getting_started/installation): For convenience, truffle is currently used to migrate Plasma contracts to the Ethereum root chain.
4. [Geth](https://github.com/ethereum/go-ethereum/wiki/Installing-Geth): To run a local private chain for testing with web sockets and lower mining difficulty.

## Installation and Setup

1. Checkout, install deps, and build:

```
mkdir -p $GOPATH/src/github.com/kyokan
git clone https://github.com/kyokan/plasma.git
cd plasma
glide install
make
```

2. Setup local private chain or testnet:

This is a simple example of setting up a private chain.  For more details refer to the [Private Network](https://github.com/ethereum/go-ethereum/wiki/Private-network) written by Ethereum.

```
mkdir -p ~/geth/chain
cd ~/geth
echo '{    "config": {        "chainId": 15,        "homesteadBlock": 0,        "eip155Block": 0,        "eip158Block": 0    },    "difficulty": "1024",    "gasLimit": "10000000",    "alloc": {        "0x44a5cae1ebd47c415630da1e2131b71d1f2f5803": { "balance": "1000000000000000000000" }    }}' > genesis.json
geth --datadir chain init genesis.json
geth account new --datadir chain
geth --datadir chain --rpc --ws --mine --unlock [YOUR_ADDRESS]
```

3. Deploy contracts:

Make sure to save the resulting Plasma contract address to be used later.

```
cd $GOPATH/src/github.com/kyokan/plasma/contracts
truffle migrate --network development
```

4. Run root nodes and validators:

Adding the ws scheme will allow us to process deposit events.

```
plasma start --node-url ws://localhost:8545 --contract-addr [ADD_PLASMA_CONTRACT_ADDRESS_HERE] &
plasma validate --node-url ws://localhost:8545 --contract-addr [ADD_PLASMA_CONTRACT_ADDRESS_HERE] &
```

## CLI Usage

**Note:** Some of these CLI commands are still in development.

```
NAME:
   Plasma - A secure and scalable solution for decentralized applications.

USAGE:
   plasma [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     start     Starts running a Plasma root node.
     validate  Starts running a Plasma validator node.
     utxos     Prints UTXOs for the given address.
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db value             Filepath for Plasma's database. (default: "~/.plasma")
   --node-url value       Full URL to a running geth node. (default: "http://localhost:8545")
   --contract-addr value  Plasma contract address. (default: "0xd1d7dddd82189ea452eb5e104d13f0ca367887d9")
   --help, -h             show help
   --version, -v          print the version
```

## Root Node API
### Send Transaction
Send a transaction to other participants.
#### Parameters
|Name|Type|Required|Description|
|---|---|---|---|
|from|Address|Yes|Sender of transaction|
|to|Address|Yes|Recipient of transaction|
|amount|Float|Yes|Amount to send|
#### Sample
```
curl http://localhost:8643/rpc -H "Content-Type: application/json" -X POST --data '{ "method": "Transaction.Send", "params": [{"From":"0x627306090abaB3A6e1400e9345bC60c78a8BEf57","To":"0xf17f52151EbEF6C7334FAD080c5704D77216b732","Amount":"3"}], "id":1}'
```

## Example Applications

Currently there are a growing number of decentralized applications using devices that offer a utility (such as routing network packets) and simultaneously leverage this data to calculate micro payments in a “pay-as-you-go” model.  Solutions such as state-channels help limit costs, but come with complexities when there are thousands of nodes, requiring thousands of channels to be opened and/or chained.  Plasma offers a great alternative solution in these scenarios because in reality the payment contract is between two parties: the decentralized app which owns these devices, and the customer using these devices.  In this way, the decentralized app can maintain their own Plasma child chain, pooling together transactions reported from their devices.  They can then fine tune their costs based on the size of the block headers and frequency these blocks are reported to the Plasma contract.

At the same time, customers of the decentralized app may run standardized validator nodes (provided by this implementation) on a VPC.  This ensures security for customers, since validator nodes will automatically exit in case of byzantine behavior from the decentralized app.  Validators will run autonomously handling the complexities of verifications, exits, and challenges.  This makes them easy to use for normal consumers.  In practice it makes sense for customers to group together into validator pools, since running a VPC could be expensive.

## Future Applications

We are actively working on extending the Plasma MVP to beyond payments.  In the use-case above, it would be beneficial if the decentralized app could "plug in" to Plasma by executing proprietary business logic while still getting the security guarantees of the Ethereum root chain.  A simple example is a contract that calculates how much network bandwidth customers use.  This gives several advantages:

1. Customers have greater visibility of how they will be charged before agreeing to use the service.
2. Customers will have confidence in the security because validators will submit challenge proofs incase of incorrect payment calculations.  Currently users would need to trust the amount charged was calculated correctly, and it would be realistically difficult to validate, especially in high transaction volume situations.
3. Decentralized Apps can more quickly develop without worrying about scalability and security.

Implementing this vision will come down to two major parts:

#### Pluggable Framework for Decentralized Apps

We will create a layer on top of Root Nodes that will allow decentralized apps to run their own business logic.  Our vision is to provide a development framework and/or protocol that is easy to use, while still being secure.  To achieve this, the functionality provided will be incrementally released as specific subsets of state transitions whose priority is determined by real world decentralized app use cases.  Decentralized apps will likely integrate with this layer via pluggable golang code and/or rpc protocols.

#### Contract Security Guarantees

Root Nodes will submit state transition types that are hashed into separate merkle roots stored in Plasma blocks on the contract.  The Plasma contract may track simplified and/or recent state transition functions and storage for each instance, to help increase security.  In the case of byzantine behavior, validators can submit proofs that refute these transitions by re-running state transitions within a challenge window and verifying that the resulting merkle roots match. 

### Conclusion

Decentralized apps and customers may then agree upon pre-defined state transitions on the child chain, with Plasma based security guarantees of the root chain. In this model two parties will review state transition functions beforehand that detail what data will be used to determine costs (i.e. network bandwidth). The devices will then broadcast state transition transactions to the Root Node, which will execute the transition on the child chain. The Root Node will then create proofs of the state transitions and store them in the block header. Validators will then autonomously validate that the latest submitted blocks look legitimate by re-running the state transitions.  This model illustrates how we plan to move towards a pluggable Plasma framework that extends beyond payments.
