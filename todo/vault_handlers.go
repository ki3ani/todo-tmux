package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func handleAPIVault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		filter := VaultFilter{
			ContentType: r.URL.Query().Get("type"),
			Search:      r.URL.Query().Get("search"),
		}

		if tags := r.URL.Query().Get("tags"); tags != "" {
			filter.TagNames = strings.Split(tags, ",")
		}

		if pinned := r.URL.Query().Get("pinned"); pinned == "true" {
			p := true
			filter.Pinned = &p
		}

		if archived := r.URL.Query().Get("archived"); archived == "true" {
			a := true
			filter.Archived = &a
		}

		items, err := GetVaultItems(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if items == nil {
			items = []VaultItem{}
		}
		json.NewEncoder(w).Encode(items)

	case "POST":
		var input struct {
			Content string   `json:"content"`
			Title   string   `json:"title"`
			Tags    []string `json:"tags"`
			Pinned  bool     `json:"pinned"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(input.Content) == "" {
			http.Error(w, "Content is required", http.StatusBadRequest)
			return
		}

		// Detect content type
		contentType := DetectContentType(input.Content)

		item := &VaultItem{
			ContentType: contentType,
			Title:       input.Title,
			Content:     input.Content,
			Pinned:      input.Pinned,
		}

		// Set URL if it's a link type
		if contentType != ContentTypeNote {
			item.URL = input.Content
			// Fetch metadata
			meta := FetchMetadata(input.Content, contentType)
			if meta != nil {
				item.MetaTitle = meta.Title
				item.MetaDescription = meta.Description
				item.MetaThumbnail = meta.Thumbnail
				item.MetaAuthor = meta.Author
				item.MetaSiteName = meta.SiteName
			}
		}

		saved, err := CreateVaultItem(item, input.Tags)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(saved)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAPIVaultItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path: /api/vault/123
	path := strings.TrimPrefix(r.URL.Path, "/api/vault/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		item, err := GetVaultItem(id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(item)

	case "PUT":
		var input struct {
			Title    string   `json:"title"`
			Content  string   `json:"content"`
			Pinned   bool     `json:"pinned"`
			Archived bool     `json:"archived"`
			Tags     []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		existing, err := GetVaultItem(id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		existing.Title = input.Title
		existing.Content = input.Content
		existing.Pinned = input.Pinned
		existing.Archived = input.Archived

		if err := UpdateVaultItem(id, existing); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if input.Tags != nil {
			SetItemTags(id, input.Tags)
		}

		item, _ := GetVaultItem(id)
		json.NewEncoder(w).Encode(item)

	case "PATCH":
		var input struct {
			Pinned   *bool `json:"pinned"`
			Archived *bool `json:"archived"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if input.Pinned != nil {
			ToggleVaultItemPin(id, *input.Pinned)
		}
		if input.Archived != nil {
			ToggleVaultItemArchive(id, *input.Archived)
		}

		item, _ := GetVaultItem(id)
		json.NewEncoder(w).Encode(item)

	case "DELETE":
		if err := DeleteVaultItem(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAPIVaultResurface(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	item, err := GetRandomVaultItem()
	if err != nil {
		http.Error(w, "No items to resurface", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(item)
}

func handleAPIVaultDetect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	contentType := DetectContentType(input.Content)

	result := map[string]interface{}{
		"content_type": contentType,
	}

	// Fetch metadata preview if it's a URL
	if contentType != ContentTypeNote {
		meta := FetchMetadata(input.Content, contentType)
		if meta != nil {
			result["meta_title"] = meta.Title
			result["meta_description"] = meta.Description
			result["meta_thumbnail"] = meta.Thumbnail
			result["meta_author"] = meta.Author
			result["meta_site_name"] = meta.SiteName
		}
	}

	json.NewEncoder(w).Encode(result)
}

func handleAPITags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		tags, err := GetAllTags()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if tags == nil {
			tags = []Tag{}
		}
		json.NewEncoder(w).Encode(tags)

	case "POST":
		var input struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		tag, err := GetOrCreateTag(input.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tag)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
