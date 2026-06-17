# CI/CD Setup

This repository uses GitHub Actions for continuous integration and continuous deployment.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **CLI Tests**: Runs go vet, go fmt, and unit tests for the CLI
- **Router Tests**: Runs go vet, go fmt, and unit tests for the Router
- **Integration Tests**: Runs integration tests for both CLI and Router
- **Build**: Builds both CLI and Router after tests pass

### 2. Integration Tests Workflow (`.github/workflows/integration-tests.yml`)

Runs on every push and pull request to `main` and `develop` branches, and can be triggered manually.

**Jobs:**
- **CLI Integration Tests**: Runs CLI integration tests with coverage
- **Router Integration Tests**: Runs Router integration tests with coverage
- **End-to-End Tests**: Tests CLI registration with Router service

## Running Tests Locally

### CLI Tests

```bash
cd apps/cli
go test -v -race ./...
go test -v -race ./internal/integration/...
```

### Router Tests

```bash
cd apps/router
go test -v -race ./...
go test -v -race ./layers/metrics/...
```

### End-to-End Tests

```bash
# Start Router in background
cd apps/router/cmd/routerd && ./routerd -port 1807 &

# Wait for Router to start
sleep 5

# Test CLI registration
cd apps/cli
TI_ROUTER_AGENT_URL=http://localhost:1807 TI_DISABLE_ROUTER_REGISTRATION=true ./ti router status

# Verify metrics
curl http://localhost:1807/metrics/prometheus | grep cli
```

## Coverage

Coverage reports are uploaded to Codecov for both CLI and Router tests.

- **CLI Coverage**: Flags `cli` and `cli-integration`
- **Router Coverage**: Flags `router` and `router-integration`

## Status Badges

Add these badges to your README.md:

```markdown
[![CI](https://github.com/your-org/ti/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/ti/actions/workflows/ci.yml)
[![Integration Tests](https://github.com/your-org/ti/actions/workflows/integration-tests.yml/badge.svg)](https://github.com/your-org/ti/actions/workflows/integration-tests.yml)
[![codecov](https://codecov.io/gh/your-org/ti/branch/main/graph/badge.svg)](https://codecov.io/gh/your-org/ti)
```

## Manual Workflow Trigger

You can manually trigger the integration tests workflow:

1. Go to the Actions tab in GitHub
2. Select "Integration Tests" workflow
3. Click "Run workflow"
4. Select the branch and click "Run workflow"

## Troubleshooting

### Tests Failing Locally but Passing in CI

Check for:
- Environment differences (Go version, OS)
- Missing dependencies
- File path issues (Windows vs Linux)

### Integration Tests Failing

Check for:
- Router service not running
- Port conflicts
- Network connectivity issues

### Coverage Not Uploading

Check for:
- Codecov token configured in repository settings
- Coverage file path is correct
- Coverage flags match

## Adding New Tests

### Unit Tests

Add to the appropriate package:
- CLI: `apps/cli/internal/<package>/<package>_test.go`
- Router: `apps/router/<package>/<package>_test.go`

### Integration Tests

Add to:
- CLI: `apps/cli/internal/integration/<feature>_test.go`
- Router: `apps/router/layers/metrics/<feature>_test.go`

## Best Practices

1. **Write tests for new features**: Add unit tests for all new code
2. **Keep tests fast**: Unit tests should run in seconds
3. **Use table-driven tests**: For testing multiple scenarios
4. **Mock external dependencies**: Use mocks for external services
5. **Test error cases**: Don't just test the happy path
6. **Keep coverage high**: Aim for >80% coverage
7. **Run tests locally before pushing**: Catch issues early

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing](https://golang.org/pkg/testing/)
- [Codecov Documentation](https://docs.codecov.com/)
- [Integration Test Architecture](apps/cli/internal/integration/architecture.md)
