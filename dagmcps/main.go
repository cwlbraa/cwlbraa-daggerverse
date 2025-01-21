package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"Dagger",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	tool := mcp.NewTool("daggershell",
		mcp.WithDescription("execute dagger shell scripts"),
		mcp.WithString("cmd",
			mcp.Required(),
			mcp.Description("the script to run. syntax is similar to bash, but uses dagger primitives instead of executables"),
		),
	)

	// Add tool handler
	s.AddTool(tool, shellHandler)

	fmt.Fprint(os.Stderr, "Starting Dagger MCP server on stdio transport")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func shellHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmdStr, ok := request.Params.Arguments["cmd"].(string)
	if !ok {
		return mcp.NewToolResultError("cmd must be a string"), nil
	}

	// Create the dagger shell command
	cmd := exec.CommandContext(ctx, "dagger", "shell", "-c", cmdStr)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("command failed: %v\nOutput: %s", err, output)), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}
