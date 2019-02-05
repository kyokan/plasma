# Plasma

This project is a golang implementation and extension of the [Minimum Viable Plasma](https://ethresear.ch/t/minimal-viable-plasma/426) specification. We aim to build a Plasma implementation that can withstand the rigors of production deployment while retaining as much trustlessness as possible.

**Note:** while Plasma is rapidly approaching mainnet readiness, it should still be considered alpha-quality software.

## Smart Contract

This implementation uses the FourthState [plasma-mvp-rootchain](https://github.com/FourthState/plasma-mvp-rootchain) smart contract. Kyokan is funding [Authio](https://authio.org/)'s efforts to perform a full security audit of the contract.

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

## Binaries

This project consists of three binaries:

1. `plasmad`, the Plasma node daemon itself.
2. `plasmacli`, a CLI client for a querying `plasmad`.
3. `plasma-harness`, a tool that simplifies local development by managing Ganache and Truffle processes.

You likely won't need to run `plasma-harness` in production.

## Prerequisites

1. [Golang](https://golang.org/doc/install): This is primarily a golang development environment.
2. [dep](https://github.com/golang/dep): We use ```dep``` for our dependency management.
3. [Node.js](https://nodejs.org/en/) and [npm](https://www.npmjs.com/get-npm)
4. [Truffle](http://truffleframework.com/docs/getting_started/installation).
5. [Ganache](https://github.com/trufflesuite/ganache): Currently we use Ganache to test against a root chain.

## Local Development Installation and Setup

### 1. Checkout, install deps, and build:

```bash
mkdir -p $GOPATH/src/github.com/kyokan
git clone https://github.com/kyokan/plasma.git
cd plasma
make deps
make build-plasmad
make build-plasmacli
make build-harness
```

### 2. Start the harness:

```bash
./target/plasma-harness start
```

`plasma-harness` will start `ganache-cli` in deterministic mode with a fixed mnemonic and automatically deploy all smart contracts. Since `ganache-cli` is in deterministic mode, the smart contract's address will always be `0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da`.

### 3. Start `plasmad`:

`plasmad` uses a YAML config file. An example configuration file suitable for local development can be found in the `build` folder. We will assume that you will use that one while starting `plasmad`. To start `plasmad`, run:

```bash
./target/plasmad --config ./build/config-local.yaml start-root
```

### 4. Set up `plasmacli`:

`plasmacli` requires a private key to sign deposits and transactions. It reads the private key from a file on-disk, and defaults to searching for it at `~/.plasma/key`. Since `plasma-harness` runs Ganache, you can use any one of the default Ganache accounts as the private key:

```bash
# private key for Ganache account index 1
mkdir -p ~/.plasma
echo "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f" > ~/.plasma/key
```

### 5. Deposit and send funds:

You're ready to start sending money! Just make a deposit and send funds when you're ready:

```bash
./target/plasmacli deposit 0xF12b5dd4EAD5F743C6BaA640B0216200e89B60Da 1000000
./target/plasmacli send 0x821aea9a577a9b44299b9c15c88cf3087f3b5544 100 
```

Deposits require an on-chain transaction. Once you've deposited, though, new Plasma blocks are created every 100ms and feel effectively instant.