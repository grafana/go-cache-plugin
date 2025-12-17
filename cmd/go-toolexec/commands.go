// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/creachadair/command"
	"github.com/grafana/go-cache-plugin/lib/otel"
	"github.com/grafana/go-cache-plugin/lib/toolexec"
	"go.opentelemetry.io/otel/attribute"
)

var toolexecFlags struct {
	TraceFile    string `flag:"log-file,default=$TOOLEXEC_TRACING_TRACE_FILE,File used for logs"`
	TraceId      string `flag:"traceId,default=$TRACING_TRACE_ID,Trace Id (optional)"`
	ParentSpanId string `flag:"parentSpanId,default=$TRACING_PARENT_SPAN_ID,Parent Span Id (optional)"`
	RunId        string `flag:"runId,default=$RUN_ID,Run ID (optional)"`
	RunAttempt   string `flag:"runAttempt,default=$RUN_ATTEMPT,Run attempt (optional)"`
	JobName      string `flag:"jobName,default=$JOB_NAME,Job name (optional)"`
	StepName     string `flag:"stepName,default=$STEP_NAME,Step name (optional)"`
	StepNumber   string `flag:"stepNumber,default=$STEP_NUMBER,Step number (optional)"`
}

func runToolexec(env *command.Env, args []string) error {
	ctx := env.Context()
	shutdownTracer, tracingContext, err := initModTracing(ctx)
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

func initModTracing(ctx context.Context) (func(context.Context) error, *otel.TracingContext, error) {
	var err error
	var tracingContext *otel.TracingContext

	if toolexecFlags.TraceId != "" && toolexecFlags.ParentSpanId != "" {
		tracingContext, err = otel.NewTracingContext(toolexecFlags.TraceId, toolexecFlags.ParentSpanId)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tracing context: %w", err)
		}
	} else {
		runId := os.Getenv("RUN_ID")
		runAttempt := os.Getenv("RUN_ATTEMPT")
		jobName := os.Getenv("JOB_NAME")
		stepName := os.Getenv("STEP_NAME")
		stepNumber := os.Getenv("STEP_NUMBER")

		tracingContext = otel.NewTracingContextFromRunData(runId, runAttempt, jobName, stepName, stepNumber)
	}

	file := toolexecFlags.TraceFile
	if file == "" {
		file = "trace.json"
	}
	shutdown, err := otel.SetupLoggingProvider(ctx, "gobuild-toolexec", toolexecFlags.TraceFile)
	if err != nil {
		return nil, nil, err
	}

	return shutdown, tracingContext, err
}
