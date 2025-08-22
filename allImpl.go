package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// fetchClipThumbnail fetches the thumbnail for a YouTube clip
func fetchClipThumbnail(encodedClipURL string) (ClipInfo string, err error) {
	// Decode the path-encoded clip URL
	clipURL, err := url.PathUnescape(encodedClipURL)
	if err != nil {
		return "", fmt.Errorf("invalid clip URL: %w", err)
	}

	return fetchClipThumbnailFromURL(clipURL)
}

// fetchClipThumbnailFromURL fetches thumbnail from the actual clip URL
func fetchClipThumbnailFromURL(clipURL string) (ClipInfo string, err error) {
	req, err := http.NewRequest("GET", clipURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Pretend to be a normal browser so YouTube gives the full HTML
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")

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

	playerRespJSON, err := extractYTInitialPlayerResponse(body)
	if err != nil {
		return "", err
	}

	var playerResp struct {
		VideoDetails struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
		} `json:"videoDetails"`
	}

	if err := json.Unmarshal(playerRespJSON, &playerResp); err != nil {
		return "", fmt.Errorf("failed to parse player response JSON: %w", err)
	}

	thumbs := playerResp.VideoDetails.Thumbnail.Thumbnails
	if len(thumbs) == 0 {
		return "", fmt.Errorf("no thumbnails found")
	}

	info := thumbs[len(thumbs)-1].URL // pick highest resolution
	return info, nil
}

// extractYTInitialPlayerResponse extracts the ytInitialPlayerResponse JSON from the HTML
func extractYTInitialPlayerResponse(body []byte) ([]byte, error) {
	// Match even if there's whitespace or semicolon after the JSON
	re := regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{.*?\})\s*;`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find ytInitialPlayerResponse")
	}
	return matches[1], nil
}
