#!/usr/bin/env node
import minimist from 'minimist';
import { zip } from '../commands/index.js';
const args = process.argv.slice(2);
const argv = minimist(args);
const commands = {
    zip,
};
const command = commands[argv._[0]] || commands.zip;
command(argv);
