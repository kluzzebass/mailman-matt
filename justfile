# Set the default Docker image name
image_name := "kluzz/mailman-matt-go"

# List available commands
default:
	@just --list

# Run the application locally
run:
	go run .

# Build the application binary
build:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o build/matt .

# Build Docker image
docker-build:
	docker build --platform linux/amd64 -t {{image_name}} .

# Build both binary and Docker image
build-all: build docker-build

# Push Docker image to Docker Hub
docker-push:
	docker push {{image_name}}

# Tag a new version (usage: just tag-version v1.0.0)
tag-version version:
	@echo "Tagging version {{version}}"
	git tag -a {{version}} -m "Release {{version}}"
	git push origin {{version}}
	docker tag {{image_name}} {{image_name}}:{{version}}
	docker tag {{image_name}} {{image_name}}:latest
	docker push {{image_name}}:{{version}}
	docker push {{image_name}}:latest

# Run the application in Docker
docker-run:
	docker run -p 3000:3000 -it --name mailman-matt --rm {{image_name}}

# Run the application in Docker (detached mode)
docker-run-detached:
	docker run -p 3000:3000 -d --name mailman-matt {{image_name}}

# Stop and remove the Docker container
docker-rm:
	docker rm -f mailman-matt || true

# Clean up build artifacts
clean:
	rm -rf build/
