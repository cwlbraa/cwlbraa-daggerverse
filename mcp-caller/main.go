// A generated module for McpCaller functions

package main

import (
	"context"
	"dagger/mcp-caller/internal/dagger"
)

type McpCaller struct {
	Results []string
}

type SearchResults struct {
	Results []string
}

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

func (m *McpCaller) MCPK8s(kubeConfig *dagger.Secret) *dagger.Container {
	return m.Playwright().
		WithEntrypoint([]string{"npx", "-y", "kubernetes-mcp-server@latest"}).
		WithMountedSecret("/root/.kube/config", kubeConfig)
}

// // this server does not currently seem to respond to ListTools requests?
// func (m *McpCaller) MCPK8s(kubeConfig *dagger.Secret) *dagger.Container {
// 	return dag.Container().From("ghcr.io/alexei-led/aws-mcp-server:latest").
// 		WithMountedSecret("/home/app-user/.kube/config", kubeConfig)
// }

// this server does not currently seem to respond to ListTools requests?
func (m *McpCaller) MCPAWS() *dagger.Container {
	return dag.Container().From("ghcr.io/alexei-led/aws-mcp-server:latest")
}

/*
	with-dev dagger -M -c 'llm |\
	  with-mcp $(container |\
	             from "mcr.microsoft.com/playwright:v1.51.1-noble" |\
	             with-entrypoint -- npx -y g-search-mcp) |\
	  with-prompt "search google for hello world. use the us locale. show me the result urls." |\
	  last-reply'
*/
func (m *McpCaller) GSearchExample(ctx context.Context) (string, error) {
	return dag.LLM().
		WithMCP(m.MCPGSearch()).
		WithPrompt("search google using the query 'hello world'. use the us locale. and show me the result urls.").
		LastReply(ctx)
}

func (m *McpCaller) K8sExample(ctx context.Context, kubeConfig *dagger.Secret) (string, error) {
	return dag.LLM().
		WithMCP(m.MCPK8s(kubeConfig)).
		WithPrompt("list tools").
		LastReply(ctx)
}

func (m *McpCaller) AWSExample(ctx context.Context, awsDir *dagger.Directory) (string, error) {
	return dag.LLM().
		WithMCP(m.MCPAWS().
			WithMountedDirectory("$HOME/.aws", awsDir, dagger.ContainerWithMountedDirectoryOpts{Expand: true}).
			WithEnvVariable("AWS_PROFILE", "dagger-ci")).
		WithPrompt("list tools").
		LastReply(ctx)
}
