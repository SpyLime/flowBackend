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
	clipURL, err := url.PathUnescape(encodedClipURL)
	if err != nil {
		return "", fmt.Errorf("invalid clip URL: %w", err)
	}
	return fetchClipTitleFromURL(clipURL)
}

// fetchClipTitleFromURL fetches the clip page and extracts the title
func fetchClipTitleFromURL(clipURL string) (string, error) {
	fmt.Printf("Fetching YouTube clip page: %s\n", clipURL)

	req, err := http.NewRequest("GET", clipURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Browser-like headers
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch clip page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("YouTube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	html := string(body)

	// Try <meta name="title" content="...">
	re := regexp.MustCompile(`(?i)<meta\s+name=["']title["']\s+content=["'](.*?)["']`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := matches[1]
		fmt.Printf("Title found via meta tag: %s\n", title)
		return title, nil
	}

	// Fallback: extract ytInitialPlayerResponse.videoDetails.title if present
	playerJSON, err := extractYTInitialPlayerResponse(body)
	if err == nil {
		title, err := extractTitleFromJSON(playerJSON)
		if err == nil {
			fmt.Printf("Title found via ytInitialPlayerResponse: %s\n", title)
			return title, nil
		}
	}

	return "", fmt.Errorf("could not find clip title")
}

// extractYTInitialPlayerResponse extracts the ytInitialPlayerResponse JSON from the HTML
func extractYTInitialPlayerResponse(body []byte) ([]byte, error) {
	re := regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{.*\})\s*;</script>`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find ytInitialPlayerResponse")
	}
	return matches[1], nil
}

// extractTitleFromJSON extracts videoDetails.title from ytInitialPlayerResponse JSON
func extractTitleFromJSON(data []byte) (string, error) {
	re := regexp.MustCompile(`"videoDetails"\s*:\s*\{.*?"title"\s*:\s*"(.*?)"`)
	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return "", fmt.Errorf("title not found in JSON")
	}
	return string(matches[1]), nil
}
