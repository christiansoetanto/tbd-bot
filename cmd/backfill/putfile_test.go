package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// --force skips the existence check and PUTs straight over the path. The
// GitHub Contents API rejects that without the current blob sha — the same
// `"sha" wasn't supplied` reply the daily cron already whitelists — so the
// flag failed in exactly the case it exists for.
func TestPutFilePayload_IncludesSHAWhenOverwriting(t *testing.T) {
	date := time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC)

	payload, err := putFilePayload("2026/2026-07-28.md", "body", date, "abc123")
	if err != nil {
		t.Fatalf("building payload: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("payload is not a JSON object: %v", err)
	}
	if got["sha"] != "abc123" {
		t.Errorf("overwriting an existing file must send the current sha, got %q", got["sha"])
	}
	if decoded, _ := base64.StdEncoding.DecodeString(got["content"]); string(decoded) != "body" {
		t.Errorf("content round trip failed, got %q", decoded)
	}
}

// Creating a new file must not send a sha at all: GitHub rejects a create
// carrying one, so an empty string cannot simply be passed through.
func TestPutFilePayload_OmitsSHAWhenCreating(t *testing.T) {
	date := time.Date(2026, time.July, 28, 0, 0, 0, 0, time.UTC)

	payload, err := putFilePayload("2026/2026-07-28.md", "body", date, "")
	if err != nil {
		t.Fatalf("building payload: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("payload is not a JSON object: %v", err)
	}
	if _, present := got["sha"]; present {
		t.Error("creating a new file must not send a sha key")
	}
	if got["message"] == "" {
		t.Error("commit message missing")
	}
}
