// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana/go-cache-plugin/lib/otel"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creachadair/command"
	"github.com/creachadair/gocache"
	"github.com/creachadair/taskgroup"
)

var flags struct {
	CacheDir             string        `flag:"cache-dir,default=$GOCACHE_DIR,Local cache directory (required)"`
	LogFile              string        `flag:"log-file,default=trace.log,File used for logs"`
	S3Bucket             string        `flag:"bucket,default=$GOCACHE_S3_BUCKET,S3 bucket name (required if no --local flag provided)"`
	S3Region             string        `flag:"region,default=$GOCACHE_S3_REGION,S3 region"`
	S3Endpoint           string        `flag:"s3-endpoint-url,default=$GOCACHE_S3_ENDPOINT_URL,S3 custom endpoint URL (if unset, use AWS default)"`
	S3PathStyle          bool          `flag:"s3-path-style,default=$GOCACHE_S3_PATH_STYLE,S3 path-style URLs (optional)"`
	LocalCache           bool          `flag:"local-cache,default=$GOCACHE_LOCAL,Runs in no cache mode (no S3)"`
	KeyPrefix            string        `flag:"prefix,default=$GOCACHE_KEY_PREFIX,S3 key prefix (optional)"`
	MinUploadSize        int64         `flag:"min-upload-size,default=$GOCACHE_MIN_SIZE,Minimum object size to upload to S3 (in bytes)"`
	Concurrency          int           `flag:"c,default=$GOCACHE_CONCURRENCY,Maximum number of concurrent requests"`
	S3Concurrency        int           `flag:"u,default=$GOCACHE_S3_CONCURRENCY,Maximum concurrency for upload to S3"`
	PrintMetrics         bool          `flag:"metrics,default=$GOCACHE_METRICS,Print summary metrics to stderr at exit"`
	Expiration           time.Duration `flag:"expiry,default=$GOCACHE_EXPIRY,Cache expiration period (optional)"`
	Verbose              bool          `flag:"v,default=$GOCACHE_VERBOSE,Enable verbose logging"`
	DebugLog             int           `flag:"debug,default=$GOCACHE_DEBUG,Enable detailed per-request debug logging (noisy)"`
	TracingEnabled       bool          `flag:"tracing,default=$ENABLE_TRACING,Enable tracing (optional)"`
	OtelCollectorAddress string        `flag:"otel-collector,default=$OTEL_COLLECTOR_ADDRESS,OTEL collector address (optional)"`
	TraceId              string        `flag:"traceId,default=$TRACING_TRACE_ID,Trace Id (optional)"`
	ParentSpanId         string        `flag:"parentSpanId,default=$TRACING_PARENT_SPAN_ID,Parent Span Id (optional)"`
	RunId                string        `flag:"runId,default=$RUN_ID,Run ID (optional)"`
	RunAttempt           string        `flag:"runAttempt,default=$RUN_ATTEMPT,Run attempt (optional)"`
	JobName              string        `flag:"jobName,default=$JOB_NAME,Job name (optional)"`
	StepName             string        `flag:"stepName,default=$STEP_NAME,Step name (optional)"`
	StepNumber           string        `flag:"stepNumber,default=$STEP_NUMBER,Step number (optional)"`
}

const (
	debugBuildCache = 1 << iota
	debugModProxy
	debugRevProxy
)

// runDirect runs a cache communicating on stdin/stdout, for use as a direct
// GOCACHEPROG plugin.
func runDirect(env *command.Env) error {
	s, _, err := initCacheServer(env)
	if err != nil {
		return err
	}
	if err := s.Run(env.Context(), os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("cache server exited with error: %w", err)
	}
	if flags.Verbose || flags.PrintMetrics {
		fmt.Fprintln(os.Stderr, s.Metrics())
	}
	return nil
}

var serveFlags struct {
	Plugin     string `flag:"plugin,default=$GOCACHE_PLUGIN,Plugin service addr (or port) (required)"`
	HTTP       string `flag:"http,default=$GOCACHE_HTTP,HTTP service address ([host]:port)"`
	ModProxy   bool   `flag:"modproxy,default=$GOCACHE_MODPROXY,Enable a Go module proxy (requires --http)"`
	ModNoCache bool   `flag:"mod-nocache,default=$GOCACHE_MODPROXY_NOCACHE,Disable the local module cache (requires --modproxy)"`
	RevProxy   string `flag:"revproxy,default=$GOCACHE_REVPROXY,Reverse proxy these hosts (comma-separated; requires --http)"`
	SumDB      string `flag:"sumdb,default=$GOCACHE_SUMDB,SumDB servers to proxy for (comma-separated)"`
}

