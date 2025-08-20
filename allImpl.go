package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// ClipInfo holds the thumbnail URL and optional start time
type ClipInfo struct {
	ThumbnailURL string
	StartSeconds int64 // optional
}

// fetchClipThumbnail scrapes a YouTube clip page and returns the thumbnail URL
func fetchClipThumbnail(clipURL string) (*ClipInfo, error) {
	resp, err := http.Get(clipURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clip page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Extract the ytInitialPlayerResponse JSON
	re := regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{.*?\});`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find ytInitialPlayerResponse")
	}

	var playerResp struct {
		VideoDetails struct {
			Thumbnail struct {
				Thumbnails []struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"thumbnail"`
		} `json:"videoDetails"`
		Microformat struct {
			PlayerMicroformatRenderer struct {
				StartTimestamp string `json:"startTimestamp"` // optional
			} `json:"playerMicroformatRenderer"`
		} `json:"microformat"`
	}

	if err := json.Unmarshal(matches[1], &playerResp); err != nil {
		return nil, fmt.Errorf("failed to parse player response JSON: %w", err)
	}

	thumbs := playerResp.VideoDetails.Thumbnail.Thumbnails
	if len(thumbs) == 0 {
		return nil, fmt.Errorf("no thumbnails found")
	}

	info := &ClipInfo{
		ThumbnailURL: thumbs[len(thumbs)-1].URL, // pick highest resolution
	}

	return info, nil
}
