package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CollectionRow struct {
	ID          int64
	ParentID    sql.NullInt64
	Name        string
	Description string
	CreatedAt   int64
	UpdatedAt   int64
}

type CollectionItemRow struct {
	ID           int64
	CollectionID int64
	ItemType     string
	ItemID       int64
	Position     int
	Note         string
	CreatedAt    int64
}

type CollectionItemWithName struct {
	CollectionItemRow
	Name      string
	ArtistName sql.NullString
	AlbumName  sql.NullString
}

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

func (r *CollectionRepository) Create(ctx context.Context, name, description string, parentID *int64) (int64, error) {
	now := time.Now().Unix()
	var pid sql.NullInt64
	if parentID != nil {
		pid = sql.NullInt64{Int64: *parentID, Valid: true}
	}
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO collections (parent_id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		pid, name, description, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create collection: %w", err)
	}
	return result.LastInsertId()
}

func (r *CollectionRepository) GetByID(ctx context.Context, id int64) (*CollectionRow, error) {
	var c CollectionRow
	err := r.db.QueryRowContext(ctx,
		`SELECT id, parent_id, name, description, created_at, updated_at FROM collections WHERE id = ?`, id,
	).Scan(&c.ID, &c.ParentID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}
	return &c, nil
}

func (r *CollectionRepository) ListRoot(ctx context.Context) ([]CollectionRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, parent_id, name, description, created_at, updated_at FROM collections WHERE parent_id IS NULL ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list root collections: %w", err)
	}
	defer rows.Close()
	return r.scanCollections(rows)
}

func (r *CollectionRepository) ListChildren(ctx context.Context, parentID int64) ([]CollectionRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, parent_id, name, description, created_at, updated_at FROM collections WHERE parent_id = ? ORDER BY name ASC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("list child collections: %w", err)
	}
	defer rows.Close()
	return r.scanCollections(rows)
}

func (r *CollectionRepository) Update(ctx context.Context, id int64, name, description string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx,
		`UPDATE collections SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, now, id,
	)
	if err != nil {
		return fmt.Errorf("update collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) Move(ctx context.Context, id int64, parentID *int64) error {
	now := time.Now().Unix()
	var pid sql.NullInt64
	if parentID != nil {
		pid = sql.NullInt64{Int64: *parentID, Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE collections SET parent_id = ?, updated_at = ? WHERE id = ?`,
		pid, now, id,
	)
	if err != nil {
		return fmt.Errorf("move collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	return nil
}

func (r *CollectionRepository) GetAncestorIDs(ctx context.Context, collectionID int64) ([]int64, error) {
	var ids []int64
	current := collectionID
	for {
		var parentID sql.NullInt64
		err := r.db.QueryRowContext(ctx,
			`SELECT parent_id FROM collections WHERE id = ?`, current,
		).Scan(&parentID)
		if err != nil {
			return nil, fmt.Errorf("get ancestors: %w", err)
		}
		if !parentID.Valid {
			break
		}
		ids = append(ids, parentID.Int64)
		current = parentID.Int64
	}
	return ids, nil
}

func (r *CollectionRepository) AddItem(ctx context.Context, collectionID int64, itemType string, itemID int64, note string) (int64, error) {
	now := time.Now().Unix()
	var maxPos int
	_ = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(position), -1) FROM collection_items WHERE collection_id = ?`, collectionID,
	).Scan(&maxPos)

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO collection_items (collection_id, item_type, item_id, position, note, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		collectionID, itemType, itemID, maxPos+1, note, now,
	)
	if err != nil {
		return 0, fmt.Errorf("add collection item: %w", err)
	}
	return result.LastInsertId()
}

func (r *CollectionRepository) RemoveItem(ctx context.Context, itemID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collection_items WHERE id = ?`, itemID)
	if err != nil {
		return fmt.Errorf("remove collection item: %w", err)
	}
	return nil
}

func (r *CollectionRepository) GetItem(ctx context.Context, itemID int64) (*CollectionItemRow, error) {
	var ci CollectionItemRow
	err := r.db.QueryRowContext(ctx,
		`SELECT id, collection_id, item_type, item_id, position, note, created_at FROM collection_items WHERE id = ?`, itemID,
	).Scan(&ci.ID, &ci.CollectionID, &ci.ItemType, &ci.ItemID, &ci.Position, &ci.Note, &ci.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get collection item: %w", err)
	}
	return &ci, nil
}

