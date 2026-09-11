package main

import (
	"bytes"
	"strings"
	"testing"
)

// These exercise the command itself, which is only possible because main
// hands its globals to run instead of reaching for them.

func TestRunRejectsBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no argument", nil},
		{"two arguments", []string{"45m", "10m"}},
		{"not a duration", []string{"banana"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args, &bytes.Buffer{}, &bytes.Buffer{}, strings.NewReader("")); err == nil {
				t.Errorf("run(%q) returned no error", tt.args)
			}
		})
	}
}

func TestRunPrintsUsageOnHelp(t *testing.T) {
	var stderr bytes.Buffer

	if err := run([]string{"-h"}, &bytes.Buffer{}, &stderr, strings.NewReader("")); err != nil {
		t.Fatalf("run returned %v", err)
	}
	if !strings.Contains(stderr.String(), "usage: timer") {
		t.Errorf("stderr = %q, want the usage text", stderr.String())
	}
}
