package capability

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  ClientVersion
	}{
		{"onedrive v2.5.11", ClientVersion{2, 5, 11}},
		{"v2.5.11", ClientVersion{2, 5, 11}},
		{"2.4.1", ClientVersion{2, 4, 1}},
		{"onedrive v2.4.25-1", ClientVersion{2, 4, 25}},
	}
	for _, tt := range tests {
		got, err := ParseVersion(tt.input)
		if err != nil {
			t.Fatalf("ParseVersion(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseVersion(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestParseVersionInvalid(t *testing.T) {
	if _, err := ParseVersion("no version here"); err == nil {
		t.Fatal("expected error for invalid version string")
	}
}

func TestAtLeast(t *testing.T) {
	if !(ClientVersion{2, 5, 11}).AtLeast(ClientVersion{2, 5, 11}) {
		t.Fatal("equal versions should satisfy AtLeast")
	}
	if !(ClientVersion{2, 6, 0}).AtLeast(ClientVersion{2, 5, 11}) {
		t.Fatal("2.6.0 should be >= 2.5.11")
	}
	if (ClientVersion{2, 4, 25}).AtLeast(ClientVersion{2, 5, 0}) {
		t.Fatal("2.4.25 should be < 2.5.0")
	}
}

func TestEvaluateMatrix(t *testing.T) {
	tests := []struct {
		name    string
		version ClientVersion
		want    FeatureSupport
	}{
		{"modern 2.5.x", ClientVersion{2, 5, 11}, FeatureSupport{true, true, true, true, true, "2.5", true}},
		{"legacy 2.4.10", ClientVersion{2, 4, 10}, FeatureSupport{false, false, false, true, true, "2.4", false}},
		{"2.4.25 ok", ClientVersion{2, 4, 25}, FeatureSupport{true, false, false, true, true, "2.4", true}},
		{"old 1.0", ClientVersion{1, 0, 0}, FeatureSupport{false, false, false, false, false, "legacy", false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.version)
			if got != tt.want {
				t.Fatalf("Evaluate(%s) = %+v, want %+v", tt.version, got, tt.want)
			}
		})
	}
}

func TestCapabilitiesOf(t *testing.T) {
	fs, err := CapabilitiesOf("onedrive v2.5.11")
	if err != nil {
		t.Fatal(err)
	}
	if !fs.DeviceAuth || fs.SchemaVersion != "2.5" {
		t.Fatalf("unexpected capabilities: %+v", fs)
	}
}
