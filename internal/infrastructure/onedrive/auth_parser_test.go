package onedrive

import "testing"

func TestParseAuthURL(t *testing.T) {
	p := &AuthParser{}
	samples := []string{
		`ATTENTION: Authorize this app visiting: https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=xxx&response_type=code&redirect_uri=...`,
		`OneDrive sync with a Microsoft account is being configured. Please visit this URL to authorize: https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize?scope=...`,
	}
	for _, s := range samples {
		url, ok := p.ParseAuthURL(s)
		if !ok || !stringsHasPrefix(url, "https://login.microsoftonline.com") {
			t.Fatalf("ParseAuthURL(%q) = %q, %v", s, url, ok)
		}
	}
}

func TestParseAuthURLNotFound(t *testing.T) {
	p := &AuthParser{}
	if _, ok := p.ParseAuthURL("Just a normal log line"); ok {
		t.Fatal("should not match normal line")
	}
}

func TestParseDeviceCode(t *testing.T) {
	p := &AuthParser{}
	code, ok := p.ParseDeviceCode("Use the following code: ABC123XYZ")
	if !ok || code != "ABC123XYZ" {
		t.Fatalf("device code = %q, %v", code, ok)
	}
}

func TestAuthSuccessFailure(t *testing.T) {
	p := &AuthParser{}
	if !p.IsAuthSuccess("Successfully authenticated with Microsoft") {
		t.Fatal("should detect success")
	}
	if !p.IsAuthFailure("Authorization failed. Invalid grant") {
		t.Fatal("should detect failure")
	}
	if p.IsAuthFailure("Successfully authenticated") {
		t.Fatal("success line should not be failure")
	}
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
