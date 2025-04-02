// A generated module for McpCaller functions

package main

import (
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
		// WithMCP(m.MCPAWS()).
		WithPrompt("list your available tools")
}
