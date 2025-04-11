package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Parse command line flags
	httpMode := flag.Bool("http", false, "Run in HTTP mode instead of stdio")
	port := flag.String("port", "8080", "Port to use in HTTP mode")
	flag.Parse()

	// Create a new MCP server
	s := server.NewMCPServer(
		"Go Dependency Analyzer",
		"1.0.0",
		server.WithResourceCapabilities(true, false), // Support static resources, no templates
		server.WithLogging(),
	)

	// Add scan dependencies tool
	scanTool := mcp.NewTool("scan_dependencies",
		mcp.WithDescription("Scan Go dependencies from go.mod content"),
		mcp.WithString("goMod",
			mcp.Required(),
			mcp.Description("Content of go.mod file"),
		),
	)

	s.AddTool(scanTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		goMod := request.Params.Arguments["goMod"].(string)

		// Call the analyzer function
		deps, err := scanDependencies(goMod)
		if err != nil {
			return nil, err
		}

		// Format the result as a table
		result := "## Dependencies\n\n"
		result += "| Module | Version | Direct |\n"
		result += "|--------|---------|--------|\n"

		for _, dep := range deps {
			direct := "No"
			if dep.Direct {
				direct = "Yes"
			}
			result += fmt.Sprintf("| %s | %s | %s |\n", dep.Path, dep.Version, direct)
		}

		return mcp.NewToolResultText(result), nil
	})

	// Add check outdated tool
	outdatedTool := mcp.NewTool("check_outdated",
		mcp.WithDescription("Check for outdated dependencies from go.mod content"),
		mcp.WithString("goMod",
			mcp.Required(),
			mcp.Description("Content of go.mod file"),
		),
	)

	s.AddTool(outdatedTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		goMod := request.Params.Arguments["goMod"].(string)

		// Call the analyzer function
		outdated, err := checkOutdatedDependencies(goMod)
		if err != nil {
			return nil, err
		}

		// Format the result as a table
		result := "## Outdated Dependencies\n\n"
		result += "| Module | Current Version | Latest Version | Update Needed |\n"
		result += "|--------|----------------|---------------|---------------|\n"

		for _, dep := range outdated {
			updateNeeded := "No"
			if dep.UpdateNeeded {
				updateNeeded = "Yes"
			}
			result += fmt.Sprintf("| %s | %s | %s | %s |\n",
				dep.Path, dep.CurrentVer, dep.LatestVer, updateNeeded)
		}

		return mcp.NewToolResultText(result), nil
	})

	// Add visualize dependencies tool
	vizTool := mcp.NewTool("visualize_dependencies",
		mcp.WithDescription("Generate a simple visualization of dependencies from go.mod content"),
		mcp.WithString("goMod",
			mcp.Required(),
			mcp.Description("Content of go.mod file"),
		),
	)

	s.AddTool(vizTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		goMod := request.Params.Arguments["goMod"].(string)

		// Call the analyzer function
		viz, err := visualizeDependencies(goMod)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(viz.Graph), nil
	})

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start the server based on mode
	if *httpMode {
		// Create SSE server
		sseServer := server.NewSSEServer(s,
			server.WithBasePath("/mcp"),
			server.WithSSEEndpoint("/sse"),
			server.WithMessageEndpoint("/messages"),
		)

		// Start HTTP server in a goroutine
		go func() {
			addr := fmt.Sprintf(":%s", *port)
			log.Printf("Starting HTTP server on %s", addr)
			log.Printf("SSE endpoint: http://localhost%s/mcp/sse", addr)
			if err := sseServer.Start(addr); err != nil {
				log.Fatalf("HTTP server error: %v", err)
			}
		}()

		// Wait for termination signal
		<-sigCh
		log.Println("Shutting down HTTP server...")
		if err := sseServer.Shutdown(context.Background()); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	} else {
		// Start stdio server
		log.Println("Starting stdio server")
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
