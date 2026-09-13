package configparser

import (
	"strings"
	"testing"
)

func TestValidateTypeError(t *testing.T) {
	schema, _ := LoadSemanticSchema("2.5")
	doc, _ := Parse(strings.NewReader("threads = \"abc\"\n"))
	results := ValidateAgainstSchema(doc, schema)
	if !HasErrors(results) {
		t.Fatalf("expected type error, got: %s", Summarize(results))
	}
}

func TestValidateRangeError(t *testing.T) {
	schema, _ := LoadSemanticSchema("2.5")
	doc, _ := Parse(strings.NewReader("threads = 999\n"))
	results := ValidateAgainstSchema(doc, schema)
	if !HasErrors(results) {
		t.Fatalf("expected range error, got: %s", Summarize(results))
	}
}

func TestValidateValid(t *testing.T) {
	schema, _ := LoadSemanticSchema("2.5")
	doc, _ := Parse(strings.NewReader("threads = 8\nsync_root_files = \"false\"\n"))
	results := ValidateAgainstSchema(doc, schema)
	if len(results) != 0 {
		t.Fatalf("expected no errors, got: %s", Summarize(results))
	}
}

func TestUnknownOptionIsWarning(t *testing.T) {
	schema, _ := LoadSemanticSchema("2.5")
	doc, _ := Parse(strings.NewReader("future_opt = 1\nthreads = 4\n"))
	results := ValidateAgainstSchema(doc, schema)
	if HasErrors(results) {
		t.Fatalf("unknown option should not be an error: %s", Summarize(results))
	}
	foundWarning := false
	for _, r := range results {
		if r.Level == "warning" && r.Key == "future_opt" {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Fatalf("expected warning for unknown option: %s", Summarize(results))
	}
}

func TestBoolValidation(t *testing.T) {
	schema, _ := LoadSemanticSchema("2.5")
	doc, _ := Parse(strings.NewReader("sync_root_files = \"yes\"\n"))
	results := ValidateAgainstSchema(doc, schema)
	if !HasErrors(results) {
		t.Fatalf("expected bool error, got: %s", Summarize(results))
	}
}