func (r *CollectionRepository) UpdateItemNote(ctx context.Context, itemID int64, note string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE collection_items SET note = ? WHERE id = ?`, note, itemID,
	)
	if err != nil {
		return fmt.Errorf("update item note: %w", err)
	}
	return nil
}

func (r *CollectionRepository) ListItems(ctx context.Context, collectionID int64) ([]CollectionItemWithName, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			ci.id, ci.collection_id, ci.item_type, ci.item_id, ci.position, ci.note, ci.created_at,
			CASE ci.item_type
				WHEN 'artist' THEN (SELECT name FROM artists WHERE id = ci.item_id)
				WHEN 'album' THEN (SELECT name FROM albums WHERE id = ci.item_id)
				WHEN 'track' THEN (SELECT name FROM tracks WHERE id = ci.item_id)
				WHEN 'collection' THEN (SELECT name FROM collections WHERE id = ci.item_id)
			END as name
		FROM collection_items ci
		WHERE ci.collection_id = ?
		ORDER BY ci.position ASC
	`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("list collection items: %w", err)
	}
	defer rows.Close()

	var items []CollectionItemWithName
	for rows.Next() {
		var ci CollectionItemWithName
		if err := rows.Scan(
			&ci.ID, &ci.CollectionID, &ci.ItemType, &ci.ItemID, &ci.Position, &ci.Note, &ci.CreatedAt,
			&ci.Name,
		); err != nil {
			return nil, fmt.Errorf("scan collection item: %w", err)
		}
		items = append(items, ci)
	}
	return items, rows.Err()
}

func (r *CollectionRepository) ReorderItems(ctx context.Context, collectionID int64, itemPositions []ItemPosition) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE collection_items SET position = ? WHERE id = ? AND collection_id = ?`)
	if err != nil {
		return fmt.Errorf("prepare reorder: %w", err)
	}
	defer stmt.Close()

	for _, ip := range itemPositions {
		if _, err := stmt.ExecContext(ctx, ip.Position, ip.ItemID, collectionID); err != nil {
			return fmt.Errorf("reorder item %d: %w", ip.ItemID, err)
		}
	}

	return tx.Commit()
}

type ItemPosition struct {
	ItemID   int64 `json:"id"`
	Position int   `json:"position"`
}

func (r *CollectionRepository) ItemExists(ctx context.Context, itemType string, itemID int64) (bool, error) {
	var exists bool
	var query string
	switch itemType {
	case "artist":
		query = `SELECT EXISTS(SELECT 1 FROM artists WHERE id = ?)`
	case "album":
		query = `SELECT EXISTS(SELECT 1 FROM albums WHERE id = ?)`
	case "track":
		query = `SELECT EXISTS(SELECT 1 FROM tracks WHERE id = ?)`
	case "collection":
		query = `SELECT EXISTS(SELECT 1 FROM collections WHERE id = ?)`
	default:
		return false, nil
	}
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(&exists)
	return exists, err
}

func (r *CollectionRepository) HasChildCollections(ctx context.Context, collectionID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM collections WHERE parent_id = ?`, collectionID,
	).Scan(&count)
	return count > 0, err
}

func (r *CollectionRepository) IsDescendant(ctx context.Context, candidateID, ancestorID int64) (bool, error) {
	current := candidateID
	for {
		var parentID sql.NullInt64
		err := r.db.QueryRowContext(ctx,
			`SELECT parent_id FROM collections WHERE id = ?`, current,
		).Scan(&parentID)
		if err != nil {
			return false, fmt.Errorf("check descendant: %w", err)
		}
		if !parentID.Valid {
			return false, nil
		}
		if parentID.Int64 == ancestorID {
			return true, nil
		}
		current = parentID.Int64
	}
}

func (r *CollectionRepository) TotalCollections(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM collections`).Scan(&count)
	return count, err
}

func (r *CollectionRepository) GetItemName(ctx context.Context, itemType string, itemID int64) (string, error) {
	var name string
	var query string
	switch itemType {
	case "artist":
		query = `SELECT name FROM artists WHERE id = ?`
	case "album":
		query = `SELECT name FROM albums WHERE id = ?`
	case "track":
		query = `SELECT name FROM tracks WHERE id = ?`
	case "collection":
		query = `SELECT name FROM collections WHERE id = ?`
	default:
		return "", fmt.Errorf("unknown item type: %s", itemType)
	}
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(&name)
	return name, err
}

func (r *CollectionRepository) scanCollections(rows *sql.Rows) ([]CollectionRow, error) {
	var collections []CollectionRow
	for rows.Next() {
		var c CollectionRow
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		collections = append(collections, c)
	}
	return collections, rows.Err()
}
