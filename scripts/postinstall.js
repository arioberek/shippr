#!/usr/bin/env node
const { spawnSync } = require('node:child_process');
const { existsSync, mkdirSync } = require('node:fs');
const { join } = require('node:path');

function fail(msg) {
  console.error(msg);
  process.exit(1);
}

const go = spawnSync('go', ['version'], { stdio: 'ignore' });
if (go.status !== 0) {
  fail('Go toolchain is required to install shippr. Please install Go (https://go.dev/dl/) and re-run `npm install`.');
}

const projectRoot = join(__dirname, '..');
const distDir = join(projectRoot, 'dist');
if (!existsSync(distDir)) mkdirSync(distDir, { recursive: true });

const isWindows = process.platform === 'win32';
const outName = isWindows ? 'shippr.exe' : 'shippr';
const outPath = join(distDir, outName);

const build = spawnSync('go', ['build', '-o', outPath, './cmd/git-shippr'], {
  cwd: projectRoot,
  stdio: 'inherit',
});
if (build.status !== 0) {
  fail('Failed to build shippr binary with Go. See output above.');
}

console.log(`shippr installed -> ${outPath}`);
