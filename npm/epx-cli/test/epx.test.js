// Tests of the launcher: it runs from a node_modules tree laid out as npm installs
// it, with a shell script standing in for the epx binary of this platform.
// Run with: node --test npm/epx-cli/test
"use strict";

const assert = require("node:assert/strict");
const { spawn, spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { test } = require("node:test");

const launcher = path.join(__dirname, "..", "bin", "epx.js");
const host = `epx-cli-${process.platform}-${process.arch}`;
const version = `1.2.3-${process.platform}-${process.arch}`;
const skip = process.platform === "win32" ? "the stand-in binary is a shell script" : false;

// The stand-in binary: "args" prints its arguments one per line, "exit N" ends with N,
// "kill SIG" ends by SIG, "trap" prints "ready" and exits 7 on SIGTERM.
const fake = String.raw`#!/bin/sh
case "$1" in
args) shift; printf '%s\n' "$@" ;;
exit) exit "$2" ;;
kill) kill -s "$2" $$ ;;
trap) trap 'exit 7' TERM; echo ready; while :; do sleep 0.05; done ;;
esac
`;

// install lays out node_modules/epx-cli (the launcher, a manifest listing the given
// platforms as aliases of epx-cli@version) and, if platformPackage is set,
// node_modules/<host>, the version found of epx-cli, with the stand-in binary as bin/epx.
function install(t, { platforms = [host], platformPackage = true, mode = 0o755, found = version } = {}) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "epx-cli-"));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const main = path.join(root, "node_modules", "epx-cli");
  fs.mkdirSync(path.join(main, "bin"), { recursive: true });
  fs.copyFileSync(launcher, path.join(main, "bin", "epx.js"));
  const optionalDependencies = Object.fromEntries(platforms.map((name) => [name, `npm:epx-cli@${version}`]));
  fs.writeFileSync(path.join(main, "package.json"), JSON.stringify({ name: "epx-cli", optionalDependencies }));
  if (platformPackage) {
    const dir = path.join(root, "node_modules", host);
    fs.mkdirSync(path.join(dir, "bin"), { recursive: true });
    fs.writeFileSync(path.join(dir, "package.json"), JSON.stringify({ name: "epx-cli", version: found }));
    fs.writeFileSync(path.join(dir, "bin", "epx"), fake, { mode });
  }
  return path.join(main, "bin", "epx.js");
}

function run(script, args) {
  return spawnSync(process.execPath, [script, ...args], { encoding: "utf8" });
}

test("forwards the arguments literally and the output", { skip }, (t) => {
  const args = ["get", "a b", "", "--query=x:'y'", "$HOME", "*"];
  const result = run(install(t), ["args", ...args]);
  assert.equal(result.status, 0);
  assert.equal(result.stdout, args.map((arg) => `${arg}\n`).join(""));
});

test("keeps the exit status of epx", { skip }, (t) => {
  const script = install(t);
  for (const code of [0, 2, 3, 5]) {
    assert.equal(run(script, ["exit", String(code)]).status, code);
  }
});

test("ends by the signal that ended epx", { skip }, (t) => {
  const result = run(install(t), ["kill", "TERM"]);
  assert.equal(result.signal, "SIGTERM");
});

test("exits 128 + n for a signal that Node ignores", { skip }, (t) => {
  const result = run(install(t), ["kill", "PIPE"]);
  assert.equal(result.status, 128 + os.constants.signals.SIGPIPE);
});

test("forwards SIGTERM to epx and waits for it", { skip, timeout: 5000 }, async (t) => {
  // Its own process group, so that a launcher that fails to forward is still cleaned up
  // with the stand-in binary when the test times out.
  const child = spawn(process.execPath, [install(t), "trap"], {
    detached: true,
    stdio: ["ignore", "pipe", "inherit"],
  });
  t.after(() => {
    try {
      process.kill(-child.pid, "SIGKILL");
    } catch {
      // The group already ended.
    }
  });
  await new Promise((resolve) => child.stdout.once("data", resolve));
  child.kill("SIGTERM");
  const [code] = await new Promise((resolve) => child.on("exit", (...end) => resolve(end)));
  assert.equal(code, 7);
});

test("names the supported platforms on another platform", { skip }, (t) => {
  const result = run(install(t, { platforms: ["epx-cli-aix-ppc64"] }), ["version"]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /no binary exists for .*; epx-cli ships aix-ppc64\./);
});

test("says that its source lists no platform package", { skip }, (t) => {
  const result = run(install(t, { platforms: [] }), ["version"]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /this is the source of epx-cli, not a packed release/);
});

test("refuses a binary package of another version", { skip }, (t) => {
  const result = run(install(t, { found: "1.0.0" }), ["version"]);
  assert.equal(result.status, 1);
  assert.ok(result.stderr.includes(`epx-cli needs ${host} ${version}, but found 1.0.0 in `), result.stderr);
  assert.equal(result.stdout, "");
});

test("explains a binary package left out", { skip }, (t) => {
  const result = run(install(t, { platformPackage: false }), ["version"]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, new RegExp(`the package ${host}, which holds the binary, is not installed`));
});

test("reports a binary it cannot run", { skip }, (t) => {
  const result = run(install(t, { mode: 0o644 }), ["version"]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /^epx: cannot run .*bin\/epx: /);
});
