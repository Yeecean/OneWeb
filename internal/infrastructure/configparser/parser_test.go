package configparser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixtureDir = func() string {
	wd, _ := os.Getwd()
	// 测试运行于包目录下，向上定位 tests/fixtures/config
	return filepath.Join(wd, "..", "..", "..", "tests", "fixtures", "config")
}()

func allFixtures() []string {
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".conf") {
			files = append(files, filepath.Join(fixtureDir, e.Name()))
		}
	}
	return files
}

func readFixture(name string) string {
	b, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestLexMinimal(t *testing.T) {
	doc, err := Parse(strings.NewReader(`sync_dir = "~/OneDrive"`))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Nodes))
	}
	n := doc.Nodes[0]
	if n.Type.String() != "key_value" || n.Key != "sync_dir" || n.Value != "~/OneDrive" {
		t.Fatalf("unexpected node: %+v", n)
	}
}

func TestLexInlineComment(t *testing.T) {
	doc, _ := Parse(strings.NewReader("threads = 8 # max 16\nsync_dir = \"~/OD\" # primary\n"))
	n := doc.Nodes[0]
	if n.InlineComment != "max 16" {
		t.Fatalf("inline comment = %q", n.InlineComment)
	}
	if n.Value != "8" {
		t.Fatalf("value = %q", n.Value)
	}
	if doc.Nodes[1].InlineComment != "primary" {
		t.Fatalf("second inline comment = %q", doc.Nodes[1].InlineComment)
	}
}

func TestLexUnicodePath(t *testing.T) {
	doc, _ := Parse(strings.NewReader(`sync_dir = "我的台式盘"`))
	n := doc.Nodes[0]
	if n.Value != "我的台式盘" {
		t.Fatalf("unicode value = %q", n.Value)
	}
}

func TestLexBOM(t *testing.T) {
	doc, _ := Parse(strings.NewReader("\uFEFFsync_dir = \"/OD\""))
	if doc.Nodes[0].Key != "sync_dir" {
		t.Fatalf("BOM not skipped: %+v", doc.Nodes[0])
	}
}

func TestLexBlankAndComment(t *testing.T) {
	doc, _ := Parse(strings.NewReader("\n# hello\n# sync_dir = \"/od\"\n"))
	if doc.Nodes[0].Type.String() != "blank_line" {
		t.Fatalf("line1 should be blank: %+v", doc.Nodes[0])
	}
	if doc.Nodes[1].Type.String() != "comment" {
		t.Fatalf("line2 should be comment: %+v", doc.Nodes[1])
	}
	if doc.Nodes[2].Type.String() != "disabled_key_value" {
		t.Fatalf("line3 should be disabled: %+v", doc.Nodes[2])
	}
	if doc.Nodes[2].Key != "sync_dir" || doc.Nodes[2].Value != "/od" {
		t.Fatalf("disabled node parse: %+v", doc.Nodes[2])
	}
}

func TestLexMalformedPreservedAsComment(t *testing.T) {
	doc, err := Parse(strings.NewReader("thsi is not valid syntax ???\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Type.String() != "comment" {
		t.Fatalf("malformed line should be preserved as comment: %+v", doc.Nodes[0])
	}
	if doc.Nodes[0].RawText != "thsi is not valid syntax ???" {
		t.Fatalf("raw text must be preserved: %q", doc.Nodes[0].RawText)
	}
}

func TestParseEmpty(t *testing.T) {
	doc, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 0 {
		t.Fatalf("empty doc should have no nodes, got %d", len(doc.Nodes))
	}
}

func TestParseLineNumbers(t *testing.T) {
	doc, _ := Parse(strings.NewReader("a = 1\n\n# c\nb = 2\n"))
	expected := []int{1, 2, 3, 4}
	for i, want := range expected {
		if doc.Nodes[i].LineNumber != want {
			t.Fatalf("node %d line = %d, want %d", i, doc.Nodes[i].LineNumber, want)
		}
	}
}
