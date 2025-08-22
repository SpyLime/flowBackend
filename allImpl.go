package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// fetchClipTitle fetches the title for a YouTube clip
func fetchClipTitle(encodedClipURL string) (string, error) {
	// Decode URL if path-encoded
	clipURL, err := url.PathUnescape(encodedClipURL)
	if err != nil {
		return "", fmt.Errorf("invalid clip URL: %w", err)
	}

	// Build the request with browser-like headers
	req, err := http.NewRequest("GET", clipURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("YouTube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	// Extract title from <meta name="title" content="...">
	re := regexp.MustCompile(`<meta\s+name="title"\s+content="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) >= 2 {
		title := string(matches[1])
		return title, nil
	}

	return "", fmt.Errorf("could not find title in page")
}
