package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type URLMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Author      string `json:"author"`
	SiteName    string `json:"site_name"`
}

// FetchMetadata fetches metadata for a URL based on content type
func FetchMetadata(urlStr string, contentType ContentType) *URLMetadata {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch contentType {
	case ContentTypeYouTube:
		return fetchYouTubeMetadata(ctx, urlStr)
	case ContentTypeTikTok:
		return fetchTikTokMetadata(ctx, urlStr)
	case ContentTypeTweet:
		return fetchTwitterMetadata(ctx, urlStr)
	case ContentTypeArticle:
		return fetchGenericMetadata(ctx, urlStr)
	default:
		return &URLMetadata{}
	}
}

// fetchYouTubeMetadata uses YouTube oEmbed API (no API key needed)
func fetchYouTubeMetadata(ctx context.Context, urlStr string) *URLMetadata {
	oembedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json", url.QueryEscape(urlStr))

	req, _ := http.NewRequestWithContext(ctx, "GET", oembedURL, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fetchGenericMetadata(ctx, urlStr)
	}
	defer resp.Body.Close()

	var data struct {
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	return &URLMetadata{
		Title:     data.Title,
		Author:    data.AuthorName,
		Thumbnail: data.ThumbnailURL,
		SiteName:  "YouTube",
	}
}

// fetchTikTokMetadata uses TikTok oEmbed API
func fetchTikTokMetadata(ctx context.Context, urlStr string) *URLMetadata {
	oembedURL := fmt.Sprintf("https://www.tiktok.com/oembed?url=%s", url.QueryEscape(urlStr))

	req, _ := http.NewRequestWithContext(ctx, "GET", oembedURL, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fetchGenericMetadata(ctx, urlStr)
	}
	defer resp.Body.Close()

	var data struct {
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	return &URLMetadata{
		Title:     data.Title,
		Author:    data.AuthorName,
		Thumbnail: data.ThumbnailURL,
		SiteName:  "TikTok",
	}
}

// fetchTwitterMetadata - Twitter limits API access, use syndication endpoint
func fetchTwitterMetadata(ctx context.Context, urlStr string) *URLMetadata {
	tweetID := ExtractTweetID(urlStr)
	if tweetID == "" {
		return &URLMetadata{Title: "Tweet", SiteName: "Twitter/X"}
	}

	// Try syndication API
	syndicationURL := fmt.Sprintf("https://cdn.syndication.twimg.com/tweet-result?id=%s&token=a", tweetID)
	req, _ := http.NewRequestWithContext(ctx, "GET", syndicationURL, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &URLMetadata{Title: "Tweet", SiteName: "Twitter/X"}
	}
	defer resp.Body.Close()

	var data struct {
		Text string `json:"text"`
		User struct {
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"user"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	title := data.Text
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	author := ""
	if data.User.ScreenName != "" {
		author = "@" + data.User.ScreenName
	}

	return &URLMetadata{
		Title:       title,
		Description: data.Text,
		Author:      author,
		SiteName:    "Twitter/X",
	}
}

// fetchGenericMetadata scrapes Open Graph meta tags
func fetchGenericMetadata(ctx context.Context, urlStr string) *URLMetadata {
	req, _ := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Vault/1.0)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &URLMetadata{}
	}
	defer resp.Body.Close()

	// Read first 100KB only
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	html := string(body)

	meta := &URLMetadata{}

	// Extract Open Graph tags
	meta.Title = extractMetaContent(html, "og:title")
	if meta.Title == "" {
		meta.Title = extractTitle(html)
	}
	meta.Description = extractMetaContent(html, "og:description")
	if meta.Description == "" {
		meta.Description = extractMetaContent(html, "description")
	}
	meta.Thumbnail = extractMetaContent(html, "og:image")
	meta.SiteName = extractMetaContent(html, "og:site_name")
	meta.Author = extractMetaContent(html, "author")

	return meta
}

func extractMetaContent(html, property string) string {
	patterns := []string{
		fmt.Sprintf(`<meta[^>]+property="%s"[^>]+content="([^"]*)"`, property),
		fmt.Sprintf(`<meta[^>]+content="([^"]*)"[^>]+property="%s"`, property),
		fmt.Sprintf(`<meta[^>]+name="%s"[^>]+content="([^"]*)"`, property),
		fmt.Sprintf(`<meta[^>]+content="([^"]*)"[^>]+name="%s"`, property),
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(html); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func extractTitle(html string) string {
	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if matches := re.FindStringSubmatch(html); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
