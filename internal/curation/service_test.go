package curation

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/musiclib/internal/db/migrations"
	"github.com/musiclib/internal/db/repositories"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Run(db); err != nil {
		t.Fatal(err)
	}

	db.Exec(`INSERT INTO artists (name, note, favorite, created_at, updated_at) VALUES ('Radiohead', '', 0, 0, 0)`)
	db.Exec(`INSERT INTO artists (name, note, favorite, created_at, updated_at) VALUES ('Lata Mangeshkar', '', 0, 0, 0)`)
	db.Exec(`INSERT INTO artists (name, note, favorite, created_at, updated_at) VALUES ('AC Lyrics', '', 0, 0, 0)`)

	db.Exec(`INSERT INTO albums (artist_id, name, note, favorite, created_at, updated_at) VALUES (1, 'OK Computer', '', 0, 0, 0)`)
	db.Exec(`INSERT INTO albums (artist_id, name, note, favorite, created_at, updated_at) VALUES (1, 'Kid A', '', 0, 0, 0)`)

	db.Exec(`INSERT INTO tracks (artist_id, name, note, favorite, created_at, updated_at) VALUES (1, 'Paranoid Android', '', 0, 0, 0)`)
	db.Exec(`INSERT INTO tracks (artist_id, name, note, favorite, created_at, updated_at) VALUES (1, 'Everything In Its Right Place', '', 0, 0, 0)`)

	return db
}

func newTestService(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	repo := repositories.NewCollectionRepository(db)
	return NewService(repo), db
}

func seedCollection(t *testing.T, svc *Service, name string, parentID *int64) *Collection {
	t.Helper()
	col, err := svc.CreateCollection(context.Background(), name, "", parentID)
	if err != nil {
		t.Fatalf("seed collection %q: %v", name, err)
	}
	return col
}

func TestCreateCollection(t *testing.T) {
	svc, _ := newTestService(t)
	col, err := svc.CreateCollection(context.Background(), "My Canon", "The best music", nil)
	if err != nil {
		t.Fatal(err)
	}
	if col.Name != "My Canon" {
		t.Errorf("expected name My Canon, got %s", col.Name)
	}
	if col.Description != "The best music" {
		t.Errorf("expected description, got %s", col.Description)
	}
	if col.ParentID != nil {
		t.Errorf("expected nil parent_id, got %v", col.ParentID)
	}
}

func TestCreateNested(t *testing.T) {
	svc, _ := newTestService(t)
	root := seedCollection(t, svc, "Root", nil)
	child := seedCollection(t, svc, "Child", &root.ID)
	if child.ParentID == nil || *child.ParentID != root.ID {
		t.Errorf("expected parent_id %d, got %v", root.ID, child.ParentID)
	}
}

