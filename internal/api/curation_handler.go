package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/curation"
	"github.com/musiclib/internal/db/repositories"
)

type CurationHandler struct {
	svc *curation.Service
}

func NewCurationHandler(svc *curation.Service) *CurationHandler {
	return &CurationHandler{svc: svc}
}

func (h *CurationHandler) ListRoot(c *gin.Context) {
	collections, err := h.svc.ListRootCollections(c.Request.Context())
	if err != nil {
		slog.Error("list root collections", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list collections"})
		return
	}
	if collections == nil {
		collections = []curation.CollectionSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"items": collections})
}

func (h *CurationHandler) Create(c *gin.Context) {
	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		ParentID    *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	col, err := h.svc.CreateCollection(c.Request.Context(), body.Name, body.Description, body.ParentID)
	if err != nil {
		slog.Error("create collection", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, col)
}

func (h *CurationHandler) Get(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	col, err := h.svc.GetCollection(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, curation.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		} else {
			slog.Error("get collection", "id", id, "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get collection"})
		}
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *CurationHandler) Update(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := h.svc.UpdateCollection(c.Request.Context(), id, body.Name, body.Description); err != nil {
		slog.Error("update collection", "id", id, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) Delete(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	if err := h.svc.DeleteCollection(c.Request.Context(), id); err != nil {
		slog.Error("delete collection", "id", id, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) Move(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		ParentID *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.MoveCollection(c.Request.Context(), id, body.ParentID); err != nil {
		switch {
		case errors.Is(err, curation.ErrCycleDetected):
			c.JSON(http.StatusBadRequest, gin.H{"error": "moving collection would create a cycle"})
		case errors.Is(err, curation.ErrCannotContainSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "collection cannot contain itself"})
		case errors.Is(err, curation.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		default:
			slog.Error("move collection", "id", id, "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to move collection"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) ListItems(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	col, err := h.svc.GetCollection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": col.Items})
}

func (h *CurationHandler) AddItem(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		ItemType string `json:"item_type" binding:"required"`
		ItemID   int64  `json:"item_id" binding:"required"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_type and item_id are required"})
		return
	}

	item, err := h.svc.AddItem(c.Request.Context(), id, body.ItemType, body.ItemID, body.Note)
	if err != nil {
		switch {
		case errors.Is(err, curation.ErrInvalidItemType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item type"})
		case errors.Is(err, curation.ErrInvalidItemRef):
			c.JSON(http.StatusBadRequest, gin.H{"error": "referenced entity does not exist"})
		case errors.Is(err, curation.ErrCycleDetected):
			c.JSON(http.StatusBadRequest, gin.H{"error": "adding this collection would create a cycle"})
		case errors.Is(err, curation.ErrCannotContainSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "collection cannot contain itself"})
		case errors.Is(err, curation.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		default:
			slog.Error("add item to collection", "collection_id", id, "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item"})
		}
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *CurationHandler) RemoveItem(c *gin.Context) {
	_, err := parseCurationID(c, "id")
	if err != nil {
		return
	}
	itemID, err := parseCurationID(c, "itemId")
	if err != nil {
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), itemID); err != nil {
		slog.Error("remove item from collection", "item_id", itemID, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) UpdateItemNote(c *gin.Context) {
	_, err := parseCurationID(c, "id")
	if err != nil {
		return
	}
	itemID, err := parseCurationID(c, "itemId")
	if err != nil {
		return
	}

	var body struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateItemNote(c.Request.Context(), itemID, body.Note); err != nil {
		slog.Error("update item note", "item_id", itemID, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) ReorderItems(c *gin.Context) {
	id, err := parseCurationID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Items []repositories.ItemPosition `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items array is required"})
		return
	}

	if err := h.svc.ReorderItems(c.Request.Context(), id, body.Items); err != nil {
		slog.Error("reorder items", "collection_id", id, "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *CurationHandler) Tree(c *gin.Context) {
	tree, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		slog.Error("get collection tree", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tree"})
		return
	}
	if tree == nil {
		tree = []curation.CollectionTree{}
	}
	c.JSON(http.StatusOK, gin.H{"items": tree})
}

func parseCurationID(c *gin.Context, param string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, err
	}
	return id, nil
}
