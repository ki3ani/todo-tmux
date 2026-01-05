package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// InitVaultDB creates vault tables
func InitVaultDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS vault_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content_type TEXT NOT NULL,
		title TEXT DEFAULT '',
		content TEXT NOT NULL,
		url TEXT DEFAULT '',
		meta_title TEXT DEFAULT '',
		meta_description TEXT DEFAULT '',
		meta_thumbnail TEXT DEFAULT '',
		meta_author TEXT DEFAULT '',
		meta_site_name TEXT DEFAULT '',
		pinned BOOLEAN DEFAULT FALSE,
		archived BOOLEAN DEFAULT FALSE,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		color TEXT DEFAULT '#8892b0',
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS item_tags (
		item_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (item_id, tag_id),
		FOREIGN KEY (item_id) REFERENCES vault_items(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_vault_content_type ON vault_items(content_type);
	CREATE INDEX IF NOT EXISTS idx_vault_pinned ON vault_items(pinned);
	CREATE INDEX IF NOT EXISTS idx_vault_archived ON vault_items(archived);
	`
	_, err := db.Exec(schema)
	return err
}

// CreateVaultItem creates a new vault item with optional tags
func CreateVaultItem(item *VaultItem, tagNames []string) (*VaultItem, error) {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO vault_items (
			content_type, title, content, url,
			meta_title, meta_description, meta_thumbnail,
			meta_author, meta_site_name,
			pinned, archived, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ContentType, item.Title, item.Content, item.URL,
		item.MetaTitle, item.MetaDescription, item.MetaThumbnail,
		item.MetaAuthor, item.MetaSiteName,
		item.Pinned, item.Archived,
		now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	item.ID = id
	item.CreatedAt = now
	item.UpdatedAt = now

	// Add tags
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		tag, err := GetOrCreateTag(tagName)
		if err != nil {
			continue
		}
		AddTagToItem(id, tag.ID)
		item.Tags = append(item.Tags, *tag)
	}

	return item, nil
}

// GetVaultItems retrieves items with filtering
func GetVaultItems(filter VaultFilter) ([]VaultItem, error) {
	query := `SELECT DISTINCT vi.id, vi.content_type, vi.title, vi.content, vi.url,
		vi.meta_title, vi.meta_description, vi.meta_thumbnail, vi.meta_author, vi.meta_site_name,
		vi.pinned, vi.archived, vi.created_at, vi.updated_at
		FROM vault_items vi`
	args := []interface{}{}
	where := []string{"1=1"}

	// Tag filtering requires join
	if len(filter.TagNames) > 0 {
		query += " LEFT JOIN item_tags it ON vi.id = it.item_id LEFT JOIN tags t ON it.tag_id = t.id"
	}

	if filter.ContentType != "" {
		where = append(where, "vi.content_type = ?")
		args = append(args, filter.ContentType)
	}

	if filter.Pinned != nil {
		where = append(where, "vi.pinned = ?")
		args = append(args, *filter.Pinned)
	}

	if filter.Archived != nil {
		where = append(where, "vi.archived = ?")
		args = append(args, *filter.Archived)
	} else {
		where = append(where, "vi.archived = FALSE")
	}

	if len(filter.TagNames) > 0 {
		placeholders := make([]string, len(filter.TagNames))
		for i, name := range filter.TagNames {
			placeholders[i] = "?"
			args = append(args, strings.ToLower(strings.TrimSpace(name)))
		}
		where = append(where, fmt.Sprintf("LOWER(t.name) IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.Search != "" {
		where = append(where, "(vi.title LIKE ? OR vi.content LIKE ? OR vi.meta_title LIKE ? OR vi.meta_description LIKE ?)")
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	query += " WHERE " + strings.Join(where, " AND ")
	query += " ORDER BY vi.pinned DESC, vi.created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []VaultItem
	for rows.Next() {
		item, err := scanVaultItem(rows)
		if err != nil {
			continue
		}
		item.Tags, _ = GetTagsForItem(item.ID)
		items = append(items, *item)
	}
	return items, nil
}

// GetVaultItem gets a single item by ID
func GetVaultItem(id int64) (*VaultItem, error) {
	row := db.QueryRow(`SELECT id, content_type, title, content, url,
		meta_title, meta_description, meta_thumbnail, meta_author, meta_site_name,
		pinned, archived, created_at, updated_at
		FROM vault_items WHERE id = ?`, id)

	item, err := scanVaultItemRow(row)
	if err != nil {
		return nil, err
	}
	item.Tags, _ = GetTagsForItem(item.ID)
	return item, nil
}

// UpdateVaultItem updates an existing item
func UpdateVaultItem(id int64, item *VaultItem) error {
	now := time.Now()
	_, err := db.Exec(`UPDATE vault_items SET
		title=?, content=?, url=?,
		meta_title=?, meta_description=?, meta_thumbnail=?,
		meta_author=?, meta_site_name=?,
		pinned=?, archived=?, updated_at=?
		WHERE id=?`,
		item.Title, item.Content, item.URL,
		item.MetaTitle, item.MetaDescription, item.MetaThumbnail,
		item.MetaAuthor, item.MetaSiteName,
		item.Pinned, item.Archived, now.Format(time.RFC3339), id,
	)
	return err
}

// ToggleVaultItemPin toggles the pinned status
func ToggleVaultItemPin(id int64, pinned bool) error {
	now := time.Now()
	_, err := db.Exec(`UPDATE vault_items SET pinned=?, updated_at=? WHERE id=?`,
		pinned, now.Format(time.RFC3339), id)
	return err
}

// ToggleVaultItemArchive toggles the archived status
func ToggleVaultItemArchive(id int64, archived bool) error {
	now := time.Now()
	_, err := db.Exec(`UPDATE vault_items SET archived=?, updated_at=? WHERE id=?`,
		archived, now.Format(time.RFC3339), id)
	return err
}

// DeleteVaultItem deletes an item
func DeleteVaultItem(id int64) error {
	_, err := db.Exec(`DELETE FROM vault_items WHERE id=?`, id)
	return err
}

// GetRandomVaultItem returns a random non-archived item for resurfacing
func GetRandomVaultItem() (*VaultItem, error) {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM vault_items WHERE archived = FALSE`).Scan(&count)
	if count == 0 {
		return nil, sql.ErrNoRows
	}

	offset := rand.Intn(count)
	row := db.QueryRow(`SELECT id, content_type, title, content, url,
		meta_title, meta_description, meta_thumbnail, meta_author, meta_site_name,
		pinned, archived, created_at, updated_at
		FROM vault_items WHERE archived = FALSE
		LIMIT 1 OFFSET ?`, offset)

	item, err := scanVaultItemRow(row)
	if err != nil {
		return nil, err
	}
	item.Tags, _ = GetTagsForItem(item.ID)
	return item, nil
}

