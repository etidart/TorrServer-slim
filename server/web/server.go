package web

import (
	"net"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"

	"server/settings"

	"log"
	"server/torr"
	"server/version"
	"server/web/api"
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error)
)

//	@title			Swagger Torrserver API
//	@version		{version.Version}
//	@description	Torrent streaming server.

//	@license.name	GPL 3.0

//	@BasePath	/

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func Start() {
	log.Println("Start TorrServer " + version.Version + " torrent " + version.GetTorrentVersion())
	err := BTS.Connect()
	if err != nil {
		log.Println("BTS.Connect() error!", err) // waitChan <- err
		os.Exit(1)                              // return
	}

	gin.SetMode(gin.ReleaseMode)

	// corsCfg := cors.DefaultConfig()
	// corsCfg.AllowAllOrigins = true
	// corsCfg.AllowHeaders = []string{"*"}
	// corsCfg.AllowMethods = []string{"*"}
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowPrivateNetwork = true
	corsCfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "X-Requested-With", "Accept", "Authorization"}

	route := gin.New()
	route.Use(gin.Recovery(), cors.New(corsCfg), location.Default())

	route.GET("/echo", echo)

	api.SetupRoute(route)

	go func() {
		var l net.Listener
		if strings.HasPrefix(settings.LAddr, "unix:") {
			l, err = net.Listen("unix", settings.LAddr[5:])
		} else {
			l, err = net.Listen("tcp", settings.LAddr)
		}
		if err != nil {
			log.Fatalln("Failed to bind on", settings.LAddr)
		}
		log.Println("Start http server at", settings.LAddr)
		waitChan <- route.RunListener(l)
	}()
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	BTS.Disconnect()
	waitChan <- nil
}

// echo godoc
//
//	@Summary		Tests server status
//	@Description	Tests whether server is alive or not
//
//	@Tags			API
//
//	@Produce		plain
//	@Success		200	{string}	string	"Server version"
//	@Router			/echo [get]
func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}
