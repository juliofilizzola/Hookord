package telegram

import (
	"strings"
	"testing"
)

func TestBuildIssueMessage(t *testing.T) {
	event := newIssueEvent(1, "issue & bug", "closed")
	msg := BuildIssueMessage(event)

	if !strings.Contains(msg, "issue &amp; bug") {
		t.Errorf("Expected escaped string in msg, got %s", msg)
	}

	if !strings.Contains(msg, "<b>Status:</b> closed") {
		t.Errorf("Expected status closed, got %s", msg)
	}
}
