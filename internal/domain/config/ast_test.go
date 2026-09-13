package config

import "testing"

func TestSetExistingKey(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeKeyValue, Key: "threads", Value: "2", LineNumber: 3},
	}}
	doc.Set("threads", "12")
	n := doc.Get("threads")
	if n == nil || n.Value != "12" {
		t.Fatalf("Set should modify existing value, got %+v", n)
	}
	if n.LineNumber != 3 {
		t.Fatalf("Set must keep original line number, got %d", n.LineNumber)
	}
	if len(doc.Nodes) != 1 {
		t.Fatalf("Set on existing key must not append, got %d nodes", len(doc.Nodes))
	}
}

func TestSetNewKeyAppends(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeKeyValue, Key: "threads", Value: "2", LineNumber: 1},
	}}
	doc.Set("sync_dir", "~/OD")
	if len(doc.Nodes) != 2 {
		t.Fatalf("Set on new key must append, got %d nodes", len(doc.Nodes))
	}
	last := doc.Nodes[len(doc.Nodes)-1]
	if last.Key != "sync_dir" || last.Value != "~/OD" || last.Type != NodeKeyValue {
		t.Fatalf("appended node mismatch: %+v", last)
	}
}

func TestEnableDisabledKey(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeDisabledKeyValue, Key: "sync_dir", Value: "~/OD", LineNumber: 2},
	}}
	doc.Enable("sync_dir")
	n := doc.Get("sync_dir")
	if n.Type != NodeKeyValue {
		t.Fatalf("Enable should flip type, got %+v", n)
	}
	if n.Value != "~/OD" {
		t.Fatalf("Enable must preserve value, got %q", n.Value)
	}
}

func TestDisableKey(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeKeyValue, Key: "threads", Value: "8", LineNumber: 4},
	}}
	doc.Disable("threads")
	n := doc.Get("threads")
	if n.Type != NodeDisabledKeyValue {
		t.Fatalf("Disable should flip type, got %+v", n)
	}
}

func TestGetMissing(t *testing.T) {
	doc := &ConfigDocument{}
	if n := doc.Get("nope"); n != nil {
		t.Fatalf("Get on missing key should return nil, got %+v", n)
	}
}

func TestKeysExcludesDisabledAndComment(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeKeyValue, Key: "a", Value: "1"},
		{Type: NodeDisabledKeyValue, Key: "b", Value: "2"},
		{Type: NodeComment, Key: "c"},
		{Type: NodeBlankLine},
	}}
	keys := doc.Keys()
	if len(keys) != 1 || keys[0] != "a" {
		t.Fatalf("Keys() = %v, want [a]", keys)
	}
}

func TestRemoveConvertsToComment(t *testing.T) {
	doc := &ConfigDocument{Nodes: []ConfigNode{
		{Type: NodeKeyValue, Key: "rate_limit", Value: "1000", LineNumber: 5},
	}}
	doc.Remove("rate_limit")
	if n := doc.Get("rate_limit"); n != nil {
		t.Fatalf("Remove must not return the node via Get, got %+v", n)
	}
	if doc.Nodes[0].Type != NodeComment {
		t.Fatalf("Remove should convert to comment, got %+v", doc.Nodes[0])
	}
}

func TestEnableMissingNoop(t *testing.T) {
	doc := &ConfigDocument{}
	doc.Enable("nope")
	doc.Disable("nope")
	doc.Remove("nope")
	if len(doc.Nodes) != 0 {
		t.Fatalf("noop operations must not mutate document")
	}
}
