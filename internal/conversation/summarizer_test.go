package conversation

import (
	"strings"
	"testing"
)

func TestSummarizerEmpty(t *testing.T) {
	s := NewSummarizer(0)
	report := s.FinalReport()
	if report != "任务完成（无输出）。" {
		t.Errorf("expected empty report, got %q", report)
	}
}

func TestSummarizerDefaultReport(t *testing.T) {
	s := NewSummarizer(0)
	s.Write("Hello world\nDone.")

	report := s.FinalReport()
	if !strings.Contains(report, "任务完成") {
		t.Error("report should contain '任务完成'")
	}
	if !strings.Contains(report, "Hello world") {
		t.Error("report should contain output")
	}
}

func TestSummarizerTestReport(t *testing.T) {
	s := NewSummarizer(0)
	s.Write("=== RUN   TestFoo\n--- PASS: TestFoo\nPASS\nok  package 0.5s")

	report := s.FinalReport()
	if !strings.Contains(report, "测试结果") {
		t.Error("report should contain '测试结果'")
	}
	if !strings.Contains(report, "PASS") {
		t.Error("report should contain PASS")
	}
}

func TestSummarizerErrorReport(t *testing.T) {
	s := NewSummarizer(0)
	s.Write("Starting...\nError: connection refused\nFATAL: cannot connect")

	report := s.FinalReport()
	if !strings.Contains(report, "错误信息") {
		t.Error("report should contain '错误信息'")
	}
	if !strings.Contains(report, "connection refused") {
		t.Error("report should contain error details")
	}
}

func TestSummarizerTruncate(t *testing.T) {
	s := NewSummarizer(100)
	s.Write(strings.Repeat("a", 200))

	report := s.FinalReport()
	if len(report) > 200 {
		t.Errorf("report should be truncated, got length %d", len(report))
	}
}

func TestSummarizerProgressSummary(t *testing.T) {
	s := NewSummarizer(0)
	s.Write("Step 1 done\nStep 2 in progress...")

	summary := s.ProgressSummary()
	if !strings.Contains(summary, "处理中") {
		t.Error("summary should contain '处理中'")
	}
}
