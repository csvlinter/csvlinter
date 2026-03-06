# Contributing to csvlinter

Thank you for considering contributing to csvlinter.

## How to contribute

1. **Fork** the repository on GitHub.
2. **Create a branch** from `main` for your change (`git checkout -b feat/your-feature` or `fix/your-fix`).
3. **Make your changes** and add or update tests as needed.
4. **Run tests and lint** locally:
   - `go test ./...`
   - `go build .`
   - `golangci-lint run` (if you have golangci-lint installed)
5. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/) so that [semantic-release](https://github.com/semantic-release/semantic-release) can version and changelog correctly:
   - `feat: add something`
   - `fix: resolve bug in X`
   - `docs: update README`
   - `test: add tests for Y`
6. **Push** your branch and open a **Pull Request** against `main`.

## Code and design

- Prefer small, focused PRs.
- The public API lives under `pkg/csvlinter`; avoid breaking changes there without a good reason and a deprecation path.
- For CLI behavior changes, update the README and ensure the PR workflow (tests + manual CLI checks) still passes.

## Questions or ideas

Open a [GitHub Discussion](https://github.com/csvlinter/csvlinter/discussions) or an [Issue](https://github.com/csvlinter/csvlinter/issues) to discuss ideas or ask questions before large changes.
