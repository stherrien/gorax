# CI/CD Setup Checklist

Use this checklist to ensure your CI/CD pipeline is fully configured and operational.

## Initial Setup

### Repository Configuration

- [ ] **Enable GitHub Actions**
  - Go to: Settings > Actions > General
  - Ensure "Allow all actions and reusable workflows" is selected

- [ ] **Configure Workflow Permissions**
  - Go to: Settings > Actions > General > Workflow permissions
  - Select: "Read and write permissions"
  - Check: "Allow GitHub Actions to create and approve pull requests"

- [ ] **Enable GitHub Container Registry**
  - Automatic for public repositories
  - For private: Settings > Packages (enabled by default)

### Branch Protection

- [ ] **Protect Main Branch**
  - Go to: Settings > Branches > Add rule
  - Branch name pattern: `main`
  - Check: "Require status checks to pass before merging"
  - Check: "Require branches to be up to date before merging"
  - Select required checks:
    - [ ] Go Tests
    - [ ] Go Lint
    - [ ] Frontend Tests
    - [ ] Frontend Lint
    - [ ] Build Verification

- [ ] **Protect Dev Branch** (optional but recommended)
  - Branch name pattern: `dev`
  - Same settings as main

### Secrets Configuration

#### Optional Secrets
- [ ] **CODECOV_TOKEN**
  - Required for private repositories
  - Get from: https://codecov.io
  - Go to: Settings > Secrets and variables > Actions > New repository secret

- [ ] **STAGING_URL**
  - Your staging environment URL
  - Default: https://staging.gorax.dev

#### For Custom Deployments (if applicable)
- [ ] **AWS_ACCESS_KEY_ID** (if using AWS)
- [ ] **AWS_SECRET_ACCESS_KEY** (if using AWS)
- [ ] **KUBE_CONFIG** (if using Kubernetes)
- [ ] **SSH_PRIVATE_KEY** (if using SSH deployments)
- [ ] **DEPLOY_TOKEN** (for custom deployment methods)

## Workflow Customization

### Staging Deployment
- [ ] **Edit deploy-staging.yml**
  - Update deployment commands for your infrastructure
  - Options: Kubernetes, Docker Compose, AWS ECS, etc.
  - Location: `.github/workflows/deploy-staging.yml`
  - Lines: 50-70

- [ ] **Configure Health Checks**
  - Update health check URLs
  - Add custom smoke tests
  - Location: `.github/workflows/deploy-staging.yml`
  - Lines: 85-115

### CI Workflow
- [ ] **Review Test Configuration**
  - Verify DATABASE_URL is correct
  - Verify REDIS_URL is correct
  - Add any custom environment variables

- [ ] **Configure Coverage Thresholds** (optional)
  - Edit CI workflow to fail on low coverage
  - Add coverage requirements

### Security Workflow
- [ ] **Review Security Settings**
  - Adjust security scan schedule if needed
  - Configure exclusions in `.golangci.yml` if needed
  - Review gosec rules

### Release Workflow
- [ ] **Configure Release Settings**
  - Verify version tag pattern (v*)
  - Update release note template if needed
  - Configure Docker registry if not using GHCR

## Security Features

### Enable GitHub Security Features
- [ ] **Dependency Graph**
  - Go to: Settings > Code security and analysis
  - Enable: Dependency graph

- [ ] **Dependabot Alerts**
  - Enable: Dependabot alerts

- [ ] **Dependabot Security Updates**
  - Enable: Dependabot security updates

- [ ] **Secret Scanning**
  - Enable: Secret scanning (if available)

- [ ] **Code Scanning**
  - Enable: CodeQL analysis
  - Workflows already configured

### Create Dependabot Configuration
- [ ] **Create .github/dependabot.yml**
  ```yaml
  version: 2
  updates:
    - package-ecosystem: "gomod"
      directory: "/"
      schedule:
        interval: "weekly"
    - package-ecosystem: "npm"
      directory: "/web"
      schedule:
        interval: "weekly"
    - package-ecosystem: "github-actions"
      directory: "/"
      schedule:
        interval: "weekly"
  ```

## Testing

### Test CI Pipeline
- [ ] **Create Test Branch**
  ```bash
  git checkout -b test-ci-pipeline
  echo "# Test" >> README.md
  git add README.md
  git commit -m "test: verify CI pipeline"
  git push origin test-ci-pipeline
  ```

- [ ] **Create Pull Request**
  - Create PR to main
  - Verify all checks run
  - Check status checks appear

- [ ] **Verify Workflows**
  - [ ] Go Tests pass
  - [ ] Go Lint passes
  - [ ] Frontend Tests pass
  - [ ] Frontend Lint passes
  - [ ] Build completes

### Test Security Scanning
- [ ] **Manual Trigger**
  - Go to: Actions > Security Scanning
  - Click: "Run workflow"
  - Verify all scans complete

- [ ] **Check Security Tab**
  - Go to: Security > Code scanning
  - Verify alerts appear (if any)

### Test Release Process
- [ ] **Create Test Release**
  ```bash
  # Don't push to production - test locally first
  git tag v0.0.1-test
  # Review what would be released
  git show v0.0.1-test
  # Delete test tag
  git tag -d v0.0.1-test
  ```

