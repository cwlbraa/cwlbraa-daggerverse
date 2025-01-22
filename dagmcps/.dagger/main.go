// A generated module for Dagmcps functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/dagmcps/internal/dagger"
)

type Dagmcps struct {
	Src *dagger.Directory
}

// Container packages the binary in a lightweight Alpine container
func (m *Dagmcps) DagmcpsContainer(ctx context.Context) *dagger.Container {
	return dag.Container().
		From("alpine:latest").
		WithFile("/usr/local/bin/dagmcps", m.Binary(ctx))
}

func New(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=["bin", ".git", "**/node_modules", "**/.venv", "**/__pycache__"]
	src *dagger.Directory,
) *Dagmcps {
	return &Dagmcps{
		Src: src,
	}
}

// Build creates the dagmcps binary
func (m *Dagmcps) Binary(ctx context.Context) *dagger.File {
	return dag.Container().
		From("golang:latest").
		WithDirectory("/work", m.Src).
		WithWorkdir("/work/dagmcps").
		WithExec([]string{"go", "build", "-o", "dagmcps"}).
		File("./dagmcps")
}
