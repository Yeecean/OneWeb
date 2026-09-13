package configparser

import (
	"os"
	"strings"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/config"
)

// TestRoundTrip 是里程碑 1 最重要的测试：
// 对任意合法 config 文件执行 Parse → Render，输出必须与原始输入逐字节一致。
func TestRoundTrip(t *testing.T) {
	for _, fixture := range allFixtures() {
		original, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("read %s: %v", fixture, err)
		}
		doc, err := Parse(strings.NewReader(string(original)))
		if err != nil {
			t.Fatalf("parse %s: %v", fixture, err)
		}
		rendered := Render(doc)
		if string(original) != rendered {
			t.Fatalf("Round-trip failed for %s\n--- original ---\n%q\n--- rendered ---\n%q",
				fixture, string(original), rendered)
		}
	}
}

// TestRoundTripSingleLine 无换行文档往返。
func TestRoundTripSingleLine(t *testing.T) {
	original := `sync_dir = "~/OneDrive"`
	doc, err := Parse(strings.NewReader(original))
	if err != nil {
		t.Fatal(err)
	}
	if got := Render(doc); got != original {
		t.Fatalf("round-trip single line mismatch:\n%q\n!=\n%q", got, original)
	}
}

func TestModifySetExisting(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("minimal.conf")))
	doc.Set("sync_dir", "~/NewLocation")
	rendered := Render(doc)
	reparsed, _ := Parse(strings.NewReader(rendered))
	n := reparsed.Get("sync_dir")
	if n == nil || n.Value != "~/NewLocation" {
		t.Fatalf("Set not applied: %+v", n)
	}
}

func TestModifySetAppendsNewKey(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("minimal.conf")))
	original := Render(doc)
	doc.Set("threads", "8")
	rendered := Render(doc)
	// onedrive 要求所有值带引号（数字/布尔亦如此）
	if !strings.HasSuffix(rendered, "threads = \"8\"\n") {
		t.Fatalf("new key should be appended with quoted value:\n%s", rendered)
	}
	if !strings.Contains(rendered, original) {
		t.Fatalf("original content must be preserved:\n%s", rendered)
	}
}

func TestModifyEnableDisabled(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("disabled_options.conf")))
	doc.Enable("threads")
	rendered := Render(doc)
	reparsed, _ := Parse(strings.NewReader(rendered))
	if n := reparsed.Get("threads"); n == nil || n.Type.String() != "key_value" {
		t.Fatalf("Enable failed: %+v", n)
	}
	// enabled 后值应为带引号格式
	if !strings.Contains(rendered, "threads = \"8\"") {
		t.Fatalf("enabled line should have quoted value, got: %s", rendered)
	}
}

func TestModifyDisableKey(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("minimal.conf")))
	doc.Disable("sync_dir")
	rendered := Render(doc)
	if !strings.Contains(rendered, "# sync_dir =") {
		t.Fatalf("disabled line should be commented: %s", rendered)
	}
}

// TestUnknownOptionsPreserved 红线 4 关键验收：
// 修改已知项后，未知选项必须原样保留。
func TestUnknownOptionsPreserved(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("unknown_options.conf")))
	doc.Set("threads", "4")
	rendered := Render(doc)
	for _, key := range []string{"unknown_future_option", "another_unknown", "experimental_mode"} {
		if !strings.Contains(rendered, key) {
			t.Fatalf("unknown option %q was lost:\n%s", key, rendered)
		}
	}
}

func TestModifyCommentsOnlyKeepsComments(t *testing.T) {
	doc, _ := Parse(strings.NewReader(readFixture("comments_only.conf")))
	doc.Set("sync_dir", "~/OD")
	rendered := Render(doc)
	original := readFixture("comments_only.conf")
	if !strings.Contains(rendered, strings.TrimSpace(original)) {
		t.Fatalf("comments must be preserved:\n%s", rendered)
	}
}

func TestModifyRoundTripThenRender(t *testing.T) {
	// 修改后重新渲染再解析，应保持一致
	doc, _ := Parse(strings.NewReader(readFixture("full.conf")))
	doc.Set("threads", "12")
	rendered := Render(doc)
	reparsed, _ := Parse(strings.NewReader(rendered))
	if reparsed.Get("threads").Value != "12" {
		t.Fatalf("round trip after modify failed")
	}
	if !strings.Contains(rendered, "Download worker threads") {
		t.Fatalf("comments lost after modify:\n%s", rendered)
	}
}

// TestRenderQuotedValues 核心保证：所有修改节点的值必须带引号，
// 以满足 onedrive 官方解析要求（裸值数字/布尔会被判为 Malformed）。
func TestRenderQuotedValues(t *testing.T) {
	tests := []struct {
		key   string
		value string
		want  string
	}{
		{"threads", "8", "threads = \"8\""},
		{"skip_dotfiles", "true", "skip_dotfiles = \"true\""},
		{"sync_dir", "/tmp/od", "sync_dir = \"/tmp/od\""},
		{"skip_file", "~*.tmp", "skip_file = \"~*.tmp\""},
		{"monitor_interval", "300", "monitor_interval = \"300\""},
	}
	for _, tt := range tests {
		doc := &config.ConfigDocument{}
		doc.Set(tt.key, tt.value)
		rendered := Render(doc)
		if !strings.Contains(rendered, tt.want) {
			t.Fatalf("key %s: expected %q in rendered, got:\n%s", tt.key, tt.want, rendered)
		}
	}
}
