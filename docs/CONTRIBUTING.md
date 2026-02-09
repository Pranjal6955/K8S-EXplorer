# Contributing to K8S Graph Explorer

Thank you for your interest in contributing to K8S Graph Explorer! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker to report bugs
- Search existing issues before creating a new one
- Include detailed information:
  - Environment (OS, Go version, Node.js version)
  - Steps to reproduce
  - Expected vs actual behavior
  - Error messages and logs

### Suggesting Features

- Open an issue with the "enhancement" label
- Describe the feature and its use case
- Explain why it would be valuable

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Add tests if applicable
5. Ensure all tests pass
6. Commit with clear messages
7. Push to your fork
8. Open a pull request

## Development Setup

See [DEVELOPMENT.md](./DEVELOPMENT.md) for detailed setup instructions.

## Coding Standards

### Go (Backend)

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Run `gofmt` before committing
- Run `golangci-lint` for linting
- Write tests for new functionality
- Document exported functions and types

```go
// Good example
// SyncResources synchronizes Kubernetes resources to Neo4j.
// It returns the number of synced resources and any errors encountered.
func SyncResources(ctx context.Context, namespace string) (int, error) {
    // ...
}
```

### TypeScript/React (Dashboard)

- Use TypeScript strict mode
- Follow React best practices (hooks, functional components)
- Use meaningful component and variable names
- Add prop types for components

```typescript
// Good example
interface ResourceCardProps {
  resource: K8sResource;
  onSelect: (uid: string) => void;
}

export function ResourceCard({ resource, onSelect }: ResourceCardProps) {
  // ...
}
```

### GraphQL

- Use descriptive field names
- Add documentation comments
- Follow schema design best practices

```graphql
"""
Represents a Kubernetes namespace with its resources
"""
type Namespace {
  """Unique identifier"""
  uid: ID!
  """Name of the namespace"""
  name: String!
  """Resources in this namespace"""
  resources: [K8sResource!]!
}
```

## Commit Messages

Use clear, descriptive commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

Example:
```
feat(backend): add deployment sync to Neo4j

- Implement deployment listing from K8s API
- Create deployment nodes in Neo4j
- Add owner relationships to ReplicaSets

Closes #42
```

## Testing

### Backend

```bash
cd backend
make test
make test-coverage
```

### Dashboard

```bash
cd dashboard
npm test
npm run test:coverage
```

## Documentation

- Update documentation when changing functionality
- Add JSDoc/GoDoc comments for public APIs
- Keep README and docs/ in sync

## Release Process

1. Update version numbers
2. Update CHANGELOG.md
3. Create a release branch
4. Run full test suite
5. Build and push Docker images
6. Create GitHub release with notes
7. Merge to main

## Getting Help

- Open an issue for questions
- Join our Discord server (if available)
- Review existing documentation

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- CONTRIBUTORS.md file
- Release notes

Thank you for contributing! 🎉
