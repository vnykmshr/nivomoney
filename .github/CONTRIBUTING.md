# Contributing to Nivo

Nivo is a portfolio project, but contributions and feedback are welcome.

## Getting Started

1. Fork the repository
2. Clone your fork and set up the development environment:

```bash
git clone https://github.com/<your-username>/nivomoney.git
cd nivomoney
cp .env.example .env
make dev
make test
```

3. Create a feature branch from `main`

## Development

- **Prerequisites:** Go 1.24+, Docker, Node.js 18+
- **Full guide:** [docs.nivomoney.com/development](https://docs.nivomoney.com/development)

### Quick reference

```bash
make dev          # Start infrastructure (postgres, redis, etc.)
make test         # Run all Go tests with race detection
make lint         # Run golangci-lint
make fmt          # Format Go code
```

Frontend (separate terminal):
```bash
cd frontend/user-app && npm install && npm run dev
```

## Code Standards

- **Commits:** [Conventional commits](https://www.conventionalcommits.org/) — `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- **Go:** Standard library style, `gofmt`, passes `golangci-lint`
- **Tests:** Table-driven tests, happy path + key edge cases
- **Errors:** Fail loudly, no silent swallowing

## Pull Requests

1. One concern per PR
2. All tests pass (`make test`)
3. Linter passes (`make lint`)
4. Clear description of what and why

## Reporting Issues

Open a [GitHub issue](https://github.com/vnykmshr/nivomoney/issues) with:
- What you expected
- What happened
- Steps to reproduce

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](../LICENSE).
