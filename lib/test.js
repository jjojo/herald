#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

function findHeraldBinary() {
  const binDir = path.join(__dirname, "..", "bin");
  const possibleNames = ["herald", "herald.exe"];

  for (const name of possibleNames) {
    const binaryPath = path.join(binDir, name);
    if (fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  }

  // In development, check if there's a herald binary in the project root
  const devBinaryPath = path.join(__dirname, "..", "..", "herald");
  if (fs.existsSync(devBinaryPath)) {
    return devBinaryPath;
  }

  return null;
}

async function testHerald() {
  console.log("Testing Herald installation...");

  const binaryPath = findHeraldBinary();

  if (!binaryPath) {
    console.error("❌ Herald binary not found");
    process.exit(1);
  }

  console.log(`✅ Herald binary found at: ${binaryPath}`);

  return new Promise((resolve, reject) => {
    const child = spawn(binaryPath, ["--help"], {
      stdio: "pipe",
      cwd: process.cwd(),
    });

    let output = "";
    child.stdout.on("data", (data) => {
      output += data.toString();
    });

    child.stderr.on("data", (data) => {
      output += data.toString();
    });

    child.on("exit", (code) => {
      if (code === 0 && output.includes("Herald")) {
        console.log("✅ Herald is working correctly");
        console.log("Test output:");
        console.log(output.substring(0, 200) + "...");
        resolve();
      } else {
        console.error("❌ Herald test failed");
        console.error("Output:", output);
        reject(new Error(`Herald test failed with exit code: ${code}`));
      }
    });

    child.on("error", (error) => {
      console.error("❌ Failed to run Herald:", error.message);
      console.error("Binary path:", binaryPath);
      console.error("Current directory:", process.cwd());
      reject(error);
    });
  });
}

if (require.main === module) {
  testHerald().catch((error) => {
    console.error("Test failed:", error.message);
    process.exit(1);
  });
}

module.exports = { testHerald };
