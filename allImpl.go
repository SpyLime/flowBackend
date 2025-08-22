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
	clipURL, err := url.PathUnescape(encodedClipURL)
	if err != nil {
		return "", fmt.Errorf("invalid clip URL: %w", err)
	}
	return fetchClipThumbnailFromURL(clipURL)
}

// fetchClipThumbnailFromURL fetches thumbnail from the actual clip URL
func fetchClipThumbnailFromURL(clipURL string) (ClipInfo string, err error) {
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
	fmt.Printf("Fetched %d bytes\n", len(body))

	playerRespJSON, err := extractYTInitialPlayerResponse(body)
	if err != nil {
		fmt.Println("Failed to extract ytInitialPlayerResponse")
		fmt.Println("Page snippet:\n", string(body[:1000]))
		return "", err
	}

	var playerResp map[string]interface{}
	if err := json.Unmarshal(playerRespJSON, &playerResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Try videoDetails.thumbnail
	if thumbs := getThumbnailsFromMap(playerResp, "videoDetails", "thumbnail", "thumbnails"); len(thumbs) > 0 {
		info := thumbs[len(thumbs)-1]["url"].(string)
		fmt.Printf("Thumbnail found in videoDetails: %s\n", info)
		return info, nil
	}

	// Try microformat.playerMicroformatRenderer.thumbnail
	if thumbs := getThumbnailsFromMap(playerResp, "microformat", "playerMicroformatRenderer", "thumbnail", "thumbnails"); len(thumbs) > 0 {
		info := thumbs[len(thumbs)-1]["url"].(string)
		fmt.Printf("Thumbnail found in microformat: %s\n", info)
		return info, nil
	}

	// Log full JSON for debugging if nothing found
	fmt.Println("No thumbnails found in JSON, logging JSON for debugging")
	fmt.Printf("%s\n", string(playerRespJSON))

	return "", fmt.Errorf("no thumbnails found")
}

// getThumbnailsFromMap safely navigates nested maps and returns []interface{} for "thumbnails" if found
func getThumbnailsFromMap(m map[string]interface{}, keys ...string) []map[string]interface{} {
	current := m
	for i, k := range keys {
		if val, ok := current[k]; ok {
			if i == len(keys)-1 {
				// Final key should be []interface{} of thumbnails
				if arr, ok := val.([]interface{}); ok {
					result := make([]map[string]interface{}, 0, len(arr))
					for _, item := range arr {
						if mitem, ok := item.(map[string]interface{}); ok {
							result = append(result, mitem)
						}
					}
					return result
				}
			} else if next, ok := val.(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return nil
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
