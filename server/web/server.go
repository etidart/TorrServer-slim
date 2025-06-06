package web

import (
	"net"
	"os"
	"sort"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/wlynxg/anet"

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
	ips := GetLocalIps()
	if len(ips) > 0 {
		log.Println("Local IPs:", ips)
	}
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
		log.Println("Start http server at", settings.IP+":"+settings.Port)
		waitChan <- route.Run(settings.IP + ":" + settings.Port)
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

func GetLocalIps() []string {
	ifaces, err := anet.Interfaces()
	if err != nil {
		log.Println("Error get local IPs")
		return nil
	}
	var list []string
	for _, i := range ifaces {
		addrs, _ := anet.InterfaceAddrsByInterface(&i)
		if i.Flags&net.FlagUp == net.FlagUp {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() {
					list = append(list, ip.String())
				}
			}
		}
	}
	sort.Strings(list)
	return list
}
