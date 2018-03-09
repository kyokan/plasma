# Plasma MVP

This project is a golang implementation of Vitalik Buterin's [Minimum Viable Plasma](https://ethresear.ch/t/minimal-viable-plasma/426) specification. It is very much a work-in-progress. Our goals are the following:

1. Implement the Plasma MVP.
2. Build a pluggable architecture on top of the MVP in order to enable people to build applications on top of a Plasma chain.

## Architecture

This implementation uses a proof-of-authority model. There are two node types: root nodes, and validator nodes. Root nodes are responsible for:

1. Processing deposits and exits via the Plasma smart contract it owns.
2. Processing transactions and packaging them into blocks.
3. Broadcasting blocks to validator nodes.
4. Reporting new blocks to the Plasma smart contract.

Validator nodes are responsible for:

1. Checking the validity of every block emitted by a root node.
2. Exiting the Plasma chain if malfeasance is detected.

## Block Emission

Blocks are emitted under the following conditions:

1. When a deposit transaction is received.
2. When 65,535 transactions are in the mempool.
3. Every 500 ms.

Every hour, the root node puts the last hour's worth of transactions into a Merkle tree and sends the Merkle root to the Plasma contract.

## Installation

This project uses glide to manage its dependencies.  To install them locally run:

```
glide install
```

## Migrate Contracts for Development

For development we are using default ganache settings.  To upload contracts run the following:

```
truffle migrate --network development
```

## CLI Usage

**Note:** Some of these CLI commands are still in development.

```
NAME:
   Plasma - Demonstrates what an example Plasma blockchain can do.

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