# Go Dependency Analyzer MCP Server

A simple MCP (Model Context Protocol) server that analyzes Go project dependencies.

## Features

- Scan dependencies in a Go project
- Check for outdated dependencies
- Generate simple dependency visualizations
- Support for both stdio and HTTP/SSE transport

## Usage

1. Build the server:
   ```
   go build -o go-deps-mcp
   ```

2. Run the server in stdio mode (default):
   ```
   ./go-deps-mcp
   ```

3. Run the server in HTTP mode:
   ```
   ./go-deps-mcp -http -port 8080
   ```
   This will start an HTTP server with the following endpoints:
   - SSE endpoint: `http://localhost:8080/mcp/sse`
   - Message endpoint: `http://localhost:8080/mcp/messages`

4. Connect to the server using an MCP client like Windsurf.

## HTTP Support

The server supports HTTP transport with the following features:

- Server-Sent Events (SSE) for real-time updates
- RESTful API for retrieving dependency information

To use the HTTP transport, run the server with the `-http` flag and specify the port using the `-port` flag.

## Tools

The server provides three main tools that analyze go.mod files:

1. `scan_dependencies` - Analyzes dependencies from go.mod content
   - Parameters: `goMod` (string) - Content of go.mod file

2. `check_outdated` - Checks for outdated dependencies from go.mod content
   - Parameters: `goMod` (string) - Content of go.mod file

3. `visualize_dependencies` - Creates a visualization from go.mod content
   - Parameters: `goMod` (string) - Content of go.mod file

## Docker Deployment

A Dockerfile is provided for easy containerization and deployment:

```bash
# Build the Docker image
docker build -t go-deps-mcp .

# Run the container
docker run -p 8080:8080 go-deps-mcp
```

The server will be accessible at:
- SSE endpoint: http://localhost:8080/mcp/sse
- Message endpoint: http://localhost:8080/mcp/messages