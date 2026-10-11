package telegram

import (
	"strings"
	"testing"
)

func TestBuildPullRequestMessage(t *testing.T) {
	event := newPREvent(1, "hello < world", "open")
	msg := BuildPullRequestMessage(event)

	if !strings.Contains(msg, "hello &lt; world") {
		t.Errorf("Expected escaped string in msg, got %s", msg)
	}

	if !strings.Contains(msg, "<b>Status:</b> open") {
		t.Errorf("Expected status open, got %s", msg)
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
	}

	for _, tt := range tests {
		got := escapeHTML(tt.input)
		if got != tt.expected {
			t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
