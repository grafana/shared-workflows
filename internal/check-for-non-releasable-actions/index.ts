import { NodeFileSystem } from "./filesystem";
import { main } from "./main";
import config from "../../release-please-config.json" assert { type: "json" };

const fs = new NodeFileSystem();
process.exit(await main(fs, config));
