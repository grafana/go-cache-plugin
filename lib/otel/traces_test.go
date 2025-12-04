package otel

import (
	"testing"
)

var (
	runId      = "19901710910"
	runAttempt = "1"
	jobName    = "build"
	stepName   = "Build Grafana"
)

func TestGenerateTraceID(t *testing.T) {
	expected := "a0bedec6b51cf41be0bbf80ec30055fb"
	actual := GenerateTraceID(runId, runAttempt)

	if actual != expected {
		t.Errorf("GenerateTraceID(%s, %s) = %s; want %s", runId, runAttempt, actual, expected)
	}
}

func TestGenerateJobSpanID(t *testing.T) {
	expected := "ab7526838c2cc769"
	actual := GenerateJobSpanID(runId, runAttempt, jobName)

	if actual != expected {
		t.Errorf("GenerateJobSpanID(%s, %s, %s) = %s; want %s", runId, runAttempt, jobName, actual, expected)
	}
}

func TestGenerateStepSpanID(t *testing.T) {
	expected := "47139316cc3bbb29"
	actual := GenerateStepSpanID(runId, runAttempt, jobName, stepName)

	if actual != expected {
		t.Errorf("GenerateStepSpanID(%s, %s, %s, %s) = %s; want %s", runId, runAttempt, jobName, stepName, actual, expected)
	}
}

func TestDeterministic(t *testing.T) {
	trace1 := GenerateTraceID(runId, runAttempt)
	trace2 := GenerateTraceID(runId, runAttempt)
	if trace1 != trace2 {
		t.Errorf("TraceID not deterministic: %s != %s", trace1, trace2)
	}

	job1 := GenerateJobSpanID(runId, runAttempt, jobName)
	job2 := GenerateJobSpanID(runId, runAttempt, jobName)
	if job1 != job2 {
		t.Errorf("JobSpanID not deterministic: %s != %s", job1, job2)
	}

	step1 := GenerateStepSpanID(runId, runAttempt, jobName, stepName)
	step2 := GenerateStepSpanID(runId, runAttempt, jobName, stepName)
	if step1 != step2 {
		t.Errorf("StepSpanID not deterministic: %s != %s", step1, step2)
	}
}