- [ ] **Manual Release Trigger**
  - Go to: Actions > Release
  - Click: "Run workflow"
  - Enter: v0.0.1-test
  - Monitor build process

### Test Staging Deployment
- [ ] **Deploy to Staging**
  ```bash
  git checkout dev
  git push origin dev
  ```

- [ ] **Verify Deployment**
  - Check Actions > Deploy to Staging
  - Verify Docker image built
  - Check deployment succeeded
  - Verify health checks pass

## Documentation

### Update Project Documentation
- [ ] **Add CI/CD Section to README**
  - Already includes badges
  - Link to docs/CI-CD.md

- [ ] **Review Workflow Documentation**
  - [ ] .github/workflows/README.md
  - [ ] .github/workflows/QUICKSTART.md
  - [ ] .github/workflows/ARCHITECTURE.md
  - [ ] docs/CI-CD.md

- [ ] **Create CONTRIBUTING.md** (if not exists)
  - Add CI/CD workflow information
  - Link to CI-CD.md

## Monitoring and Maintenance

### Set Up Monitoring
- [ ] **GitHub Actions Usage**
  - Check: Settings > Billing > Actions
  - Review monthly usage

- [ ] **Enable Notifications**
  - Go to: Profile > Settings > Notifications
  - Configure: Actions notifications

- [ ] **Set Up Codecov** (optional)
  - Sign up at https://codecov.io
  - Add repository
  - Configure coverage requirements

### Create Maintenance Schedule
- [ ] **Weekly Tasks**
  - Review failed workflow runs
  - Check security alerts
  - Review Dependabot PRs

- [ ] **Monthly Tasks**
  - Update GitHub Actions versions
  - Review workflow performance
  - Audit secrets

- [ ] **Quarterly Tasks**
  - Review .golangci.yml configuration
  - Audit CI/CD costs
  - Update documentation

## Team Onboarding

### Share with Team
- [ ] **Distribute Documentation**
  - Share CI-CD-SETUP-SUMMARY.md
  - Share QUICKSTART.md
  - Share docs/CI-CD.md

- [ ] **Train Team Members**
  - CI/CD workflow overview
  - How to create releases
  - How to deploy to staging
  - How to read security reports

- [ ] **Create Runbooks**
  - Troubleshooting common issues
  - Rollback procedures
  - Incident response

## Compliance and Policies

### Review Policies
- [ ] **Code Review Policy**
  - Require code reviews before merge
  - Number of required approvers

- [ ] **Security Policy**
  - Create SECURITY.md if needed
  - Define vulnerability disclosure process

- [ ] **License Compliance**
  - Verify dependency licenses
  - Add license scanning if needed

## Production Readiness

### Pre-Production Checklist
- [ ] **All Tests Passing**
  - CI workflow successful
  - Security scans clean
  - No critical vulnerabilities

- [ ] **Documentation Complete**
  - All docs reviewed
  - Team trained
  - Runbooks created

- [ ] **Monitoring Configured**
  - Notifications enabled
  - Alerts configured
  - Dashboards created

- [ ] **Rollback Plan**
  - Documented rollback procedure
  - Tested rollback process
  - Team trained on rollback

### Launch Checklist
- [ ] **Final Review**
  - All workflows tested
  - All integrations verified
  - Team sign-off obtained

- [ ] **Enable Production Workflows**
  - Merge CI/CD changes to main
  - Tag first production release
  - Monitor first production deployment

- [ ] **Post-Launch**
  - Monitor for 24 hours
  - Review any issues
  - Document lessons learned

## Optimization

### Performance Optimization
- [ ] **Review Cache Efficiency**
  - Check cache hit rates
  - Optimize cache keys

- [ ] **Optimize Workflow Execution**
  - Review job dependencies
  - Maximize parallelization
  - Reduce redundant steps

- [ ] **Reduce CI Minutes**
  - Review workflow triggers
  - Optimize concurrency groups
  - Skip unnecessary jobs

### Cost Optimization
- [ ] **Review GitHub Actions Usage**
  - Check monthly minutes used
  - Optimize expensive workflows
  - Consider self-hosted runners if needed

- [ ] **Artifact Retention**
  - Review artifact sizes
  - Adjust retention periods
  - Clean up old artifacts

## Sign-Off

### Team Sign-Off
- [ ] **Development Team** - Approves CI/CD workflows
- [ ] **DevOps Team** - Approves infrastructure configuration
- [ ] **Security Team** - Approves security scanning setup
- [ ] **Management** - Approves rollout plan

### Final Verification
- [ ] All checklist items completed
- [ ] All workflows tested
- [ ] Documentation complete
- [ ] Team trained
- [ ] Ready for production

---

## Quick Links

- [Full Documentation](../../../docs/CI-CD.md)
- [Quick Start Guide](./QUICKSTART.md)
- [Architecture Overview](./ARCHITECTURE.md)
- [Workflows README](./README.md)

---

**Completion Status**: _____ / 100 items

**Completed By**: _______________ **Date**: ___________

**Reviewed By**: _______________ **Date**: ___________

**Approved By**: _______________ **Date**: ___________
