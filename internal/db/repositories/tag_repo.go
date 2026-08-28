package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Tag struct {
	ID   int64
	Name string
}

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) FindOrCreate(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}

	result, err := r.db.ExecContext(ctx, `INSERT INTO tags (name) VALUES (?)`, name)
	if err != nil {
		return 0, fmt.Errorf("insert tag: %w", err)
	}
	return result.LastInsertId()
}

func (r *TagRepository) AddToEntity(ctx context.Context, tagID int64, entityType string, entityID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO entity_tags (tag_id, entity_type, entity_id, created_at)
		VALUES (?, ?, ?, ?)
	`, tagID, entityType, entityID, time.Now())
	return err
}

func (r *TagRepository) RemoveFromEntity(ctx context.Context, tagID int64, entityType string, entityID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM entity_tags WHERE tag_id = ? AND entity_type = ? AND entity_id = ?
	`, tagID, entityType, entityID)
	return err
}

func (r *TagRepository) GetForEntity(ctx context.Context, entityType string, entityID int64) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name
		FROM tags t
		JOIN entity_tags et ON et.tag_id = t.id
		WHERE et.entity_type = ? AND et.entity_id = ?
		ORDER BY t.name
	`, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("get tags for entity: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *TagRepository) ListAll(ctx context.Context) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM tags ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
