// A generated module for McpCaller functions

package main

import (
	"context"
	"dagger/mcp-caller/internal/dagger"
)

type McpCaller struct{}

func (m *McpCaller) WolfiNode() *dagger.Container {
	return dag.Wolfi().Container(dagger.WolfiContainerOpts{
		Packages: []string{"nodejs", "npm"},
	})
}

func (m *McpCaller) Playwright() *dagger.Container {
	return dag.Container().From("mcr.microsoft.com/playwright:v1.51.1-noble")
}

func (m *McpCaller) MCPGSearch() *dagger.Container {
	return m.Playwright().
		WithEntrypoint([]string{"npx", "-y", "g-search-mcp"})
}

func (m *McpCaller) MCPAWS() *dagger.Container {
	return dag.Container().From("ghcr.io/alexei-led/aws-mcp-server:latest")
}

func (m *McpCaller) LlmWithMCP() *dagger.LLM {
	return dag.LLM().
		WithMCP(m.MCPGSearch()).
		WithPrompt("list your available tools")
}

// with-dev dagger -c "llm-with-mcp | with-prompt 'search google for hello world. use the us locale. and show me the result urls.'| last-reply"
func (m *McpCaller) Example(ctx context.Context) *dagger.LLM {
	return m.LlmWithMCP().WithPrompt(
		"search google for hello world. use the us locale. and show me the result urls.",
	)
}
