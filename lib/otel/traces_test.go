package otel

import (
	"testing"
)

var (
	repo       = "grafana/grafana"
	runId      = "20137834310"
	runAttempt = "1"
	jobName    = "build"
	stepName   = "Build Grafana"
	stepNumber = "8"
)

func TestGenerateTraceID(t *testing.T) {
	expected := "29f022cfd2ef0bec231e3facb2e5e888"
	actual := GenerateTraceID(repo, runId, runAttempt)

	if actual != expected {
		t.Errorf("GenerateTraceID(%s, %s, %s) = %s; want %s", repo, runId, runAttempt, actual, expected)
	}
}

func TestGenerateRootSpanID(t *testing.T) {
	expected := "35dc80e3696a695f"
	actual := GenerateParentSpanID(repo, runId, runAttempt)

	if actual != expected {
		t.Errorf("GenerateParentSpanID(%s, %s, %s) = %s; want %s", repo, runId, runAttempt, actual, expected)
	}
}

func TestGenerateJobSpanID(t *testing.T) {
	expected := "38d13486123da355"
	actual := GenerateJobSpanID(repo, runId, runAttempt, jobName)

	if actual != expected {
		t.Errorf("GenerateJobSpanID(%s, %s, %s, %s) = %s; want %s", repo, runId, runAttempt, jobName, actual, expected)
	}
}

func TestGenerateStepSpanID(t *testing.T) {
	expected := "52b5804425ed1c83"
	actual := GenerateStepSpanID(repo, runId, runAttempt, jobName, stepName)

	if actual != expected {
		t.Errorf("GenerateStepSpanID(%s, %s, %s, %s, %s) = %s; want %s", repo, runId, runAttempt, jobName, stepName, actual, expected)
	}
}

func TestGenerateStepSpanID_Number(t *testing.T) {
	expected := "5ac3072b544fdedc"
	actual := GenerateStepSpanID_Number(repo, runId, runAttempt, jobName, stepNumber)

	if actual != expected {
		t.Errorf("GenerateStepSpanID_Number(%s, %s, %s, %s, %s) = %s; want %s", repo, runId, runAttempt, jobName, stepNumber, actual, expected)
	}
}
func TestDeterministic(t *testing.T) {
	trace1 := GenerateTraceID(repo, runId, runAttempt)
	trace2 := GenerateTraceID(repo, runId, runAttempt)
	if trace1 != trace2 {
		t.Errorf("TraceID not deterministic: %s != %s", trace1, trace2)
	}

	job1 := GenerateJobSpanID(repo, runId, runAttempt, jobName)
	job2 := GenerateJobSpanID(repo, runId, runAttempt, jobName)
	if job1 != job2 {
		t.Errorf("JobSpanID not deterministic: %s != %s", job1, job2)
	}

	step1 := GenerateStepSpanID(repo, runId, runAttempt, jobName, stepName)
	step2 := GenerateStepSpanID(repo, runId, runAttempt, jobName, stepName)
	if step1 != step2 {
		t.Errorf("StepSpanID not deterministic: %s != %s", step1, step2)
	}
}
