import {assert} from 'chai';
import {Ganache} from './Ganache';

export class GanacheBuilder {
  private mnemonic: string = 'candy maple cake sugar pudding cream honey rich smooth crumble sweet treat';

  private port: number = 8545;

  private deterministic: boolean = true;

  private verboseRPC: boolean = true;

  private networkID: string = 'development';

  private dbPath: string = '/tmp/ganache';

  private unlockedAccounts: number[] = [];

  private blockTime: number = 3;

  public withMnemonic (mnemonic: string): GanacheBuilder {
    assert.lengthOf(mnemonic.split(' '), 12, 'must be a 12 word mnemonic');
    this.mnemonic = mnemonic;
    return this;
  }

  public withPort (port: number): GanacheBuilder {
    this.port = port;
    return this;
  }

  public withDeterministic (deterministic: boolean): GanacheBuilder {
    this.deterministic = deterministic;
    return this;
  }

  public withVerboseRPC (verboseRPC: boolean): GanacheBuilder {
    this.verboseRPC = verboseRPC;
    return this;
  }

  public withNetworkID (networkID: string): GanacheBuilder {
    this.networkID = networkID;
    return this;
  }

  public withDbPath (dbPath: string): GanacheBuilder {
    assert.isAtLeast(dbPath.length, 1, 'dbPath must not be empty');
    this.dbPath = dbPath;
    return this;
  }

  public withUnlockedAccounts (unlockedAccounts: number[]): GanacheBuilder {
    this.unlockedAccounts = unlockedAccounts;
    return this;
  }

  public withBlockTime (blockTime: number): GanacheBuilder {
    this.blockTime = blockTime;
    return this;
  }

  public build (): Ganache {
    return new Ganache(
      this.mnemonic,
      this.port,
      this.deterministic,
      this.verboseRPC,
      this.networkID,
      this.dbPath,
      this.unlockedAccounts,
      this.blockTime,
    );
  }
}