// Tag operations
func GetOrCreateTag(name string) (*Tag, error) {
	name = strings.TrimSpace(strings.ToLower(name))

	var tag Tag
	var createdAt string
	err := db.QueryRow(`SELECT id, name, color, created_at FROM tags WHERE LOWER(name) = ?`, name).
		Scan(&tag.ID, &tag.Name, &tag.Color, &createdAt)

	if err == sql.ErrNoRows {
		now := time.Now()
		result, err := db.Exec(`INSERT INTO tags (name, created_at) VALUES (?, ?)`,
			name, now.Format(time.RFC3339))
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		return &Tag{ID: id, Name: name, Color: "#8892b0", CreatedAt: now}, nil
	}
	if err != nil {
		return nil, err
	}
	tag.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &tag, nil
}

func GetAllTags() ([]Tag, error) {
	rows, err := db.Query(`SELECT id, name, color, created_at FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		var createdAt string
		rows.Scan(&t.ID, &t.Name, &t.Color, &createdAt)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tags = append(tags, t)
	}
	return tags, nil
}

func GetTagsForItem(itemID int64) ([]Tag, error) {
	rows, err := db.Query(`
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		var createdAt string
		rows.Scan(&t.ID, &t.Name, &t.Color, &createdAt)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tags = append(tags, t)
	}
	return tags, nil
}

func AddTagToItem(itemID, tagID int64) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO item_tags (item_id, tag_id) VALUES (?, ?)`,
		itemID, tagID)
	return err
}

func SetItemTags(itemID int64, tagNames []string) error {
	db.Exec(`DELETE FROM item_tags WHERE item_id = ?`, itemID)

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tag, err := GetOrCreateTag(name)
		if err != nil {
			continue
		}
		AddTagToItem(itemID, tag.ID)
	}
	return nil
}

// Helper scan functions
func scanVaultItem(rows *sql.Rows) (*VaultItem, error) {
	var item VaultItem
	var createdAt, updatedAt string
	var contentType string
	err := rows.Scan(
		&item.ID, &contentType, &item.Title, &item.Content, &item.URL,
		&item.MetaTitle, &item.MetaDescription, &item.MetaThumbnail,
		&item.MetaAuthor, &item.MetaSiteName,
		&item.Pinned, &item.Archived,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.ContentType = ContentType(contentType)
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &item, nil
}

func scanVaultItemRow(row *sql.Row) (*VaultItem, error) {
	var item VaultItem
	var createdAt, updatedAt string
	var contentType string
	err := row.Scan(
		&item.ID, &contentType, &item.Title, &item.Content, &item.URL,
		&item.MetaTitle, &item.MetaDescription, &item.MetaThumbnail,
		&item.MetaAuthor, &item.MetaSiteName,
		&item.Pinned, &item.Archived,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.ContentType = ContentType(contentType)
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &item, nil
}