func noopClose(context.Context) error { return nil }

// runServe runs a cache communicating over a local TCP socket.
func runServe(env *command.Env) error {
	if serveFlags.Plugin == "" {
		return env.Usagef("you must provide a --plugin addr (or port)")
	}

	// Initialize the cache server. Unlike a direct server, only close down and
	// wait for cache cleanup when the whole process exits.
	s, s3c, err := initCacheServer(env)
	if err != nil {
		return err
	}
	closeHook := s.Close
	s.Close = noopClose

	pluginAddr := serveFlags.Plugin
	if !strings.Contains(pluginAddr, ":") {
		pluginAddr = fmt.Sprintf("127.0.0.1:%s", serveFlags.Plugin)
	}

	// Listen for connections from the Go toolchain on the specified socket.
	lst, err := net.Listen("tcp", pluginAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("plugin listening at %q", lst.Addr())

	ctx, cancel := signal.NotifyContext(env.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var g taskgroup.Group
	g.Run(func() {
		<-ctx.Done()
		log.Printf("closing plugin listener")
		lst.Close()
	})

	// If a module proxy is enabled, start it.
	modProxy, modCleanup, err := initModProxy(env.SetContext(ctx), s3c)
	if err != nil {
		lst.Close()
		return fmt.Errorf("module proxy: %w", err)
	}
	defer modCleanup()

	// If a reverse proxy is enabled, start it.
	revProxy, err := initRevProxy(env.SetContext(ctx), s3c, &g)
	if err != nil {
		lst.Close()
		return fmt.Errorf("reverse proxy: %w", err)
	}

	// If an HTTP server is enabled, start it up with debug routes
	// and whatever other services were requested.
	if serveFlags.HTTP != "" {
		otelCleanup, tracingContext, err := initModTracing(ctx, "gobuild-modcache")
		if err != nil {
			return fmt.Errorf("tracing: %w", err)
		}
		if err != nil {
			return fmt.Errorf("tracing: %w", err)
		}

		srv := &http.Server{
			Addr:    serveFlags.HTTP,
			Handler: makeHandler(modProxy, revProxy, tracingContext),
		}
		g.Go(srv.ListenAndServe)
		vprintf("HTTP server listening at %q", serveFlags.HTTP)
		g.Run(func() {
			<-ctx.Done()
			_ = otelCleanup(ctx)
			vprintf("stopping HTTP service")
			_ = srv.Shutdown(context.Background())
		})
	}

	for {
		conn, err := lst.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("accept failed: %v, exiting server loop", err)
			}
			break
		}
		log.Printf("new client connection")
		g.Go(func() error {
			defer func() {
				log.Printf("client connection closed")
				conn.Close()
			}()
			return s.Run(ctx, conn, conn)
		})
	}
	log.Printf("server loop exited, waiting for client exit")
	g.Wait()
	if closeHook != nil {
		ctx := gocache.WithLogf(context.Background(), log.Printf)
		if err := closeHook(ctx); err != nil {
			log.Printf("server close: %v (ignored)", err)
		}
	}
	return nil
}

