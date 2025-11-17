package types

import (
	"errors"
	"testing"
)

func TestDocumentTypeString(t *testing.T) {
	tests := []struct {
		name     string
		docType  DocumentType
		expected string
	}{
		{"HTML type", TypeHTML, "HTML"},
		{"Markdown type", TypeMarkdown, "Markdown"},
		{"RST type", TypeReStructuredText, "reStructuredText"},
		{"Unknown type", TypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.docType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStatsAddError(t *testing.T) {
	stats := &Stats{}
	err := errors.New("test error")

	stats.AddError(err)

	if len(stats.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(stats.Errors))
	}

	if stats.Errors[0] != err {
		t.Errorf("expected error %v, got %v", err, stats.Errors[0])
	}
}

func TestStatsHasErrors(t *testing.T) {
	stats := &Stats{}

	if stats.HasErrors() {
		t.Error("expected HasErrors to return false for empty stats")
	}

	stats.AddError(errors.New("test error"))

	if !stats.HasErrors() {
		t.Error("expected HasErrors to return true after adding error")
	}
}
