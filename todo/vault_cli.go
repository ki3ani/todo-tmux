package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func handleVaultSave() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vault save <url-or-text> [-t tags] [-p]")
		return
	}

	content := os.Args[2]
	tags := []string{}
	pinned := false

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-t", "--tags":
			if i+1 < len(os.Args) {
				tags = strings.Split(os.Args[i+1], ",")
				i++
			}
		case "-p", "--pin":
			pinned = true
		}
	}

	// Detect type
	contentType := DetectContentType(content)

	item := &VaultItem{
		ContentType: contentType,
		Content:     content,
		Pinned:      pinned,
	}

	// Set URL and fetch metadata for links
	if contentType != ContentTypeNote {
		item.URL = content
		fmt.Printf("Detected: %s\n", contentType)
		fmt.Println("Fetching metadata...")

		meta := FetchMetadata(content, contentType)
		if meta != nil {
			item.MetaTitle = meta.Title
			item.MetaDescription = meta.Description
			item.MetaThumbnail = meta.Thumbnail
			item.MetaAuthor = meta.Author
			item.MetaSiteName = meta.SiteName
		}
	}

	saved, err := CreateVaultItem(item, tags)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("\nSaved [%d] %s\n", saved.ID, saved.ContentType)
	if saved.MetaTitle != "" {
		fmt.Printf("  %s\n", saved.MetaTitle)
	}
	if saved.MetaAuthor != "" {
		fmt.Printf("  by %s\n", saved.MetaAuthor)
	}
	if len(tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	}
	if pinned {
		fmt.Println("  Pinned")
	}
}

func handleVaultNote() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vault note <text> [-t tags] [-p]")
		return
	}

	content := os.Args[2]
	tags := []string{}
	pinned := false

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-t", "--tags":
			if i+1 < len(os.Args) {
				tags = strings.Split(os.Args[i+1], ",")
				i++
			}
		case "-p", "--pin":
			pinned = true
		}
	}

	item := &VaultItem{
		ContentType: ContentTypeNote,
		Content:     content,
		Title:       content,
		Pinned:      pinned,
	}

	saved, err := CreateVaultItem(item, tags)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Saved note [%d]\n", saved.ID)
	if len(tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	}
}

func handleVaultList() {
	filter := VaultFilter{}

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-t", "--type":
			if i+1 < len(os.Args) {
				filter.ContentType = os.Args[i+1]
				i++
			}
		case "--tags":
			if i+1 < len(os.Args) {
				filter.TagNames = strings.Split(os.Args[i+1], ",")
				i++
			}
		case "-s", "--search":
			if i+1 < len(os.Args) {
				filter.Search = os.Args[i+1]
				i++
			}
		case "--pinned":
			p := true
			filter.Pinned = &p
		case "--archived":
			a := true
			filter.Archived = &a
		}
	}

	items, err := GetVaultItems(filter)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("No items in vault. Save something with: vault save <url>")
		return
	}

	fmt.Println()
	for _, item := range items {
		icon := getTypeIcon(item.ContentType)
		title := item.MetaTitle
		if title == "" {
			title = item.Title
		}
		if title == "" {
			title = truncateStr(item.Content, 50)
		}

		pin := ""
		if item.Pinned {
			pin = " [pinned]"
		}

		fmt.Printf("  %s %d. %s%s\n", icon, item.ID, title, pin)

		if len(item.Tags) > 0 {
			tagNames := make([]string, len(item.Tags))
			for i, t := range item.Tags {
				tagNames[i] = "#" + t.Name
			}
			fmt.Printf("     %s\n", strings.Join(tagNames, " "))
		}
	}
	fmt.Println()
}

