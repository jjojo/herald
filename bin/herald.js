#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

// Platform-specific binary names
const platformMap = {
  linux: "linux",
  darwin: "darwin",
  win32: "windows",
  freebsd: "freebsd",
};

function getBinaryPath() {
  const platform = platformMap[process.platform];
  const ext = process.platform === "win32" ? ".exe" : "";
  const binaryName = `herald${ext}`;

  // Check if binary exists in the bin directory
  const binPath = path.join(__dirname, binaryName);

  if (fs.existsSync(binPath)) {
    return binPath;
  }

  // Fallback: try to find in the extracted directory structure
  const extractedPath = path.join(__dirname, "herald", binaryName);
  if (fs.existsSync(extractedPath)) {
    return extractedPath;
  }

  // Another fallback: try platform-specific name
  if (platform) {
    const platformBinary = path.join(
      __dirname,
      `herald_${platform}_*`,
      binaryName
    );
    try {
      const glob = require("glob");
      const matches = glob.sync(platformBinary);
      if (matches.length > 0) {
        return matches[0];
      }
    } catch (e) {
      // glob not available, continue
    }
  }

  return null;
}

async function runHerald() {
  try {
    const binaryPath = getBinaryPath();

    if (!binaryPath) {
      console.error("Herald binary not found!");
      console.error("Please try reinstalling: npm install herald");
      console.error(
        "Or install manually from: https://github.com/jjojo/herald/releases/latest"
      );
      process.exit(1);
    }

    // Make sure binary is executable
    if (process.platform !== "win32") {
      try {
        fs.chmodSync(binaryPath, "755");
      } catch (e) {
        // Ignore chmod errors
      }
    }

    // Get command line arguments (excluding node and script name)
    const args = process.argv.slice(2);

    // Spawn the Herald binary
    const child = spawn(binaryPath, args, {
      stdio: "inherit",
      cwd: process.cwd(),
    });

    // Handle process signals
    process.on("SIGINT", () => {
      child.kill("SIGINT");
    });

    process.on("SIGTERM", () => {
      child.kill("SIGTERM");
    });

    // Exit with the same code as the child process
    child.on("exit", (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code || 0);
      }
    });

    child.on("error", (error) => {
      console.error("Failed to start Herald:", error.message);
      process.exit(1);
    });
  } catch (error) {
    console.error("Error running Herald:", error.message);
    process.exit(1);
  }
}

// Only run if called directly
if (require.main === module) {
  runHerald();
}

module.exports = { runHerald };
