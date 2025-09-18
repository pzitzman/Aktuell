# Contributing to Aktuell

Thank you for your interest in contributing to **Aktuell**! We welcome contributions of all kinds, including bug reports, feature requests, documentation improvements, and code contributions.

## 🤝 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Use our issue templates** when available
3. **Provide clear, detailed information** including:
   - Operating system and version
   - Go version
   - MongoDB version
   - Steps to reproduce
   - Expected vs. actual behavior
   - Relevant logs or screenshots

### Feature Requests

We love hearing about new ideas! Please:

1. **Check existing feature requests** first
2. **Describe the problem** you're trying to solve
3. **Explain your proposed solution** with examples
4. **Consider the impact** on existing users
5. **Be open to discussion** about alternative approaches

### Code Contributions

#### Prerequisites

- Go 1.21 or higher
- MongoDB 4.0+ (for change streams support)
- Node.js 16+ (for React client)
- Git

#### Development Setup

1. **Fork the repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/pzitzman/aktuell.git
   cd aktuell
   ```

2. **Set up development environment**
   ```bash
   # Install Go dependencies
   go mod download
   
   # Install React dependencies
   cd client
   npm install
   cd ..
   
   # Start MongoDB for testing
   docker run -d --name aktuell-mongo -p 27017:27017 mongo:latest
   ```

3. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

#### Development Workflow

1. **Make your changes**
   - Write clean, readable code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation as needed

2. **Run tests locally**
   ```bash
   # Run Go tests
   make test
   
   # Run integration tests
   make test-integration
   
   # Run React tests
   cd client && npm test
   ```

3. **Check code quality**
   ```bash
   # Format code
   make format
   
   # Run linters
   make lint
   
   # Check security
   make security-check
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing new feature"
   ```

## 📋 Code Standards

### Go Code Style

- **Follow standard Go conventions**
- **Use `gofmt`** for formatting
- **Write meaningful variable names**
- **Add comments for exported functions**
- **Handle errors appropriately**
- **Use context for cancellation**

Example:
```go
// ProcessChangeStream processes MongoDB change stream events
// and forwards them to connected WebSocket clients.
func (s *Server) ProcessChangeStream(ctx context.Context, stream *mongo.ChangeStream) error {
    for stream.Next(ctx) {
        var change models.ChangeEvent
        if err := stream.Decode(&change); err != nil {
            return fmt.Errorf("failed to decode change event: %w", err)
        }
        
        s.broadcastChange(change)
    }
    
    return stream.Err()
}
```

### React/TypeScript Style

- **Use TypeScript strictly**
- **Follow React best practices**
- **Use functional components with hooks**
- **Implement proper error boundaries**
- **Write reusable components**

Example:
```typescript
interface ChangeEventProps {
  event: ChangeEvent;
  onRetry?: () => void;
}

export const ChangeEventDisplay: React.FC<ChangeEventProps> = ({
  event,
  onRetry
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  
  return (
    <div className="change-event">
      <div className="event-header">
        <span className={`operation-type ${event.operationType}`}>
          {event.operationType}
        </span>
        <span className="timestamp">
          {new Date(event.timestamp).toLocaleString()}
        </span>
      </div>
      {/* More component code */}
    </div>
  );
};
```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(server): add support for multiple database monitoring
fix(client): resolve WebSocket reconnection issues
docs: update installation instructions
test: add integration tests for change stream processing
```

### Testing Requirements

#### Unit Tests
- **Test public functions** and methods
- **Use table-driven tests** where appropriate
- **Mock external dependencies**
- **Aim for >80% coverage**

```go
func TestChangeStreamProcessor(t *testing.T) {
    tests := []struct {
        name           string
        input          bson.Raw
        expectedOutput models.ChangeEvent
        expectedError  error
    }{
        {
            name:  "valid insert operation",
            input: bson.Raw(`{"operationType": "insert", "fullDocument": {...}}`),
            expectedOutput: models.ChangeEvent{
                OperationType: "insert",
                // ... more fields
            },
        },
        // More test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Tests
- **Test complete workflows**
- **Use real MongoDB instances**
- **Test WebSocket connections**
- **Verify end-to-end functionality**

## 🚀 Pull Request Process

1. **Create detailed PR description**
   - Describe what changes you made
   - Explain why you made them
   - Link to relevant issues
   - Include screenshots for UI changes

2. **Ensure all checks pass**
   - All tests pass ✅
   - Code coverage maintained ✅
   - Linting passes ✅
   - Security scans pass ✅

3. **Request review**
   - Tag relevant maintainers
   - Be responsive to feedback
   - Make requested changes promptly

4. **Update documentation**
   - Update README if needed
   - Add/update code comments
   - Update API documentation

## 🔒 Security Guidelines

### Reporting Security Issues

**Do NOT create public issues for security vulnerabilities.**

Instead, please report security issues using one of these methods:

1. **GitHub Security Advisories** (Recommended): Go to the [Security](https://github.com/pzitzman/aktuell/security/advisories) tab and click "Report a vulnerability"
2. **GitHub Issues**: Create a private security advisory through GitHub's interface

We will respond within 48 hours and work with you to resolve the issue quickly.

### Security Best Practices

- **Never commit secrets** (API keys, passwords, tokens)
- **Validate all inputs** from external sources
- **Use parameterized queries** for database operations
- **Implement proper authentication** and authorization
- **Keep dependencies updated** for security patches

## 📚 Documentation

### Code Documentation

- **Document all exported functions** with clear descriptions
- **Include usage examples** where helpful
- **Document complex algorithms** or business logic
- **Keep README.md updated** with new features

### API Documentation

- **Document all endpoints** with request/response examples
- **Describe error conditions** and status codes
- **Include authentication requirements**
- **Provide cURL examples** for testing

## 🏆 Recognition

Contributors will be recognized in:

- **GitHub contributors list**
- **Release notes** for significant contributions
- **README.md acknowledgments**
- **Project documentation**

## 📞 Getting Help

- **GitHub Discussions**: For general questions and discussions
- **GitHub Issues**: For bug reports and feature requests
- **Discord/Slack**: [Link to community chat if available]
- **Email**: [maintainer-email] for private concerns

## 📄 License

By contributing to Aktuell, you agree that your contributions will be licensed under the same [MIT License](LICENSE) that covers the project.

---

Thank you for contributing to **Aktuell**! Your help makes this project better for everyone. 🙏

## Quick Checklist

Before submitting your contribution:

- [ ] Code follows project style guidelines
- [ ] All tests pass locally
- [ ] New tests added for new functionality
- [ ] Documentation updated as needed
- [ ] Commit messages follow conventional format
- [ ] PR description is clear and detailed
- [ ] No secrets or sensitive data committed
- [ ] Changes are backwards compatible (or breaking changes documented)

**Happy coding!** 🚀