func handleVaultRandom() {
	item, err := GetRandomVaultItem()
	if err != nil {
		fmt.Println("No items in vault to resurface")
		return
	}

	icon := getTypeIcon(item.ContentType)
	title := item.MetaTitle
	if title == "" {
		title = item.Title
	}
	if title == "" {
		title = truncateStr(item.Content, 80)
	}

	fmt.Println("\nFrom your vault:")
	fmt.Printf("  %s %s\n", icon, title)
	if item.MetaAuthor != "" {
		fmt.Printf("     by %s\n", item.MetaAuthor)
	}
	if item.URL != "" {
		fmt.Printf("     %s\n", item.URL)
	}
	if len(item.Tags) > 0 {
		tagNames := make([]string, len(item.Tags))
		for i, t := range item.Tags {
			tagNames[i] = "#" + t.Name
		}
		fmt.Printf("     %s\n", strings.Join(tagNames, " "))
	}
	fmt.Println()
}

func handleVaultPin(pin bool) {
	if len(os.Args) < 3 {
		if pin {
			fmt.Println("Usage: vault pin <id>")
		} else {
			fmt.Println("Usage: vault unpin <id>")
		}
		return
	}

	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	item, err := GetVaultItem(id)
	if err != nil {
		fmt.Println("Item not found")
		return
	}

	ToggleVaultItemPin(id, pin)
	if pin {
		fmt.Printf("Pinned [%d] %s\n", id, item.MetaTitle)
	} else {
		fmt.Printf("Unpinned [%d] %s\n", id, item.MetaTitle)
	}
}

func handleVaultArchive(archive bool) {
	if len(os.Args) < 3 {
		if archive {
			fmt.Println("Usage: vault archive <id>")
		} else {
			fmt.Println("Usage: vault unarchive <id>")
		}
		return
	}

	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	item, err := GetVaultItem(id)
	if err != nil {
		fmt.Println("Item not found")
		return
	}

	ToggleVaultItemArchive(id, archive)
	if archive {
		fmt.Printf("Archived [%d] %s\n", id, item.MetaTitle)
	} else {
		fmt.Printf("Unarchived [%d] %s\n", id, item.MetaTitle)
	}
}

func handleVaultDelete() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vault rm <id>")
		return
	}

	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	item, err := GetVaultItem(id)
	if err != nil {
		fmt.Println("Item not found")
		return
	}

	DeleteVaultItem(id)
	title := item.MetaTitle
	if title == "" {
		title = truncateStr(item.Content, 40)
	}
	fmt.Printf("Deleted [%d] %s\n", id, title)
}

func handleVaultTags() {
	tags, err := GetAllTags()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(tags) == 0 {
		fmt.Println("No tags yet")
		return
	}

	fmt.Println("\nTags:")
	for _, t := range tags {
		fmt.Printf("  #%s\n", t.Name)
	}
	fmt.Println()
}

func handleVaultSetTags() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: vault tag <id> <tag1,tag2,...>")
		return
	}

	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	tags := strings.Split(os.Args[3], ",")
	SetItemTags(id, tags)
	fmt.Printf("Updated tags for [%d]: %s\n", id, strings.Join(tags, ", "))
}

func getTypeIcon(t ContentType) string {
	switch t {
	case ContentTypeTweet:
		return "[X]"
	case ContentTypeTikTok:
		return "[TT]"
	case ContentTypeYouTube:
		return "[YT]"
	case ContentTypeArticle:
		return "[ART]"
	case ContentTypeNote:
		return "[NOTE]"
	default:
		return "[?]"
	}
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func printVaultUsage() {
	fmt.Println(`vault - personal content manager

Vault Commands:
  vault save <url> [-t tags] [-p]   Save a link (auto-detects type)
  vault note <text> [-t tags] [-p]  Save a quick note
  vault list [-t type] [--tags x]   List saved items
  vault random                      Resurface a random old item
  vault pin <id>                    Pin an item
  vault unpin <id>                  Unpin an item
  vault archive <id>                Archive an item
  vault rm <id>                     Delete an item
  vault tags                        List all tags
  vault tag <id> <tags>             Set tags for an item
  vault server                      Start web UI

Todo Commands (still work):
  vault add <task> [-p priority] [-c category] [-d due]
  vault list [-s status] [-p priority] [-c category]
  vault done <id>
  vault rm <id>

Examples:
  vault save "https://youtube.com/watch?v=..." -t music,favorites
  vault note "Great idea for app" -t ideas -p
  vault list --tags coding
  vault random`)
}
