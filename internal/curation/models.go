package curation

type Collection struct {
	ID          int64          `json:"id"`
	ParentID    *int64         `json:"parent_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	Items       []CollectionItem `json:"items"`
}

type CollectionItem struct {
	ID       int64  `json:"id"`
	ItemType string `json:"item_type"`
	ItemID   int64  `json:"item_id"`
	Position int    `json:"position"`
	Note     string `json:"note"`
	Name     string `json:"name"`
}

type CollectionSummary struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parent_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ChildCount  int    `json:"child_count"`
	ItemCount   int    `json:"item_count"`
}

type CollectionTree struct {
	ID       int64            `json:"id"`
	Name     string           `json:"name"`
	Children []CollectionTree `json:"children"`
}
