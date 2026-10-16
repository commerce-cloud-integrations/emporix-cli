#!/usr/bin/env node
// Runs the epx binary of this platform, which npm installed from the matching
// optional dependency (epx-cli-<platform>-<arch>).
"use strict";

const { spawn } = require("node:child_process");
const { constants } = require("node:os");
const path = require("node:path");
const { optionalDependencies } = require("../package.json");

const windows = process.platform === "win32";
const prefix = "epx-cli-";
const pkg = `${prefix}${process.platform}-${process.arch}`;
const reinstall =
  "Reinstall epx-cli without --omit=optional or --no-optional; in a project whose " +
  "package-lock.json was made on another platform, delete node_modules and " +
  "package-lock.json, then install again.";

// fail prints the message and ends the launcher; it never returns.
function fail(message) {
  console.error(`epx: ${message}`);
  process.exit(1);
}

// locate returns the path of the binary in the platform package that npm installed with
// this epx-cli, at its version: Node's lookup could otherwise reach an older one
// installed elsewhere.
function locate() {
  const wanted = optionalDependencies[pkg];
  if (wanted === undefined) {
    const platforms = Object.keys(optionalDependencies).map((name) => name.slice(prefix.length));
    fail(
      platforms.length === 0
        ? "this is the source of epx-cli, not a packed release: no platform package is listed yet."
        : `no binary exists for ${process.platform}-${process.arch}; epx-cli ships ${platforms.join(", ")}.`,
    );
  }
  let manifest;
  try {
    manifest = require.resolve(`${pkg}/package.json`);
  } catch {
    fail(`the package ${pkg}, which holds the binary, is not installed. ${reinstall}`);
  }
  const dir = path.dirname(manifest);
  const found = require(manifest).version;
  if (found !== wanted) {
    fail(`epx-cli needs ${pkg} ${wanted}, but found ${found} in ${dir}. ${reinstall}`);
  }
  return path.join(dir, "bin", windows ? "epx.exe" : "epx");
}

const binary = locate();
const child = spawn(binary, process.argv.slice(2), { stdio: "inherit" });

// A signal sent to this launcher reaches epx too. On Windows the console delivers
// Ctrl+C to epx itself, and kill() would end epx at once instead; the empty handler
// still keeps this launcher alive until epx exits, so its exit status is kept.
const forward = windows ? () => {} : (signal) => child.kill(signal);
for (const signal of ["SIGINT", "SIGTERM", "SIGHUP"]) {
  process.on(signal, forward);
}

child.on("error", (err) => fail(`cannot run ${binary}: ${err.message}`));
child.on("exit", (code, signal) => {
  if (signal) {
    // End the same way epx ended; Node ignores some signals (SIGPIPE), so the shell's
    // 128 + signal number is the fallback.
    process.removeAllListeners(signal);
    process.kill(process.pid, signal);
    setImmediate(() => process.exit(128 + constants.signals[signal]));
    return;
  }
  // Node always sets the code when no signal ended epx; a missing one is a failure.
  process.exit(code ?? 1);
});