func TestAddArtist(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	item, err := svc.AddItem(context.Background(), col.ID, "artist", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if item.ItemType != "artist" || item.ItemID != 1 {
		t.Errorf("expected artist/1, got %s/%d", item.ItemType, item.ItemID)
	}
	if item.Name != "Radiohead" {
		t.Errorf("expected Radiohead, got %s", item.Name)
	}
}

func TestAddAlbum(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	item, err := svc.AddItem(context.Background(), col.ID, "album", 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "OK Computer" {
		t.Errorf("expected OK Computer, got %s", item.Name)
	}
}

func TestAddTrack(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	item, err := svc.AddItem(context.Background(), col.ID, "track", 1, "Great song")
	if err != nil {
		t.Fatal(err)
	}
	if item.Note != "Great song" {
		t.Errorf("expected note, got %s", item.Note)
	}
}

func TestAddCollection(t *testing.T) {
	svc, _ := newTestService(t)
	parent := seedCollection(t, svc, "Parent", nil)
	child := seedCollection(t, svc, "Child", nil)
	item, err := svc.AddItem(context.Background(), parent.ID, "collection", child.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "Child" {
		t.Errorf("expected Child, got %s", item.Name)
	}
}

func TestInvalidItemID(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	_, err := svc.AddItem(context.Background(), col.ID, "artist", 999, "")
	if err != ErrInvalidItemRef {
		t.Errorf("expected ErrInvalidItemRef, got %v", err)
	}
}

func TestInvalidItemType(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	_, err := svc.AddItem(context.Background(), col.ID, "invalid", 1, "")
	if err != ErrInvalidItemType {
		t.Errorf("expected ErrInvalidItemType, got %v", err)
	}
}

func TestCyclePreventionMove(t *testing.T) {
	svc, _ := newTestService(t)
	a := seedCollection(t, svc, "A", nil)
	b := seedCollection(t, svc, "B", &a.ID)
	c := seedCollection(t, svc, "C", &b.ID)

	err := svc.MoveCollection(context.Background(), a.ID, &c.ID)
	if err != ErrCycleDetected {
		t.Errorf("expected ErrCycleDetected, got %v", err)
	}
}

func TestCyclePreventionAddItem(t *testing.T) {
	svc, _ := newTestService(t)
	a := seedCollection(t, svc, "A", nil)
	b := seedCollection(t, svc, "B", &a.ID)
	c := seedCollection(t, svc, "C", &b.ID)

	_, err := svc.AddItem(context.Background(), c.ID, "collection", a.ID, "")
	if err != ErrCycleDetected {
		t.Errorf("expected ErrCycleDetected, got %v", err)
	}
}

func TestCannotContainSelf(t *testing.T) {
	svc, _ := newTestService(t)
	a := seedCollection(t, svc, "A", nil)
	err := svc.MoveCollection(context.Background(), a.ID, &a.ID)
	if err != ErrCannotContainSelf {
		t.Errorf("expected ErrCannotContainSelf, got %v", err)
	}
}

func TestMoveCollection(t *testing.T) {
	svc, _ := newTestService(t)
	a := seedCollection(t, svc, "A", nil)
	b := seedCollection(t, svc, "B", nil)
	c := seedCollection(t, svc, "C", &a.ID)

	err := svc.MoveCollection(context.Background(), c.ID, &b.ID)
	if err != nil {
		t.Fatal(err)
	}
	col, _ := svc.GetCollection(context.Background(), c.ID)
	if col.ParentID == nil || *col.ParentID != b.ID {
		t.Errorf("expected parent_id %d, got %v", b.ID, col.ParentID)
	}
}

func TestDeleteCollection(t *testing.T) {
	svc, db := newTestService(t)
	col := seedCollection(t, svc, "ToDelete", nil)
	svc.AddItem(context.Background(), col.ID, "artist", 1, "")

	err := svc.DeleteCollection(context.Background(), col.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetCollection(context.Background(), col.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM artists WHERE id = 1`).Scan(&count)
	if count != 1 {
		t.Error("deleting collection should not delete library entities")
	}
}

func TestReorder(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	i1, _ := svc.AddItem(context.Background(), col.ID, "artist", 1, "")
	i2, _ := svc.AddItem(context.Background(), col.ID, "album", 1, "")
	i3, _ := svc.AddItem(context.Background(), col.ID, "track", 1, "")

	err := svc.ReorderItems(context.Background(), col.ID, []repositories.ItemPosition{
		{ItemID: i3.ID, Position: 0},
		{ItemID: i1.ID, Position: 1},
		{ItemID: i2.ID, Position: 2},
	})
	if err != nil {
		t.Fatal(err)
	}

	col2, _ := svc.GetCollection(context.Background(), col.ID)
	if col2.Items[0].ItemType != "track" {
		t.Errorf("expected track first, got %s", col2.Items[0].ItemType)
	}
	if col2.Items[1].ItemType != "artist" {
		t.Errorf("expected artist second, got %s", col2.Items[1].ItemType)
	}
}

func TestTree(t *testing.T) {
	svc, _ := newTestService(t)
	a := seedCollection(t, svc, "A", nil)
	seedCollection(t, svc, "B", &a.ID)
	seedCollection(t, svc, "C", nil)

	tree, err := svc.GetTree(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 2 {
		t.Errorf("expected 2 root nodes, got %d", len(tree))
	}
}

func TestRemoveItem(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	item, _ := svc.AddItem(context.Background(), col.ID, "artist", 1, "")

	err := svc.RemoveItem(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}

	col2, _ := svc.GetCollection(context.Background(), col.ID)
	if len(col2.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col2.Items))
	}
}

func TestUpdateItemNote(t *testing.T) {
	svc, _ := newTestService(t)
	col := seedCollection(t, svc, "Test", nil)
	item, _ := svc.AddItem(context.Background(), col.ID, "artist", 1, "old")

	err := svc.UpdateItemNote(context.Background(), item.ID, "new note")
	if err != nil {
		t.Fatal(err)
	}

	col2, _ := svc.GetCollection(context.Background(), col.ID)
	if col2.Items[0].Note != "new note" {
		t.Errorf("expected new note, got %s", col2.Items[0].Note)
	}
}
