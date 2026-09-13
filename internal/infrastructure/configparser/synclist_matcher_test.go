package configparser

import (
	"strings"
	"testing"
)

func mustSet(t *testing.T, s string) *Matcher {
	t.Helper()
	set, err := ParseSyncList(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	return NewMatcher(set)
}

func TestMatcherDefaultExcludesAll(t *testing.T) {
	m := mustSet(t, "/Documents\n")
	if m.IsIncluded("Videos/foo.mp4") {
		t.Fatal("unmatched path must be excluded")
	}
}

func TestMatcherRootedInclude(t *testing.T) {
	m := mustSet(t, "/Documents\n")
	if !m.IsIncluded("Documents/report.pdf") {
		t.Fatal("Documents/report.pdf should be included")
	}
	if m.IsIncluded("Archive/Documents/x") {
		t.Fatal("nested Documents must not match rooted rule")
	}
}

func TestMatcherBareNameRecursive(t *testing.T) {
	m := mustSet(t, "Documents\n")
	if !m.IsIncluded("Archive/Inner/Documents/foo.txt") {
		t.Fatal("bare name should match at any depth")
	}
}

func TestMatcherExcludeOverrides(t *testing.T) {
	m := mustSet(t, "/Documents\n!/Documents/Temp\n")
	if !m.IsIncluded("Documents/keep.txt") {
		t.Fatal("Documents/keep.txt should be included")
	}
	if m.IsIncluded("Documents/Temp/x.txt") {
		t.Fatal("Documents/Temp should be excluded (later rule wins)")
	}
}

func TestMatcherWildcard(t *testing.T) {
	m := mustSet(t, "/Documents/*.doc\n")
	if !m.IsIncluded("Documents/a.doc") {
		t.Fatal("a.doc should match")
	}
	if m.IsIncluded("Documents/a.txt") {
		t.Fatal("a.txt should not match *.doc")
	}
	if m.IsIncluded("Documents/sub/a.doc") {
		t.Fatal("nested path should not match single-level *")
	}
}

func TestMatcherDoubleStar(t *testing.T) {
	m := mustSet(t, "/Videos/**/*.mp4\n")
	if !m.IsIncluded("Videos/2024/x.mp4") {
		t.Fatal("Videos/2024/x.mp4 should match **")
	}
	if !m.IsIncluded("Videos/a/b/c/x.mp4") {
		t.Fatal("deep nested mp4 should match")
	}
	if m.IsIncluded("Videos/2024/x.mkv") {
		t.Fatal("mkv should not match")
	}
}

func TestMatcherOrderSensitivity(t *testing.T) {
	// 后置规则覆盖前置规则
	m := mustSet(t, "/Documents\n!/Documents/Private\n/Documents/Private/Keep\n")
	if !m.IsIncluded("Documents/Private/Keep/file.txt") {
		t.Fatal("Keep should be re-included by later rule")
	}
	if m.IsIncluded("Documents/Private/other.txt") {
		t.Fatal("other in Private should stay excluded")
	}
}

func TestMatcherOrderReversed(t *testing.T) {
	// 先排除后包含 → 包含胜出（后置规则优先）
	m := mustSet(t, "!/Documents/Private\n/Documents\n")
	if !m.IsIncluded("Documents/Private/x") {
		t.Fatal("later include should override earlier exclude")
	}
}
