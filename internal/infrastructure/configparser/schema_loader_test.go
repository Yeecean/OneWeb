package configparser

import (
	"testing"

	"github.com/yeecean/oneweb/internal/domain/config"
)

func TestLoadSemanticSchemaV25(t *testing.T) {
	schema, err := LoadSemanticSchema("2.5")
	if err != nil {
		t.Fatal(err)
	}
	if schema.SchemaVersion != "2.5" {
		t.Fatalf("schema version = %q", schema.SchemaVersion)
	}
	if schema.GetOption("sync_dir") == nil {
		t.Fatal("sync_dir option missing")
	}
	if schema.GetOption("threads") == nil {
		t.Fatal("threads option missing")
	}
	threads := schema.GetOption("threads")
	if threads.Type != config.TypeInt {
		t.Fatalf("threads type = %v", threads.Type)
	}
	if threads.Constraints == nil || *threads.Constraints.Min != 1 || *threads.Constraints.Max != 16 {
		t.Fatalf("threads constraints = %+v", threads.Constraints)
	}
	if schema.GetOption("nonexistent") != nil {
		t.Fatal("nonexistent option should be nil")
	}
}

func TestLoadUISchemaV25(t *testing.T) {
	ui, err := LoadUISchema("2.5")
	if err != nil {
		t.Fatal(err)
	}
	if len(ui.Groups) < 4 {
		t.Fatalf("expected >=4 groups, got %d", len(ui.Groups))
	}
	w, ok := ui.Widgets["threads"]
	if !ok {
		t.Fatal("threads widget missing")
	}
	if w.Widget != "slider" {
		t.Fatalf("threads widget = %q", w.Widget)
	}
}
