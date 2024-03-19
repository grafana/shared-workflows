#!/usr/bin/env node

import minimist from 'minimist';
import { zip, ZipArgs } from '../commands/index.js';

const args = process.argv.slice(2);
const argv = minimist(args);

const commands: Record<string, (argv: minimist.ParsedArgs) => void> = {
  zip: (argv: minimist.ParsedArgs) => {
    zip(argv as unknown as ZipArgs);
  },
};

const command = commands[argv._[0]] || commands.zip;

command(argv);
