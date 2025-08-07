.PHONY: build run clean container-build container-run container-clean deploy undeploy

# Build the rosa-mcp-server binary
build:
	go build -o rosa-mcp-server ./cmd/rosa-mcp-server

# Build and run the server with stdio transport
run: build
	./rosa-mcp-server --transport=stdio

# Clean build artifacts
clean:
	rm -f rosa-mcp-server

# Build container image
container-build:
	podman build -t rosa-mcp-server:latest .

# Run container with SSE transport and verbose logging
container-run: container-build
	podman run --rm -p 8080:8080 \
		rosa-mcp-server:latest ./rosa-mcp-server --transport=sse --port=8080 -v=3

# Clean container image
container-clean:
	podman rmi rosa-mcp-server:latest

# Deploy to OpenShift using template
deploy:
	oc process -f openshift/template.yaml | oc apply -f -

# Remove deployed resources from OpenShift
undeploy:
	oc process -f openshift/template.yaml | oc delete -f -
