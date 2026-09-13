package configparser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/sync"
)

var synclistDir = func() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "..", "..", "tests", "fixtures", "synclist")
}()

func synclistFixtures() []string {
	entries, err := os.ReadDir(synclistDir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			files = append(files, e.Name())
		}
	}
	return files
}

func readSynclistFixture(name string) string {
	b, err := os.ReadFile(filepath.Join(synclistDir, name))
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestParseSyncListBasic(t *testing.T) {
	set, err := ParseSyncList(strings.NewReader("/Documents/*\n/Pictures/*\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(set.Rules))
	}
	if set.Rules[0].Type != sync.RuleInclude || !set.Rules[0].IsRooted {
		t.Fatalf("rule0 = %+v", set.Rules[0])
	}
}

func TestParseSyncListExclude(t *testing.T) {
	set, err := ParseSyncList(strings.NewReader("/Documents\n!/Documents/Temp\n"))
	if err != nil {
		t.Fatal(err)
	}
	if set.Rules[1].Type != sync.RuleExclude {
		t.Fatalf("rule1 type = %s", set.Rules[1].Type)
	}
	if set.Rules[1].Pattern != "/Documents/Temp" {
		t.Fatalf("rule1 pattern = %q", set.Rules[1].Pattern)
	}
}

func TestParseSyncListComments(t *testing.T) {
	set, err := ParseSyncList(strings.NewReader("# comment\n/Documents\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Rules) != 2 {
		t.Fatalf("comments should be preserved as rules, got %d", len(set.Rules))
	}
	if !strings.HasPrefix(set.Rules[0].RawText, "#") {
		t.Fatalf("comment rule raw = %q", set.Rules[0].RawText)
	}
}

func TestParseSyncListEmpty(t *testing.T) {
	set, err := ParseSyncList(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Rules) != 0 {
		t.Fatalf("expected empty ruleset")
	}
}

func TestCompileSyncListRoundTrip(t *testing.T) {
	for _, fixture := range synclistFixtures() {
		original := readSynclistFixture(fixture)
		set, err := ParseSyncList(strings.NewReader(original))
		if err != nil {
			t.Fatalf("parse %s: %v", fixture, err)
		}
		compiled := CompileSyncList(set)
		if compiled != original {
			t.Fatalf("round-trip failed for %s\n--- original ---\n%q\n--- compiled ---\n%q", fixture, original, compiled)
		}
	}
}

func TestValidateSyncListPerformance(t *testing.T) {
	set, _ := ParseSyncList(strings.NewReader("Documents\nPictures\n"))
	warnings := ValidateSyncList(set)
	perfCount := 0
	for _, w := range warnings {
		if w.Level == "performance" {
			perfCount++
		}
	}
	if perfCount != 2 {
		t.Fatalf("expected 2 performance warnings, got %d: %+v", perfCount, warnings)
	}
}

func TestValidateSyncListRooted(t *testing.T) {
	set, _ := ParseSyncList(strings.NewReader("/Documents\n/Pictures\n"))
	warnings := ValidateSyncList(set)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for rooted rules, got %+v", warnings)
	}
}

func TestValidateSyncListOrdering(t *testing.T) {
	set, _ := ParseSyncList(strings.NewReader("/Documents\n!/Documents/Temp\n"))
	warnings := ValidateSyncList(set)
	orderingFound := false
	for _, w := range warnings {
		if w.Level == "ordering" {
			orderingFound = true
		}
	}
	if !orderingFound {
		t.Fatalf("expected ordering warning, got %+v", warnings)
	}
}
