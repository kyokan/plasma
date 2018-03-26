#!/usr/bin/env node

var program = require('commander');
var shell = require('shelljs');
var fs = require("fs");

program
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

  shell.mkdir('-p', ['abi','gen']);

  generate(program.filename);
  process.exit(0);
} else {
  console.log("Must provide a filename option.");
  process.exit(1);
}

function generate(filename) {
  const parts = filename.split("/");
  const path = parts.slice(1, parts.length - 1).join("/");

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
