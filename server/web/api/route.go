package api

import "github.com/gin-gonic/gin"

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route gin.IRouter) {
	route.GET("/shutdown", shutdown)
	route.GET("/shutdown/*reason", shutdown)

	route.POST("/settings", settings)

	route.POST("/torrents", torrents)

	route.POST("/torrent/upload", torrentUpload)

	route.POST("/cache", cache)

	route.HEAD("/stream", stream)
	route.GET("/stream", stream)

	route.HEAD("/stream/*fname", stream)
	route.GET("/stream/*fname", stream)

	route.HEAD("/play/:hash/:id", play)
	route.GET("/play/:hash/:id", play)

	route.POST("/viewed", viewed)

	route.GET("/playlistall/all.m3u", allPlayList)

	route.GET("/playlist", playList)
	route.GET("/playlist/*fname", playList)

	route.GET("/download/:size", download)
}
