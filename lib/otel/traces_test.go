package otel

import (
	"testing"
)

var (
	runId      = "20137834310"
	runAttempt = "1"
	jobName    = "build"
	//jobNumber  = "57796136451"
	stepName   = "Build Grafana"
	stepNumber = "8"
)

func TestGenerateTraceID(t *testing.T) {
	expected := "cfe93a0cde6f53f539e0eaff28e05efc"
	actual := GenerateTraceID(runId, runAttempt)

	if actual != expected {
		t.Errorf("GenerateTraceID(%s, %s) = %s; want %s", runId, runAttempt, actual, expected)
	}
}

func TestGenerateRootSpanID(t *testing.T) {
	expected := "7e173813d01cc668"
	actual := GenerateParentSpanID(runId, runAttempt)

	if actual != expected {
		t.Errorf("GenerateParentSpanID(%s, %s) = %s; want %s", runId, runAttempt, actual, expected)
	}
}

func TestGenerateJobSpanID(t *testing.T) {
	expected := "b6f4de49eeda5bb7"
	actual := GenerateJobSpanID(runId, runAttempt, jobName)

	if actual != expected {
		t.Errorf("GenerateJobSpanID(%s, %s, %s) = %s; want %s", runId, runAttempt, jobName, actual, expected)
	}
}

func TestGenerateStepSpanID(t *testing.T) {
	expected := "eecd067487a1f884"
	actual := GenerateStepSpanID(runId, runAttempt, jobName, stepName)

	if actual != expected {
		t.Errorf("GenerateStepSpanID(%s, %s, %s, %s) = %s; want %s", runId, runAttempt, jobName, stepName, actual, expected)
	}
}

func TestGenerateStepSpanID_Number(t *testing.T) {
	expected := "d70487f07693281c"
	actual := GenerateStepSpanID_Number(runId, runAttempt, jobName, stepNumber)

	if actual != expected {
		t.Errorf("GenerateStepSpanID_Number(%s, %s, %s, %s) = %s; want %s", runId, runAttempt, jobName, stepNumber, actual, expected)
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
