# Setting up NPM Publishing with GitHub Actions

## Prerequisites

1. **NPM Account**: You need an npm account with publish permissions
2. **NPM Token**: You need to create an NPM access token

## Step 1: Create NPM Access Token

1. Go to [npmjs.com](https://www.npmjs.com) and log in
2. Click on your profile picture → **Access Tokens**
3. Click **Generate New Token** → **Automation**
4. Give it a name like "GitHub Actions Herald"
5. Copy the token (starts with `npm_`)

## Step 2: Add Token to GitHub Secrets

1. Go to your GitHub repository: https://github.com/jjojo/herald
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Name: `NPM_TOKEN`
5. Value: Paste your NPM token
6. Click **Add secret**

## Step 3: Test the Workflow

Once the secret is added, the workflow will automatically run when you:

1. **Create a new release**:

   ```bash
   git tag v3.0.4
   git push origin v3.0.4
   ```

2. **Or use Herald itself**:
   ```bash
   herald release
   ```

## Workflow Features

The automated workflow will:

1. ✅ **Build Go binaries** with GoReleaser
2. ✅ **Create GitHub release** with all platform binaries
3. ✅ **Update package.json** version to match release tag
4. ✅ **Publish to NPM** automatically
5. ✅ **Verify publication** was successful
6. ✅ **Create summary** with installation instructions

## Manual Override

If you need to publish manually:

```bash
# Update version in package.json
npm version patch  # or minor, major

# Publish to npm
npm publish
```

## Troubleshooting

- **403 Forbidden**: Check that your NPM_TOKEN has publish permissions
- **Version conflict**: Make sure the version in package.json matches the git tag
- **Token expired**: Generate a new NPM token and update the GitHub secret
