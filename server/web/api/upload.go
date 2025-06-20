package api

import (
	"net/http"

	"log"
	set "server/settings"
	"server/torr"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
)

// torrentUpload godoc
//
//	@Summary		Add .torrent file
//	@Description	Only one file support.
//
//	@Tags			API
//
//	@Param			file	formData	file	true	"Torrent file to insert"
//	@Param			save	formData	string	false	"Save to DB"
//	@Param			title	formData	string	false	"Torrent title"
//	@Param			category	formData	string	false	"Torrent category"
//	@Param			poster	formData	string	false	"Torrent poster"
//	@Param			data	formData	string	false	"Torrent data"
//
//	@Accept			multipart/form-data
//
//	@Produce		json
//	@Success		200	{object}	state.TorrentStatus	"Torrent status"
//	@Router			/torrent/upload [post]
func torrentUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer form.RemoveAll()

	save := len(form.Value["save"]) > 0
	title := ""
	if len(form.Value["title"]) > 0 {
		title = form.Value["title"][0]
	}
	category := ""
	if len(form.Value["category"]) > 0 {
		category = form.Value["category"][0]
	}
	poster := ""
	if len(form.Value["poster"]) > 0 {
		poster = form.Value["poster"][0]
	}
	data := ""
	if len(form.Value["data"]) > 0 {
		data = form.Value["data"][0]
	}
	var tor *torr.Torrent
	for name, file := range form.File {
		log.Println("add .torrent", name)

		torrFile, err := file[0].Open()
		if err != nil {
			log.Println("error upload torrent:", err)
			continue
		}
		defer torrFile.Close()

		spec, err := utils.ParseFile(torrFile)
		if err != nil {
			log.Println("error upload torrent:", err)
			continue
		}

		tor, err = torr.AddTorrent(spec, title, poster, data, category)

		if tor.Data != "" && set.BTsets.EnableDebug {
			log.Println("torrent data:", tor.Data)
		}
		if tor.Category != "" && set.BTsets.EnableDebug {
			log.Println("torrent category:", tor.Category)
		}

		if err != nil {
			log.Println("error upload torrent:", err)
			continue
		}

		go func() {
			if !tor.GotInfo() {
				log.Println("error add torrent:", "timeout connection torrent")
				return
			}

			if tor.Title == "" {
				tor.Title = tor.Name()
			}

			if save {
				torr.SaveTorrentToDB(tor)
			}
		}()

		break
	}
	c.JSON(200, tor.Status())
}
