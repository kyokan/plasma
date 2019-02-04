import {GanacheBuilder} from './lib/GanacheBuilder';
import * as path from 'path';
import * as os from 'os';
import chalk from 'chalk';
import {Ganache} from './lib/Ganache';
import rimraf from 'rimraf';
import Web3 from 'web3';
import {wait} from './lib/wait';
import * as fs from 'fs';
import {ChildProcess, spawn} from 'child_process';
import {Config} from './Config';
import {SharedWeb3} from './lib/SharedWeb3';

const tmpDir = path.resolve(os.tmpdir(), `plasma-${Date.now()}`);
const ganacheDir = path.join(tmpDir, 'ganache');
const plasmaDbDir = path.join(tmpDir, 'plasma');
const {log} = console;
let ganache: Ganache;
let plasma: ChildProcess;
let testsFinished = false;

async function setup () {
  logRunner('Creating test directories...');
  fs.mkdirSync(tmpDir);
  fs.mkdirSync(ganacheDir);
  fs.mkdirSync(plasmaDbDir);
  logRunner('Done.');

  ganache = new GanacheBuilder()
    .withMnemonic('candy maple cake sugar pudding cream honey rich smooth crumble sweet treat')
    .withPort(8545)
    .withDeterministic(true)
    .withVerboseRPC(true)
    .withNetworkID('development')
    .withDbPath(ganacheDir)
    .withUnlockedAccounts([0, 1, 2, 3, 4, 5, 6, 7, 8, 9])
    .withBlockTime(1)
    .build();

  logRunner('Starting ganache...');
  await ganache.start();
  ganache.stdout().on('data', (d: Buffer) => logPrefixed('ganache-stdout', 'blue', d.toString('utf-8')));
  ganache.stderr().on('data', (d: Buffer) => logPrefixed('ganache-stderr', 'blue', d.toString('utf-8')));
  ganache.onClose((code, signal) => {
    if (testsFinished) {
      logRunner('Ganache successfully stopped.');
    } else {
      logRunner(`Ganache stopped unexpectedly! Code: ${code}, Signal: ${signal}`);
      process.exit(1);
    }
  });
  logRunner('Done. Waiting for ganache to initialize...');

  const w3 = SharedWeb3.getShared();
  let started = false;

  for (let i = 0; i < 3; i++) {
    try {
      logRunner('Checking initialization...');
      await checkGanacheStatus(w3);
      started = true;
    } catch (e) {
      logRunner('Ganache isn\'t started yet. Retrying in 3 seconds...');
      await wait(3000);
    }
  }

  if (!started) {
    logRunner('Timed out waiting for Ganache to start.');
    throw new Error('Timed out.');
  }

  logRunner('Done. Migrating contract...');
  await migrateContract();
  logRunner('Done. Starting Plasma node...');
  await startPlasma();
}

async function teardown () {
  testsFinished = true;
  await ganache.stop();
  await plasma.kill('SIGTERM');

  if (process.env.KEEP_FILES) {
    return;
  }

  logRunner('Cleaning up temporary files...');
  await new Promise((resolve, reject) => rimraf(tmpDir, (err) => {
    if (err) {
      return reject(err);
    }

    return resolve();
  }));
}

async function checkGanacheStatus (w3: Web3) {
  return new Promise((resolve, reject) => {
    w3.eth.getBlockNumber().then(resolve).catch(reject);
    setTimeout(() => reject('timed out'), 100);
  });
}

async function migrateContract () {
  return new Promise((resolve, reject) => {
    let finished = false;
    const truffle = spawn(path.resolve(__dirname, '..', 'node_modules', '.bin', 'truffle'), [
      'migrate',
      '--reset',
    ], {
      cwd: path.resolve(__dirname, '..', '..', 'plasma-mvp-rootchain'),
    });
    truffle.stdout.on('data', (d: Buffer) => {
      const dStr = d.toString('utf-8');
      if (dStr.indexOf(Config.PLASMA_CONTRACT_ADDRESS) > -1) {
        finished = true;
      }

      logPrefixed('truffle-stdout', 'cyan', dStr);
    });
    truffle.stderr.on('data', (d: Buffer) => logPrefixed('truffle-stderr', 'cyan', d.toString('utf-8')));
    truffle.on('close', (code, signal) => {
      if (finished) {
        resolve();
        return;
      }

      logRunner(`Migrations stopped unexpectedly! Code: ${code}, Signal: ${signal}`);
      reject(new Error('migrations terminated unexpectedly'));
    });
  });
}

async function startPlasma () {
  return new Promise((resolve, reject) => {
    setTimeout(() => reject(new Error('timed out starting plasma')), 10000);
    plasma = spawn(path.resolve(__dirname, '..', '..', 'target', 'plasma'), [
      'start-root',
      '--config',
      path.resolve(__dirname, '..', 'config', 'test-config.yml'),
      '--db',
      plasmaDbDir
    ]);
    plasma.stdout.on('data', (d:Buffer) => logPrefixed('plasma-stdout', 'yellow', d.toString('utf-8')));
    plasma.stderr.on('data', (d:Buffer) => {
      const dStr = d.toString('utf-8');
      if (dStr.match(/Started RPC server on port 6545/im)) {
        resolve();
      }

      logPrefixed('plasma-stderr', 'yellow', d.toString('utf-8'));
    });
    plasma.on('close', async (code, signal) => {
      if (testsFinished) {
        logRunner('Plasma successfully stopped.');
        return;
      }

      logRunner(`Plasma stopped unexpectedly! Code: ${code}, Signal: ${signal}`);
      // ganache sometimes sticks around even after process death,
      // so we stop it manually here
      await ganache.stop();
      process.exit(1);
    });
  });
}

function logRunner (msg: string) {
  logPrefixed('runner', 'green', msg);
}

function logPrefixed (prefix: string, color: 'blue' | 'green' | 'red' | 'cyan' | 'yellow', rawMsg: string) {
  rawMsg.trim().split('\n').forEach((line: string) => log(chalk[color](`[${prefix}]`), line));
}

before(async function () {
  this.timeout(60000);
  await setup();
});

after(async () => {
  await teardown();
});