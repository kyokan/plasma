#!/usr/bin/env node

var program = require('commander');
var shell = require('shelljs');
var fs = require("fs");

program
  .option('-n, --network [value]', 'Network to deploy to')
  .option('-c, --clean', 'Clean build folders')
  .option('-f, --filename [value]', 'Single file to generate')
  .version('0.1.0');

program.parse(process.argv);
 
if (program.clean) {
  console.log("Removing directories: build, abi, and gen.")
  shell.rm('-rf', 'build');
  shell.rm('-rf', 'abi');
  shell.rm('-rf', 'gen');
  process.exit(0);
}

console.log('Generate go contract models using truffle generated abis.');

if (program.filename) {
  console.log('Truffle build filename:', program.filename);

  generate(program.filename);
  
  process.exit(0);
}

// Do the default thing
generateDefault();

function generateDefault() {
  console.log('Do default generate logic...');

  const network = program.network ? program.network : "ganache";

  console.log(`Using network: ${network}`);

  shell.exec(`truffle migrate --network ${network} --reset`);
  generate('./build/contracts/Plasma.json');
  generate('./build/contracts/PriorityQueue.json');

  process.exit(0);
}

function generate(filename) {
  shell.mkdir('-p', ['abi','gen']);

  const parts = filename.split("/");
  const path = parts.slice(2, parts.length - 1).join("/");

  shell.mkdir('-p', [`abi/${path}`, `gen/${path}`])

  const data = fs.readFileSync(filename, {encoding: 'utf8'});
  const contract = JSON.parse(data);
  const abi = contract.abi;

  if (abi.length === 0) {
    throw new Error("abi must exist");
  }

  const name = contract.contractName;
  const newFilename = (
    name.replace(/([a-z])([A-Z])/g, '$1_$2').replace(/([A-Z])([A-Z][a-z])/g, '$1_$2')
  ).toLowerCase();

  const abiPath = `abi/${path}/${name}.abi`;
  const genPath = `gen/${path}/${newFilename}.go`;

  console.log(`Generating abi file ${abiPath}`);
  console.log(`Generating gen file ${genPath}`);

  fs.writeFileSync(abiPath, JSON.stringify(abi));

  shell.exec(`abigen --abi ${abiPath} --pkg contracts --type ${name} --out ${genPath}`);
}
