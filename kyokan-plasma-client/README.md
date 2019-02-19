# kyokan-plasma-client

This is the JavaScript SDK for Kyokan's Plasma implementation. Plasma enables instant, low-fee, trustless payments on the Ethereum blockchain.

## Getting Started

### Installation

First, install the Node module via NPM or yarn:

```bash
npm install --save kyokan-plasma-client
# or
yarn add kyokan-plasma client
```

Since the client is written in Typescript, type definitions are included inside the package itself.

### Usage

You'll need the following things to use Plasma:

1. A `Web3` instance that supports signing, either through a custom `Web3` provider or by pointing to an Ethereum node with unlocked accounts.
2. The hostname and port of a running `plasmad` node.
3. The address of the Plasma smart contract.
4. Your private key.

Once you have those things, create an instance of the `Plasma` class like this:

```javascript
import Plasma from 'kyokan-plasma-client';

const web3 = new Web3('...');

const client = new Plasma({
	web3: web3,
	contractAddress: '0x123abc',
	rootUrl: 'localhost:6545',
	privateKey: '0x12345',
});
```

To deposit funds, use the `deposit()` method:

```javascript
// all values are in Wei
const depositReceipt = await client.deposit('1000000000000000000');
```

To spend that deposit, pass in the `depositNonce` included in the receipt to `send`:

```javascript
const tx = await client.send('0x1234', '999', '0', depositReceipt.nonce);
```

To send funds once you've spent your deposits, simply call `send` without a `depositNonce`. The client will pick the correct UTXOs and spend them for you.