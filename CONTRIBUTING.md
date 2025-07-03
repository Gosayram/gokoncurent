# Contributing to GoKoncurent

Thank you for considering contributing to GoKoncurent! This document outlines the process for contributing to this project.

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

Before submitting a bug report, please check if the issue has already been reported by searching our [Issues](https://github.com/Gosayram/gokoncurent/issues).

When creating a bug report, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Go version (`go version`)
- Operating system and version
- Any relevant code snippets or error messages

### Suggesting Enhancements

Enhancement suggestions are welcome! Please:

1. Check if the enhancement has already been suggested
2. Create a new issue with the label "enhancement"
3. Provide a clear description of the enhancement
4. Explain why this enhancement would be useful

### Pull Requests

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Make your changes
4. Ensure all tests pass
5. Add or update tests as necessary
6. Update documentation if needed
7. Submit a pull request

#### Pull Request Process

1. Update the README.md with details of changes if applicable
2. Update the CHANGELOG.md with your changes
3. Ensure your code follows the project's coding standards
4. Make sure all tests pass
5. Request a review from maintainers

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git

### Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/your-username/gokoncurent.git
   cd gokoncurent
   ```

2. Install development tools:
   ```bash
   go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint
   ```

3. Run tests to ensure everything is working:
   ```bash
   go test ./...
   ```

### Development Workflow

1. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Run tests:
   ```bash
   go test ./...
   ```

4. Run linter:
   ```bash
   go tool golangci-lint run
   ```

5. Format code:
   ```bash
   go fmt ./...
   ```

6. Run static analysis:
   ```bash
   go vet ./...
   ```

7. Commit your changes:
   ```bash
   git commit -m "feat: add new feature"
   ```

8. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

9. Create a pull request

## Coding Standards

### Code Style

- Follow the official Go style guide
- Use `go fmt` to format your code
- Use meaningful variable and function names
- Write clear and concise comments
- Keep functions small and focused

### Testing

- Write unit tests for all new functionality
- Aim for >90% test coverage
- Use table-driven tests when appropriate
- Mock external dependencies
- Test both success and failure scenarios

### Documentation

- Document all exported functions, types, and variables
- Use godoc format for documentation
- Include examples in documentation when helpful
- Update README.md for significant changes

### Error Handling

- Handle errors explicitly
- Use error wrapping with `fmt.Errorf("message: %w", err)`
- Define custom error types for specific conditions
- Document error conditions in function comments

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
- `feat: add Arc[T] atomic reference counting`
- `fix: resolve race condition in ArcMutex`
- `docs: update README with OnceCell examples`

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

## Release Process

1. Update version in `.release-version`
2. Update version in `gokoncurent.go`
3. Update CHANGELOG.md
4. Create a Git tag
5. Push tag to trigger release workflow

## Questions?

If you have questions about contributing, please:

1. Check the [FAQ](docs/FAQ.md)
2. Search existing [Issues](https://github.com/Gosayram/gokoncurent/issues)
3. Create a new issue with the "question" label
4. Join our [Discussions](https://github.com/Gosayram/gokoncurent/discussions)

## License

By contributing to GoKoncurent, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- The project's README.md
- Release notes
- The AUTHORS file

Thank you for contributing to GoKoncurent! 🎉 