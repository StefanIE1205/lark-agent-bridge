package conversation

import (
	"strings"
)

// Summarizer collects agent output and produces structured summaries.
type Summarizer struct {
	buf      strings.Builder
	maxChars int
}

// NewSummarizer creates a new output summarizer.
func NewSummarizer(maxChars int) *Summarizer {
	if maxChars <= 0 {
		maxChars = 2000
	}
	return &Summarizer{maxChars: maxChars}
}

// Write adds text to the buffer.
func (s *Summarizer) Write(text string) {
	s.buf.WriteString(text)
}

// ProgressSummary returns a brief progress update.
func (s *Summarizer) ProgressSummary() string {
	text := s.buf.String()
	if len(text) == 0 {
		return "处理中..."
	}

	// Take last 200 chars for progress
	if len(text) > 200 {
		text = text[len(text)-200:]
	}

	return "处理中...\n\n" + strings.TrimSpace(text)
}

// FinalReport returns a structured final report.
func (s *Summarizer) FinalReport() string {
	text := s.buf.String()
	if len(text) == 0 {
		return "任务完成（无输出）。"
	}

	// Truncate if too long
	if len(text) > s.maxChars {
		text = text[len(text)-s.maxChars:]
	}

	// Check for common patterns
	lower := strings.ToLower(text)

	// Test results
	if strings.Contains(lower, "pass") || strings.Contains(lower, "fail") {
		return s.formatTestReport(text)
	}

	// Error patterns
	if strings.Contains(lower, "error") || strings.Contains(lower, "panic") || strings.Contains(lower, "fatal") {
		return s.formatErrorReport(text)
	}

	// Default format
	return s.formatDefaultReport(text)
}

func (s *Summarizer) formatTestReport(text string) string {
	var sb strings.Builder
	sb.WriteString("任务完成\n\n")

	// Extract test summary
	lines := strings.Split(text, "\n")
	var summary []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "ok") || strings.Contains(lower, "fail") ||
			strings.Contains(lower, "pass") || strings.Contains(lower, "---") {
			summary = append(summary, line)
		}
	}

	if len(summary) > 0 {
		sb.WriteString("测试结果：\n")
		for _, line := range summary {
			sb.WriteString("- ")
			sb.WriteString(strings.TrimSpace(line))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(text)
	}

	return sb.String()
}

func (s *Summarizer) formatErrorReport(text string) string {
	var sb strings.Builder
	sb.WriteString("执行遇到问题\n\n")

	// Extract error lines
	lines := strings.Split(text, "\n")
	var errors []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "panic") ||
			strings.Contains(lower, "fatal") || strings.Contains(lower, "exception") {
			errors = append(errors, line)
		}
	}

	if len(errors) > 0 {
		sb.WriteString("错误信息：\n")
		for _, line := range errors {
			sb.WriteString("- ")
			sb.WriteString(strings.TrimSpace(line))
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(text)
	}

	return sb.String()
}

func (s *Summarizer) formatDefaultReport(text string) string {
	var sb strings.Builder
	sb.WriteString("任务完成\n\n")
	sb.WriteString(text)
	return sb.String()
}
