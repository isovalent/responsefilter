# Publishing to GitHub

This guide explains how to publish the responsefilter plugin to `github.com/isovalent/responsefilter`.

## Prerequisites

- GitHub account with access to create repositories under the `isovalent` organization
- Git installed locally
- Plugin files ready in this directory

## Steps to Publish

### 1. Create GitHub Repository

1. Go to https://github.com/organizations/isovalent/repositories/new
2. Repository name: `responsefilter`
3. Description: `CoreDNS plugin to filter DNS responses based on FQDN and IP CIDR blocklists`
4. Visibility: Public
5. Do NOT initialize with README, .gitignore, or license (we have these already)
6. Click "Create repository"

### 2. Prepare Local Repository

From the `plugin/responsefilter` directory:

```bash
# Initialize git repository
git init

# Add all files
git add .

# Commit
git commit -m "Initial commit: CoreDNS responsefilter plugin"

# Add remote
git remote add origin git@github.com:isovalent/responsefilter.git

# Push to GitHub
git branch -M main
git push -u origin main
```

### 3. Rename README

After pushing, rename `REPO_README.md` to `README.md` in the GitHub repository:

```bash
git mv REPO_README.md README.md
git commit -m "Use main README"
git push
```

### 4. Create Initial Release

1. Go to https://github.com/isovalent/responsefilter/releases/new
2. Tag version: `v0.1.0`
3. Release title: `v0.1.0 - Initial Release`
4. Description:
   ```
   Initial release of the CoreDNS responsefilter plugin.
   
   Features:
   - Filter DNS responses based on FQDN and IP CIDR blocklists
   - Support for IPv4 (A records) and IPv6 (AAAA records)
   - Subdomain matching
   - Multiple CIDR ranges per domain
   
   Usage:
   Add to CoreDNS plugin.cfg:
   ```
   responsefilter:github.com/isovalent/responsefilter
   ```
   ```
5. Click "Publish release"

### 5. Update CoreDNS plugin.cfg

Users can now add the plugin to their CoreDNS builds:

```
# In CoreDNS plugin.cfg, add before the forward plugin:
responsefilter:github.com/isovalent/responsefilter
```

Then rebuild:
```bash
go generate
go build
```

## Files to Include

The following files should be in the repository:

- `README.md` (renamed from REPO_README.md) - Main documentation
- `responsefilter.go` - Core plugin logic
- `setup.go` - Configuration parser
- `go.mod` - Go module definition
- `LICENSE` - Apache 2.0 license
- `.gitignore` - Git ignore rules
- `PUBLISHING.md` - This file (optional, can be removed after publishing)

## Maintenance

### Creating New Releases

```bash
# Tag new version
git tag v0.2.0
git push origin v0.2.0

# Create release on GitHub
# Go to Releases → Draft a new release
# Select the tag and add release notes
```

### Updating Documentation

```bash
# Edit README.md
git add README.md
git commit -m "Update documentation"
git push
```

## Testing the Published Plugin

After publishing, test that it can be consumed:

```bash
# In a CoreDNS repository
echo "responsefilter:github.com/isovalent/responsefilter" >> plugin.cfg
go generate
go build

# Should download and compile the plugin
```

## Support

For issues with the publishing process, contact the Isovalent GitHub administrators.
