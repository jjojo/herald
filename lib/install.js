#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");

// Platform and architecture mapping
const platformMap = {
  linux: "linux",
  darwin: "darwin",
  win32: "windows",
  freebsd: "freebsd",
};

const archMap = {
  x64: "amd64",
  arm64: "arm64",
  arm: "arm",
};

function getPlatformInfo() {
  const platform = platformMap[process.platform];
  const arch = archMap[process.arch];

  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform}-${process.arch}`
    );
  }

  return { platform, arch };
}

function getDownloadUrl(version, platform, arch) {
  const repo = "jjojo/herald";
  const ext = platform === "windows" ? "zip" : "tar.gz";
  const filename = `herald_${platform}_${arch}.${ext}`;

  if (version === "latest") {
    return `https://github.com/${repo}/releases/latest/download/${filename}`;
  } else {
    return `https://github.com/${repo}/releases/download/${version}/${filename}`;
  }
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading Herald from: ${url}`);

    const file = fs.createWriteStream(dest);

    https
      .get(url, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
          // Follow redirect
          return downloadFile(response.headers.location, dest)
            .then(resolve)
            .catch(reject);
        }

        if (response.statusCode !== 200) {
          reject(
            new Error(`Download failed with status: ${response.statusCode}`)
          );
          return;
        }

        response.pipe(file);

        file.on("finish", () => {
          file.close();
          console.log("Download completed");
          resolve();
        });

        file.on("error", (err) => {
          fs.unlink(dest, () => {}); // Delete partial file
          reject(err);
        });
      })
      .on("error", reject);
  });
}

function extractBinary(archivePath, platform, targetDir) {
  console.log("Extracting Herald binary...");

  try {
    if (platform === "windows") {
      // For Windows ZIP files
      execSync(
        `powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${targetDir}' -Force"`,
        {
          stdio: "inherit",
        }
      );
    } else {
      // For tar.gz files
      execSync(`tar -xzf "${archivePath}" -C "${targetDir}"`, {
        stdio: "inherit",
      });
    }

    // Make binary executable on Unix systems
    if (platform !== "windows") {
      const binaryPath = path.join(targetDir, "herald");
      if (fs.existsSync(binaryPath)) {
        fs.chmodSync(binaryPath, "755");
      }
    }

    console.log("Extraction completed");
  } catch (error) {
    throw new Error(`Failed to extract binary: ${error.message}`);
  }
}

async function install() {
  try {
    const { platform, arch } = getPlatformInfo();
    const packageJson = JSON.parse(
      fs.readFileSync(path.join(__dirname, "..", "package.json"), "utf8")
    );
    const version = process.env.HERALD_VERSION || "latest";

    console.log(`Installing Herald for ${platform}-${arch}...`);

    const url = getDownloadUrl(version, platform, arch);
    const ext = platform === "windows" ? "zip" : "tar.gz";
    const archivePath = path.join(__dirname, `herald.${ext}`);
    const binDir = path.join(__dirname, "..", "bin");

    // Ensure bin directory exists
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    // Download the archive
    await downloadFile(url, archivePath);

    // Extract the binary
    extractBinary(archivePath, platform, binDir);

    // Clean up archive
    fs.unlinkSync(archivePath);

    console.log("Herald installed successfully!");
    console.log("You can now use: npx herald --help");
  } catch (error) {
    console.error("Installation failed:", error.message);
    console.error("\nYou can manually download Herald from:");
    console.error("https://github.com/jjojo/herald/releases/latest");
    process.exit(1);
  }
}

// Only run if called directly
if (require.main === module) {
  install();
}

module.exports = { install };
