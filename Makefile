.PHONY: build run docker-build docker-run clean test k8s-deploy k8s-delete

# Application name
APP_NAME=telegpt
DOCKER_IMAGE=telegpt:latest

# Go build flags
GO_BUILD_FLAGS=-ldflags="-s -w"

# Build the application
build:
	go build $(GO_BUILD_FLAGS) -o $(APP_NAME) ./cmd/bot

# Run the application
run:
	go run ./cmd/bot

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run:
	docker run --rm -it \
		--env-file .env \
		$(DOCKER_IMAGE)

# Clean build artifacts
clean:
	rm -f $(APP_NAME)
	go clean

# Run tests
test:
	go test -v ./...

# Deploy to Kubernetes (requires kubectl and valid kubeconfig)
k8s-deploy:
	kubectl apply -f kubernetes/deployment.yaml

# Delete Kubernetes deployment
k8s-delete:
	kubectl delete -f kubernetes/deployment.yaml

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linting
lint:
	go vet ./...

# Create development config files from examples
init-dev:
	test -f config.yaml || cp config.yaml.example config.yaml
	test -f .env || cp .env.example .env

# Display help information
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the application"
	@echo "  run          Run the application"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run Docker container"
	@echo "  clean        Clean build artifacts"
	@echo "  test         Run tests"
	@echo "  k8s-deploy   Deploy to Kubernetes"
	@echo "  k8s-delete   Delete Kubernetes deployment"
	@echo "  deps         Install dependencies"
	@echo "  fmt          Format code"
	@echo "  lint         Run linting"
	@echo "  init-dev     Create dev config files from examples"