// runConnect implements a direct cache proxy by connecting to a remote server.
func runConnect(env *command.Env, plugin string) error {

	ctx := env.Context()
	shutdownTracer, reportSpan, err := initTracing(ctx, "gobuild-gocacheprog-connect")
	if err != nil {
		return err
	}
	addr := plugin
	// If the caller has not specified a host/port, then likely this is an older usage which only specifies port
	if !strings.Contains(plugin, ":") {
		port, err := strconv.Atoi(plugin)
		if err != nil {
			return fmt.Errorf("invalid plugin port: %w", err)
		}

		addr = fmt.Sprintf(":%d", port)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	start := time.Now()
	vprintf("connected to %q", conn.RemoteAddr())

	out := taskgroup.Go(func() error {
		defer conn.(*net.TCPConn).CloseWrite() // let the server finish
		return copy(conn, os.Stdin, reportSpan)
	})
	if rerr := copy(os.Stdout, conn, reportSpan); rerr != nil {
		vprintf("read responses: %v", err)
	}
	out.Wait()
	conn.Close()

	shutdownTracer(context.Background())
	vprintf(fmt.Sprintf("connection closed (%v elapsed)", time.Since(start)))
	return nil
}

func initTracing(ctx context.Context, service string) (func(context.Context) error, func([]byte), error) {
	if !flags.TracingEnabled {
		return func(context.Context) error { return nil }, func([]byte) {}, nil
	}

	var shutdown func(context.Context) error
	var err error
	if flags.OtelCollectorAddress != "" {
		shutdown, err = otel.SetupOtelTraceProvider(ctx, service, flags.OtelCollectorAddress)
	} else if flags.LogFile != "" {
		log.Printf("Otel Collector address not specified, starting with the logging reporter, log file: %s", flags.LogFile)
		shutdown, err = otel.SetupLoggingProvider(ctx, service, flags.LogFile)
	} else {
		log.Printf("please specify either --otel-collector or --log-file to setup tracing or disable tracing")
		return nil, nil, errors.New("otel exporter not initialized")
	}
	if err != nil {
		return nil, nil, err
	}

	tracingContext, err := initTracingContext()
	if err != nil {
		return nil, nil, err
	}
	spanner := otel.NewAwesomeSpanner(tracingContext)

	spanReporter := func(buffer []byte) {
		id, err := parseId(buffer[:])
		if err != nil {
			return
		}
		spanner.ProcessId(ctx, id)
	}

	return shutdown, spanReporter, err
}
func initModTracing(ctx context.Context, service string) (func(context.Context) error, *otel.TracingContext, error) {
	if !flags.TracingEnabled {
		return func(context.Context) error { return nil }, nil, nil
	}

	var shutdown func(context.Context) error
	var err error
	if flags.OtelCollectorAddress != "" {
		shutdown, err = otel.SetupOtelTraceProvider(ctx, service, flags.OtelCollectorAddress)
	} else if flags.LogFile != "" {
		log.Printf("Otel Collector address not specified, starting with the logging reporter, log file: %s", flags.LogFile)
		shutdown, err = otel.SetupLoggingProvider(ctx, service, flags.LogFile)
	} else {
		log.Printf("please specify either --otel-collector or --log-file to setup tracing or disable tracing")
		return nil, nil, errors.New("otel exporter not initialized")
	}

	tracingContext, err := initTracingContext()
	if err != nil {
		return nil, nil, err
	}

	return shutdown, tracingContext, err
}

func initTracingContext() (*otel.TracingContext, error) {
	if flags.TraceId != "" && flags.ParentSpanId != "" {
		return otel.NewTracingContext(flags.TraceId, flags.ParentSpanId)
	} else {
		return otel.NewTracingContextFromRunData(flags.RunId, flags.RunAttempt, flags.JobName, flags.StepName, flags.StepNumber), nil
	}
}

// copy emulates the base case of io.Copy, but does not attempt to use the
// io.ReaderFrom or io.WriterTo implementations.
//
// TODO(creachadair): For some reason io.Copy does not work correctly when r is
// a pipe (e.g., stdin) and w is a TCP socket. Figure out why.
func copy(w io.Writer, r io.Reader, reportTrace func([]byte)) error {
	var buf [4096]byte
	for {
		nr, err := r.Read(buf[:])
		if nr > 0 {
			if nw, err := w.Write(buf[:nr]); err != nil {
				return fmt.Errorf("copy to: %w", err)
			} else if nw < nr {
				return fmt.Errorf("wrote %d < %d bytes: %w", nw, nr, io.ErrShortWrite)
			}

			reportTrace(buf[:])
		}
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("copy from: %w", err)
		}
	}
}

type name struct {
}

func parseId(buf []byte) (otel.CacheRequest, error) {
	var data map[string]any

	err := json.Unmarshal(buf, &data)
	if err != nil {
		slice, _, found := bytes.Cut(buf, []byte{'\n'})
		if found {
			err2 := json.Unmarshal(slice, &data)
			if err2 != nil {
				return otel.CacheRequest{}, err
			}
		}
	}

	var id string
	var miss bool
	var command string
	var actionId string

	id1, ok := data["ID"]
	if !ok {
		return otel.CacheRequest{}, errors.New("id not found in the request")
	} else {
		id = fmt.Sprint(id1)
	}

	command1, ok := data["Command"]
	if !ok {
		command = ""
	} else {
		command = fmt.Sprint(command1)
	}

	_, ok = data["Miss"]
	if !ok {
		miss = false
	} else {
		miss = true
	}

	actionId1, ok := data["ActionID"]
	if !ok {
		actionId = ""
	} else {
		actionId = fmt.Sprint(actionId1)
	}

	data2 := otel.CacheRequest{
		Id:       id,
		ActionId: actionId,
		Miss:     miss,
		Command:  command,
	}
	return data2, nil
}
