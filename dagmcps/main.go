package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	_ "embed"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed dagger_manual.md
var daggershellDescription string

const commandDescription string = `the script to run in the dagger engine. 
syntax is similar to bash, but uses dagger primitives instead of executables.
use .stdlib to see what's available at the root.
use .doc to inspect documentation.

for example, ".stdlib | .doc" gives something like this:

<result>
COMMANDS
  cache-volume   Constructs a cache volume for a given cache key.
  container      Creates a scratch container.

                 Optional platform argument initializes new containers to execute and publish as that platform. Platform defaults to that of the builder's
                 host.
  directory      Creates an empty directory.
  engine         The Dagger engine container configuration and state
  git            Queries a Git repository.
  host           Queries the host environment.
  http           Returns a file containing an http remote url content.
  set-secret     Sets a secret given a user defined name to its plaintext and returns the secret.

                 The plaintext value is limited to a size of 128000 bytes.
  version        Get the current Dagger Engine version.

Use ".stdlib | .doc <command>" for more information on a command.
</result>
`
const cwdDescription = `
the working directory to run the shell in.
use this instead of trying to cd if you're asked to work on code local to the users' machine, like via the filesystem mcp.
provide an absolute path, not a relative path.
`

func main() {
	s := server.NewMCPServer(
		"Dagger",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	s.AddTool(mcp.NewTool("dagger_shell",
		mcp.WithDescription(daggershellDescription),
		mcp.WithString("cmd",
			mcp.Required(),
			mcp.Title("Command"),
			mcp.Description(commandDescription),
		),
		mcp.WithString("cwd",
			mcp.Title("Current Working Directory"),
			mcp.Description(cwdDescription),
		),
	), shellHandler)

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
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "NO_COLOR=1")

	cwdStr, ok := request.Params.Arguments["cwd"].(string)
	if ok {
		cmd.Dir = cwdStr
	}

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("command failed: %v\nOutput: %s", err, output)), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}
