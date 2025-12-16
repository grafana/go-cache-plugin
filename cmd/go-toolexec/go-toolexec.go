// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// Program gocache implements the experimental GOCACHEPROG protocol over an S3
// bucket, for use in builder and CI workers.
package main

import (
	"log"
	"os"

	"github.com/creachadair/command"
	"github.com/creachadair/flax"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	root := &command.C{
		Name:     command.ProgramName(),
		SetFlags: command.Flags(flax.MustBind, &toolexecFlags),
		Run:      command.Adapt(runToolexec),
	}
	command.RunOrFail(root.NewEnv(nil), os.Args[1:])
}
