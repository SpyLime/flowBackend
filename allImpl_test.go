package main

import (
	"testing"
)

func TestFetchClipThumbnail(t *testing.T) {
	clipURL := "https://youtube.com/clip/UgkxbG8R2t72qb0l7FWcaBW9r-hU3kGPf-ms"
	expected := "https://i.ytimg.com/vi/tX6yZBJ1n0M/maxresdefault.jpg"

	info, err := fetchClipThumbnail(clipURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info != expected {
		t.Fatalf("got %q, want %q", info, expected)
	}

	t.Logf("Thumbnail URL: %s", info)
}
