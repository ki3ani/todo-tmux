package main

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	twitterPattern = regexp.MustCompile(`(?i)^https?://(?:www\.)?(twitter\.com|x\.com)/\w+/status/\d+`)
	tiktokPattern  = regexp.MustCompile(`(?i)^https?://(?:www\.|vm\.)?tiktok\.com/`)
	youtubePattern = regexp.MustCompile(`(?i)^https?://(?:www\.)?(youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)`)
)

// DetectContentType determines the content type from input
func DetectContentType(input string) ContentType {
	input = strings.TrimSpace(input)

	// Check if it's a URL
	parsedURL, err := url.Parse(input)
	if err != nil || parsedURL.Scheme == "" {
		return ContentTypeNote
	}

	// Match against known patterns
	switch {
	case twitterPattern.MatchString(input):
		return ContentTypeTweet
	case tiktokPattern.MatchString(input):
		return ContentTypeTikTok
	case youtubePattern.MatchString(input):
		return ContentTypeYouTube
	default:
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			return ContentTypeArticle
		}
		return ContentTypeNote
	}
}

// ExtractYouTubeID extracts video ID from YouTube URLs
func ExtractYouTubeID(urlStr string) string {
	// Handle youtube.com/watch?v=ID
	if strings.Contains(urlStr, "youtube.com/watch") {
		u, _ := url.Parse(urlStr)
		return u.Query().Get("v")
	}
	// Handle youtu.be/ID
	if strings.Contains(urlStr, "youtu.be/") {
		parts := strings.Split(urlStr, "youtu.be/")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "?")[0]
			return strings.TrimSuffix(id, "/")
		}
	}
	// Handle youtube.com/shorts/ID
	if strings.Contains(urlStr, "/shorts/") {
		parts := strings.Split(urlStr, "/shorts/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "?")[0]
		}
	}
	return ""
}

// ExtractTweetID extracts tweet ID from Twitter/X URLs
func ExtractTweetID(urlStr string) string {
	matches := regexp.MustCompile(`/status/(\d+)`).FindStringSubmatch(urlStr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
