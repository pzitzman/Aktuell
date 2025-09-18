# GitHub Deployment Guide for Aktuell

This guide will help you deploy Aktuell to GitHub with all the CI/CD workflows we've created.

## 🚀 GitHub Repository Setup

1. **Create a new GitHub repository:**
   ```bash
   # Create a new repository on GitHub (replace YOUR_USERNAME)
   # https://github.com/new
   ```

2. **Initialize and push your code:**
   ```bash
   git init
   git add .
   git commit -m "Initial commit: Aktuell - MongoDB Change Streams"
   git branch -M main
   git remote add origin https://github.com/YOUR_USERNAME/aktuell.git
   git push -u origin main
   ```

## 🔧 Repository Configuration

### Required Secrets

Add these secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

| Secret Name | Description | Required For |
|-------------|-------------|--------------|
| `GITHUB_TOKEN` | Automatically provided by GitHub | All workflows |
| `DOCKER_USERNAME` | Docker Hub username (optional) | Docker workflow |
| `DOCKER_PASSWORD` | Docker Hub password (optional) | Docker workflow |

### Repository Permissions

Ensure these permissions are enabled in `Settings > Actions > General`:

- ✅ **Read and write permissions** for GITHUB_TOKEN
- ✅ **Allow GitHub Actions to create and approve pull requests**

## 📋 Workflow Overview

The project includes 5 comprehensive GitHub Actions workflows:

### 1. **CI Workflow** (`.github/workflows/ci.yml`)
**Triggers:** Push, Pull Request
**Features:**
- Tests Go code across versions 1.23, 1.24, 1.25
- Runs React TypeScript builds
- MongoDB integration testing
- Code coverage reporting
- Linting and formatting checks

### 2. **Release Workflow** (`.github/workflows/release.yml`)
**Triggers:** Git tags (v*)
**Features:**
- Multi-platform binary builds (Linux, macOS, Windows)
- Automated GitHub releases with checksums
- Docker image publishing to GHCR
- Release notes generation

### 3. **Docker Workflow** (`.github/workflows/docker.yml`)
**Triggers:** Push to main, Pull Request, Manual
**Features:**
- Multi-architecture Docker builds (amd64, arm64)
- Trivy security scanning
- GHCR (GitHub Container Registry) publishing
- Automated tagging and metadata

### 4. **Dependencies Workflow** (`.github/workflows/dependencies.yml`)
**Triggers:** Weekly schedule, Manual
**Features:**
- Automated Go dependency updates
- React/npm dependency updates
- Automated testing of updates
- Pull request creation for approved updates

### 5. **CodeQL Security Workflow** (`.github/workflows/codeql.yml`)
**Triggers:** Push, Pull Request, Weekly schedule
**Features:**
- Security analysis for Go and JavaScript
- Vulnerability detection
- SARIF result uploads
- Security advisory integration

## 🏃‍♂️ Testing Your Workflows

### 1. Test CI Workflow
```bash
# Push a change to trigger CI
echo "# Test CI" >> README.md
git add README.md
git commit -m "test: trigger CI workflow"
git push
```

### 2. Test Release Workflow
```bash
# Create a release tag
git tag v1.0.0
git push origin v1.0.0
```

### 3. Test Docker Workflow
```bash
# Docker workflow runs automatically on main branch pushes
# Check the "Actions" tab in your GitHub repository
```

## 📊 Monitoring Workflows

### GitHub Actions Dashboard
- Go to your repository's **Actions** tab
- Monitor workflow runs and logs
- Review security alerts in **Security** tab

### Status Badges
Add these badges to your README (replace YOUR_USERNAME):

```markdown
[![CI](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/ci.yml)
[![Release](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/release.yml/badge.svg)](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/release.yml)
[![Docker](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/docker.yml/badge.svg)](https://github.com/YOUR_USERNAME/aktuell/actions/workflows/docker.yml)
```

## 🐳 Using Published Docker Images

After the workflows run, your Docker images will be available at:

```bash
# Pull the latest image
docker pull ghcr.io/YOUR_USERNAME/aktuell:latest

# Run the container
docker run -p 8080:8080 -p 3001:3001 ghcr.io/YOUR_USERNAME/aktuell:latest
```

## 📦 Using Release Binaries

Download pre-compiled binaries from:
`https://github.com/YOUR_USERNAME/aktuell/releases`

Available for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## 🔒 Security Features

### Automated Security Scanning
- **Trivy**: Container vulnerability scanning
- **CodeQL**: Static code analysis
- **Dependency scanning**: Automated vulnerability detection
- **Go security advisories**: Integration with Go vulnerability database

### Security Best Practices Implemented
- Distroless Docker images
- Non-root container execution
- Minimal attack surface
- Automated security updates
- SARIF result integration

## 🚨 Troubleshooting

### Common Issues

**1. Workflow fails with permission errors:**
- Check repository permissions in Settings > Actions
- Ensure GITHUB_TOKEN has write permissions

**2. Docker workflow fails:**
- Verify Docker Hub credentials (if using)
- Check GHCR permissions

**3. Go build fails:**
- Ensure go.mod is properly configured
- Check Go version compatibility

**4. React build fails:**
- Verify package.json exists in client directory
- Check Node.js version compatibility

### Debugging Workflows
- Check workflow logs in Actions tab
- Enable debug logging by setting repository secret: `ACTIONS_STEP_DEBUG=true`
- Review individual step outputs

## 📈 Next Steps

1. **Custom Domain**: Set up custom domain for GitHub Pages
2. **Monitoring**: Integrate with monitoring services
3. **Documentation**: Add comprehensive API documentation
4. **Performance**: Add performance benchmarks to CI
5. **Integration Tests**: Extend integration test coverage

## 🎯 Production Deployment

For production deployment, consider:

1. **Environment Variables**: Configure production MongoDB URIs
2. **Secrets Management**: Use GitHub secrets for sensitive data
3. **Load Balancing**: Use multiple container instances
4. **Monitoring**: Integrate with APM tools
5. **Backup Strategy**: Implement data backup procedures

---

**Your Aktuell project is now ready for professional GitHub deployment with comprehensive CI/CD!** 🚀