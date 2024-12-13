// A module for benchmarking the dagger engine
//
// This module contains a series of functions intended to stress the engine in various ways.
//
// Many are adapted from Dagger's core integration tests.

package main

import (
	"context"
	"dagger/cwlbraa-benchmarks/internal/dagger"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"golang.org/x/sync/errgroup"
)

type CwlbraaBenchmarks struct{}

const bigCat = "https://w.wallhaven.cc/full/jx/wallhaven-jxzzmp.jpg"
const curlCat = "for i in {1..20}; do curl -s %s -o \"cat$i.jpg\" && sync && sleep 0.1; done"

func (m *CwlbraaBenchmarks) IoTest(ctx context.Context) error {
	_, err := dag.Container().
		From("ubuntu:24.10").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "curl", "coreutils"}).
		WithEnvVariable("cache_bust", uuid.NewString()).
		WithNewFile("download.sh", fmt.Sprintf(curlCat, bigCat), dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithExec([]string{"bash", "-c", "./download.sh"}).
		WithExec([]string{"ls", "-alh", "cat1.jpg"}).
		Stdout(ctx)
	return err
}

// Returns a container that echoes whatever string argument is provided
func (m *CwlbraaBenchmarks) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

func (m *CwlbraaBenchmarks) BenchmarkEcho(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < 256; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			guid := uuid.New().String()
			echoMessage := fmt.Sprintf("Echo call #%d - GUID: %s", i+1, guid)
			_, err := m.ContainerEcho(echoMessage).Stdout(ctx)
			return err
		})
	}

	return g.Wait()
}

// Helper function to generate Go module source code
func getModMainSrc(name string, depNames []string) string {
	mainSrc := fmt.Sprintf(`package main

import "context"

type %s struct {}

func (m *%s) Fn(ctx context.Context) (string, error) {
	s := "%s"
	var depS string
	_ = depS
	var err error
	_ = err
`, strcase.ToCamel(name), strcase.ToCamel(name), name)

	for _, depName := range depNames {
		mainSrc += fmt.Sprintf(`
	depS, err = dag.%s().Fn(ctx)
	if err != nil {
		return "", err
	}
	s += depS
`, strcase.ToCamel(depName))
	}
	mainSrc += "return s, nil\n}\n"
	return mainSrc
}

// Helper function to create module config
func createModuleConfig(name string, depNames []string) map[string]interface{} {
	deps := make([]map[string]string, 0)
	for _, depName := range depNames {
		// Use absolute path from workspace root, but without /work prefix
		deps = append(deps, map[string]string{
			"name":   depName,
			"source": filepath.Join("..", depName),
		})
	}

	return map[string]interface{}{
		"name":         name,
		"sdk":          "go",
		"dependencies": deps,
	}
}

// CwlbraaBenchmarksModules creates a stack of nested Dagger modules to benchmark performance
func (m *CwlbraaBenchmarks) BenchmarkModules(ctx context.Context, cliBin *dagger.File) (*dagger.Directory, error) {
	modGen := dag.Container().
		From("golang:1.21").
		WithMountedFile("/bin/dagger", cliBin).
		WithWorkdir("/work").
		WithExec([]string{"git", "init"})

	modCount := 0
	var curDeps []string

	// Helper to add new modules with dependencies
	addModulesWithDeps := func(newMods int, depNames []string) []string {
		var newModNames []string

		for i := 0; i < newMods; i++ {
			name := fmt.Sprintf("mod%d", modCount)
			modCount++
			newModNames = append(newModNames, name)

			// Create module directory at workspace root level
			modGen = modGen.WithWorkdir("/work/" + name)

			// Generate main.go
			mainContent := getModMainSrc(name, depNames)
			modGen = modGen.WithNewFile("main.go", mainContent)

			// Create dagger.json
			configContent, _ := json.MarshalIndent(createModuleConfig(name, depNames), "", "  ")
			modGen = modGen.WithNewFile("dagger.json", string(configContent))
		}

		return newModNames
	}

	// Create initial module
	curDeps = addModulesWithDeps(1, nil)

	// Create 6 layers of modules
	for i := 0; i < 6; i++ {
		curDeps = addModulesWithDeps(len(curDeps)+1, curDeps)
	}

	// Create final top module
	addModulesWithDeps(1, curDeps)

	// Create root dagger.json
	rootConfig := map[string]interface{}{
		"name":   "root",
		"sdk":    "go",
		"source": ".",
	}
	configContent, _ := json.MarshalIndent(rootConfig, "", "  ")
	modGen = modGen.WithNewFile("../dagger.json", string(configContent))

	// return modGen.
	// 	WithWorkdir("/work").
	// 	WithExec([]string{"dagger", "call", "--mod", "mod28", "fn"}, dagger.ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).
	// 	Sync(ctx)
	return modGen.Directory("/work"), nil
}

func (m *CwlbraaBenchmarks) TestPacketLoss(ctx context.Context) (string, error) {
	return dag.Container().
		From("ubuntu:24.10").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "iproute2", "iptables", "curl", "iputils-ping"}).
		// WithEnvVariable("cache_bust", uuid.NewString()).
		WithExec([]string{
			"bash", "-c",
			"iptables -A INPUT -m statistic --mode random --probability 0.15 -j DROP && " +
				"iptables -A OUTPUT -m statistic --mode random --probability 0.15 -j DROP && " +
				"echo 'Network emulation configured with 15% packet  wat drops' && " +
				"echo '-------------------' && " +
				"ping -c 10 8.8.8.8 &&" +
				"for i in {1..10}; do " +
				"echo \"Request $i:\" && " +
				"curl -v -s -w '\\nHTTP Status: %{http_code}\\nTime: %{time_total}s\\n' " +
				"http://httpstat.us/200 || true; " +
				"echo '-------------------'; " +
				"sleep 1; " +
				"done"}, dagger.ContainerWithExecOpts{InsecureRootCapabilities: true}).
		Stdout(ctx)
}
