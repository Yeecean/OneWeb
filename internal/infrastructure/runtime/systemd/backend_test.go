package systemd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/infrastructure/journal"
)

func profileFor(id string) profile.Profile {
	return profile.Profile{ID: id, RuntimeType: profile.RuntimeSystemd, ConfDir: "/tmp/od"}
}

func profileForTarget(target string) profile.Profile {
	return profile.Profile{ID: "custom", RuntimeType: profile.RuntimeSystemd, ConfDir: "/tmp/od", RuntimeTarget: target}
}

func jsonUnmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

func TestParseStatusActive(t *testing.T) {
	out := `ActiveState=active
SubState=running
MainPID=1234
ExecMainStartTimestamp=Sat 2026-09-12 10:00:00 UTC
MemoryCurrent=524288
`
	rs, err := parseStatus(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(rs.State) != "running" {
		t.Fatalf("state = %s", rs.State)
	}
	if rs.PID != 1234 {
		t.Fatalf("pid = %d", rs.PID)
	}
	if rs.MemoryBytes != 524288 {
		t.Fatalf("mem = %d", rs.MemoryBytes)
	}
}

func TestParseStatusInactive(t *testing.T) {
	out := "ActiveState=inactive\nSubState=dead\n"
	rs, err := parseStatus(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(rs.State) != "stopped" {
		t.Fatalf("state = %s", rs.State)
	}
}

func TestParseStatusFailed(t *testing.T) {
	out := "ActiveState=failed\nSubState=failed\n"
	rs, _ := parseStatus(out)
	if string(rs.State) != "failed" {
		t.Fatalf("state = %s", rs.State)
	}
}

func TestUnitResolution(t *testing.T) {
	b := &Backend{}
	u := b.unit(profileFor("default"))
	if u != "onedrive@default.service" {
		t.Fatalf("unit = %q", u)
	}
}

func TestUnitResolutionCustom(t *testing.T) {
	b := &Backend{}
	u := b.unit(profileForTarget("onedrive-personal.service"))
	if u != "onedrive-personal.service" {
		t.Fatalf("unit = %q", u)
	}
}

func TestStreamLogsSampleJSON(t *testing.T) {
	sample := `{"MESSAGE":"Downloading 5 items","PRIORITY":6,"__REALTIME_TIMESTAMP":"1726113600000000"}`
	var je journal.JournalEntry
	if err := jsonUnmarshal([]byte(sample), &je); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(je.Message, "Downloading") {
		t.Fatalf("message = %q", je.Message)
	}
	if je.Priority != 6 {
		t.Fatalf("priority = %d", je.Priority)
	}
}
