import {ChildProcess, spawn} from 'child_process';
import path = require('path');

export class Ganache {
  private readonly mnemonic: string;

  private readonly port: number;

  private readonly deterministic: boolean;

  private readonly verboseRPC: boolean;

  private readonly networkID: string;

  private readonly dbPath: string;

  private readonly unlockedAccounts: number[];

  private readonly blockTime: number;

  private ganache: ChildProcess | null = null;

  constructor (mnemonic: string, port: number, deterministic: boolean, verboseRPC: boolean, networkID: string, dbPath: string, unlockedAccounts: number[], blockTime: number) {
    this.mnemonic = mnemonic;
    this.port = port;
    this.deterministic = deterministic;
    this.verboseRPC = verboseRPC;
    this.networkID = networkID;
    this.dbPath = dbPath;
    this.unlockedAccounts = unlockedAccounts;
    this.blockTime = blockTime;
  }

  public async start () {
    if (this.ganache) {
      throw new Error('Ganache is already running.');
    }

    const args = this.buildArgs();
    this.ganache = spawn(path.resolve(__dirname, '..', '..', 'node_modules', '.bin', 'ganache-cli'), args);
  }

  public async stop () {
    this.ensureRunning();
    this.ganache!.kill('SIGTERM');
  }

  public stdout () {
    this.ensureRunning();
    return this.ganache!.stdout;
  }

  public stdin () {
    this.ensureRunning();
    return this.ganache!.stdin;
  }

  public stderr () {
    this.ensureRunning();
    return this.ganache!.stderr;
  }

  public onClose (fn: (code: number, signal: string) => void) {
    this.ensureRunning();
    this.ganache!.on('close', fn);
  }

  private buildArgs (): string[] {
    const args: string[] = [];

    if (this.mnemonic) {
      args.push('-m', this.mnemonic);
    }

    if (this.port) {
      args.push('-p', this.port.toString());
    }

    if (this.deterministic) {
      args.push('--deterministic');
    }

    if (this.verboseRPC) {
      args.push('--verbose-rpc');
    }

    if (this.networkID) {
      args.push('--networkId', this.networkID);
    }

    if (this.dbPath) {
      args.push('--db', this.dbPath);
    }

    if (this.unlockedAccounts) {
      args.push('--unlock', this.unlockedAccounts.join(','));
    }

    if (this.blockTime) {
      args.push('-b', this.blockTime.toString());
    }

    return args;
  }

  private ensureRunning () {
    if (!this.ganache) {
      throw new Error('Ganache is not running.');
    }
  }
}