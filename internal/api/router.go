package api

import (
	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/curation"
	"github.com/musiclib/internal/library"
)

func NewRouter(svc *library.Service, curationSvc *curation.Service, lastfmSyncHandler *LastfmSyncHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware())

	h := NewHandler(svc)
	ch := NewCurationHandler(curationSvc)

	api := r.Group("/api")
	{
		api.GET("/summary", h.Summary)
		api.GET("/search", h.Search)

		api.GET("/artists", h.ListArtists)
		api.GET("/artists/:id", h.GetArtist)
		api.PATCH("/artists/:id", h.UpdateArtist)
		api.GET("/artists/:id/albums", h.GetArtistAlbums)
		api.GET("/artists/:id/tracks", h.GetArtistTracks)

		api.GET("/albums", h.ListAlbums)
		api.GET("/albums/:id", h.GetAlbum)
		api.PATCH("/albums/:id", h.UpdateAlbum)
		api.GET("/albums/:id/tracks", h.GetAlbumTracks)

		api.GET("/tracks", h.ListTracks)
		api.GET("/tracks/:id", h.GetTrack)
		api.PATCH("/tracks/:id", h.UpdateTrack)
		api.GET("/tracks/:id/history", h.GetTrackHistory)

		api.GET("/stats/:type/:id", h.GetStats)

		api.GET("/tags/:type/:id", h.GetTags)
		api.POST("/tags/:type/:id", h.AddTag)
		api.DELETE("/tags/:type/:id/:tagId", h.RemoveTag)

		api.POST("/sync/lastfm", lastfmSyncHandler.Start)

		api.GET("/collections", ch.ListRoot)
		api.POST("/collections", ch.Create)
		api.GET("/collections/tree", ch.Tree)
		api.GET("/collections/:id", ch.Get)
		api.PATCH("/collections/:id", ch.Update)
		api.DELETE("/collections/:id", ch.Delete)
		api.POST("/collections/:id/move", ch.Move)
		api.GET("/collections/:id/items", ch.ListItems)
		api.POST("/collections/:id/items", ch.AddItem)
		api.DELETE("/collections/:id/items/:itemId", ch.RemoveItem)
		api.PATCH("/collections/:id/items/:itemId", ch.UpdateItemNote)
		api.POST("/collections/:id/items/reorder", ch.ReorderItems)
	}

	return r
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				gin.DefaultErrorWriter.Write([]byte(e.Error() + "\n"))
			}
		}
	}
}
