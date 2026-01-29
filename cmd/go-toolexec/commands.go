// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/creachadair/command"
	"github.com/grafana/go-cache-plugin/lib/otel"
	"github.com/grafana/go-cache-plugin/lib/toolexec"
	"go.opentelemetry.io/otel/attribute"
)

var toolexecFlags struct {
	TracesFile       string `flag:"tracesFile,default=$TOOLEXEC_TRACING_TRACE_FILE,File used for logs"`
	GithubRepo       string `flag:"githubRepo,default=$GITHUB_REPO,Repo name (optional)"`
	GithubRunId      string `flag:"githubRunId,default=$GITHUB_RUN_ID,Run ID (optional)"`
	GithubRunAttempt string `flag:"githubRunAttempt,default=$GITHUB_RUN_ATTEMPT,Run attempt (optional)"`
	GithubJobName    string `flag:"githubJobName,default=$GITHUB_JOB_NAME,Job name (optional)"`
	GithubStepName   string `flag:"githubStepName,default=$GITHUB_STEP_ID,Step name (optional)"`
}

func runToolexec(env *command.Env, args []string) error {
	ctx := env.Context()
	shutdownTracer, tracingContext, err := initTracingProvider(ctx)
	if err != nil {
		return err
	}
	defer shutdownTracer(ctx)

	tool := args[0]
	rest := args[1:]
	pkg := toolexec.GetPackage(rest)

	_, span := tracingContext.SpanWithContext(
		ctx,
		toolexec.GetTool(tool),
		attribute.KeyValue{Key: "package", Value: attribute.StringValue(pkg)},
	)

	defer span.End()
	cmd := exec.Command(tool, rest...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initTracingProvider(ctx context.Context) (func(context.Context) error, *otel.TracingContext, error) {
	var err error
	var shutdown func(context.Context) error
	var tracingContext *otel.TracingContext

	tracingContext = otel.NewTracingContextFromRunData(toolexecFlags.GithubRepo, toolexecFlags.GithubRunId, toolexecFlags.GithubRunAttempt, toolexecFlags.GithubJobName, toolexecFlags.GithubStepName)

	if toolexecFlags.TracesFile != "" {
		shutdown, err = otel.SetupLoggingProvider(ctx, toolexecFlags.TracesFile)
	} else {
		shutdown, err = otel.SetupOtelTraceProvider(ctx)
	}
	return shutdown, tracingContext, err
}
