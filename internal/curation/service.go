package curation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/musiclib/internal/db/repositories"
)

var validItemTypes = map[string]bool{
	"artist": true, "album": true, "track": true, "collection": true,
}

type Service struct {
	repo *repositories.CollectionRepository
}

func NewService(repo *repositories.CollectionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCollection(ctx context.Context, name, description string, parentID *int64) (*Collection, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if parentID != nil {
		exists, err := s.repo.ItemExists(ctx, "collection", *parentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}

	id, err := s.repo.Create(ctx, name, description, parentID)
	if err != nil {
		return nil, err
	}

	slog.Info("collection created", "id", id, "name", name)
	return s.GetCollection(ctx, id)
}

func (s *Service) GetCollection(ctx context.Context, id int64) (*Collection, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	items, err := s.repo.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}

	c := collectionFromRow(row)
	c.Items = make([]CollectionItem, len(items))
	for i, item := range items {
		c.Items[i] = CollectionItem{
			ID:       item.ID,
			ItemType: item.ItemType,
			ItemID:   item.ItemID,
			Position: item.Position,
			Note:     item.Note,
			Name:     item.Name,
		}
	}
	return c, nil
}

func (s *Service) ListRootCollections(ctx context.Context) ([]CollectionSummary, error) {
	rows, err := s.repo.ListRoot(ctx)
	if err != nil {
		return nil, err
	}
	return s.summarizeCollections(ctx, rows)
}

func (s *Service) ListChildren(ctx context.Context, parentID int64) ([]CollectionSummary, error) {
	rows, err := s.repo.ListChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}
	return s.summarizeCollections(ctx, rows)
}

func (s *Service) UpdateCollection(ctx context.Context, id int64, name, description string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	slog.Info("collection updated", "id", id)
	return s.repo.Update(ctx, id, name, description)
}

func (s *Service) DeleteCollection(ctx context.Context, id int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	slog.Info("collection deleted", "id", id)
	return s.repo.Delete(ctx, id)
}

func (s *Service) MoveCollection(ctx context.Context, id int64, newParentID *int64) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}

	if newParentID != nil {
		if *newParentID == id {
			return ErrCannotContainSelf
		}

		exists, err := s.repo.ItemExists(ctx, "collection", *newParentID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}

		descendant, err := s.repo.IsDescendant(ctx, *newParentID, id)
		if err != nil {
			return err
		}
		if descendant {
			slog.Warn("cycle detected in collection move", "id", id, "target", *newParentID)
			return ErrCycleDetected
		}
	}

	slog.Info("collection moved", "id", id, "new_parent", newParentID)
	return s.repo.Move(ctx, id, newParentID)
}

func (s *Service) AddItem(ctx context.Context, collectionID int64, itemType string, itemID int64, note string) (*CollectionItem, error) {
	if !validItemTypes[itemType] {
		return nil, ErrInvalidItemType
	}

	_, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, ErrNotFound
	}

	if itemType == "collection" {
		if itemID == collectionID {
			return nil, ErrCannotContainSelf
		}
		exists, err := s.repo.ItemExists(ctx, "collection", itemID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}

		descendant, err := s.repo.IsDescendant(ctx, collectionID, itemID)
		if err != nil {
			return nil, err
		}
		if descendant {
			slog.Warn("cycle detected adding collection", "collection_id", collectionID, "item_id", itemID)
			return nil, ErrCycleDetected
		}
	} else {
		exists, err := s.repo.ItemExists(ctx, itemType, itemID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidItemRef
		}
	}

	itemIDdb, err := s.repo.AddItem(ctx, collectionID, itemType, itemID, note)
	if err != nil {
		return nil, err
	}

	item, err := s.repo.GetItem(ctx, itemIDdb)
	if err != nil {
		return nil, err
	}

	name, _ := s.repo.GetItemName(ctx, itemType, itemID)

	slog.Info("item added to collection", "collection_id", collectionID, "item_type", itemType, "item_id", itemID)
	return &CollectionItem{
		ID:       item.ID,
		ItemType: item.ItemType,
		ItemID:   item.ItemID,
		Position: item.Position,
		Note:     item.Note,
		Name:     name,
	}, nil
}

func (s *Service) RemoveItem(ctx context.Context, itemID int64) error {
	item, err := s.repo.GetItem(ctx, itemID)
	if err != nil {
		return ErrItemNotFound
	}
	slog.Info("item removed from collection", "collection_id", item.CollectionID, "item_id", itemID)
	return s.repo.RemoveItem(ctx, itemID)
}

func (s *Service) UpdateItemNote(ctx context.Context, itemID int64, note string) error {
	_, err := s.repo.GetItem(ctx, itemID)
	if err != nil {
		return ErrItemNotFound
	}
	return s.repo.UpdateItemNote(ctx, itemID, note)
}

func (s *Service) ReorderItems(ctx context.Context, collectionID int64, positions []repositories.ItemPosition) error {
	_, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return ErrNotFound
	}
	slog.Info("items reordered", "collection_id", collectionID, "count", len(positions))
	return s.repo.ReorderItems(ctx, collectionID, positions)
}

func (s *Service) GetTree(ctx context.Context) ([]CollectionTree, error) {
	rows, err := s.repo.ListRoot(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildTree(ctx, rows)
}

func (s *Service) buildTree(ctx context.Context, rows []repositories.CollectionRow) ([]CollectionTree, error) {
	tree := make([]CollectionTree, 0, len(rows))
	for _, row := range rows {
		node := CollectionTree{
			ID:   row.ID,
			Name: row.Name,
		}
		children, err := s.repo.ListChildren(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			node.Children, err = s.buildTree(ctx, children)
			if err != nil {
				return nil, err
			}
		}
		tree = append(tree, node)
	}
	return tree, nil
}

func (s *Service) TotalCollections(ctx context.Context) (int, error) {
	return s.repo.TotalCollections(ctx)
}

func (s *Service) summarizeCollections(ctx context.Context, rows []repositories.CollectionRow) ([]CollectionSummary, error) {
	summaries := make([]CollectionSummary, 0, len(rows))
	for _, row := range rows {
		cs := CollectionSummary{
			ID:          row.ID,
			ParentID:    rowParentID(&row),
			Name:        row.Name,
			Description: row.Description,
		}
		children, err := s.repo.ListChildren(ctx, row.ID)
		if err == nil {
			cs.ChildCount = len(children)
		}
		items, err := s.repo.ListItems(ctx, row.ID)
		if err == nil {
			cs.ItemCount = len(items)
		}
		summaries = append(summaries, cs)
	}
	return summaries, nil
}

func collectionFromRow(row *repositories.CollectionRow) *Collection {
	return &Collection{
		ID:          row.ID,
		ParentID:    rowParentID(row),
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func rowParentID(row *repositories.CollectionRow) *int64 {
	if row.ParentID.Valid {
		v := row.ParentID.Int64
		return &v
	}
	return nil
}

func nowUnix() int64 {
	return time.Now().Unix()
}
