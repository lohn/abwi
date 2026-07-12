#!/usr/bin/env node
const { spawnSync } = require('child_process');
const path = require('path');

const platformPackages = {
  darwin: {
    arm64: '@abwi/cli-darwin-arm64',
    x64: '@abwi/cli-darwin-x64',
  },
  linux: {
    arm64: '@abwi/cli-linux-arm64',
    x64: '@abwi/cli-linux-x64',
  },
  win32: {
    arm64: '@abwi/cli-win32-arm64',
    x64: '@abwi/cli-win32-x64',
  },
};

const platform = process.platform;
const arch = process.arch;
const pkg = platformPackages[platform]?.[arch];

if (!pkg) {
  console.error(`Error: Unsupported platform: ${platform}-${arch}`);
  console.error('Please install abwi manually:');
  console.error('  https://github.com/lohn/abwi/releases');
  process.exit(1);
}

const binaryName = platform === 'win32' ? 'abwi.exe' : 'abwi';

let binaryPath;
try {
  binaryPath = require.resolve(path.join(pkg, binaryName));
} catch (e) {
  console.error(`Error: Could not find binary for ${pkg}.`);
  console.error(
    'Your platform may not be supported. Please install abwi manually:',
  );
  console.error('  https://github.com/lohn/abwi/releases');
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
});
process.exit(result.status ?? 1);
