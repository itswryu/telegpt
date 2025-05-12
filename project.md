# TeleGPT Development Guidelines

## Project Overview

TeleGPT is a Go-based Telegram bot that uses OpenAI's GPT-4.1-nano model to respond to user messages. The bot is designed to be secure, scalable, and containerized.

## Development Guidelines

### Code Structure

- **cmd/bot**: Main application entry point
- **pkg/config**: Configuration management
- **pkg/telegram**: Telegram bot implementation
- **pkg/openai**: OpenAI API client
- **kubernetes/**: Kubernetes deployment files

### Coding Standards

- Follow Go's standard code formatting (use `gofmt` or `go fmt`)
- Use meaningful variable and function names
- Add comments for non-obvious code sections
- Follow the principle of least privilege
- Write tests for critical functionality

### Security Practices

- Never commit sensitive information to version control
- Use environment variables or secure config files for sensitive data
- Validate all user inputs
- Follow the principle of least privilege in Docker containers
- Use Kubernetes secrets for sensitive information

### Configuration Management

- Use environment variables for deployment-specific configuration
- Use config.yaml for application configuration
- Always provide example files with clear documentation
- Validate configuration at startup

## Deployment Process

### Docker Build

```bash
make docker-build
```

### Kubernetes Deployment

1. Create the necessary secrets
2. Apply the Kubernetes deployment files
```bash
make k8s-deploy
```

## Task Management

- [ ] Implement unit tests for core functionality
- [x] Add monitoring and logging
- [x] Create a CI/CD pipeline for automated testing and deployment
- [x] Implement user authentication with Chat IDs
- [x] Implement OpenAI integration
- [x] Implement Telegram bot integration
- [x] Create deployment configuration files
- [x] Add conversation history support
- [x] Implement graceful shutdown
- [x] Create system service files
- [x] Add Docker and Kubernetes deployment options

## Performance Considerations

- Use goroutines for handling messages to improve concurrency
- Implement connection pooling for API calls
- Consider caching mechanisms for frequent requests
- Monitor memory usage and response times

## Versioning

- Use semantic versioning for releases
- Tag all releases in git
- Document changes in a CHANGELOG.md file

## Contributing

Before submitting code changes:
1. Run all tests
2. Format code with `go fmt`
3. Run `go vet` to check for potential issues
4. Verify that the Docker build succeeds
