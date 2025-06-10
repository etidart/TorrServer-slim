package server

import (
	"net"
	"os"

	"log"
	"server/settings"
	"server/web"
)

func Start(port, ip string, roSets bool) {
	settings.InitSets(roSets)

	if port == "" {
		port = "8090"
	}

	log.Println("Check web port", port)
	l, err := net.Listen("tcp", ip+":"+port)
	if l != nil {
		l.Close()
	}
	if err != nil {
		log.Println("Port", port, "already in use! Please set different port for HTTP. Abort")
		os.Exit(1)
	}

	// set settings http and https ports. Start web server.
	settings.Port = port
	settings.IP = ip

	web.Start()
}

func WaitServer() string {
	err := web.Wait()
	if err != nil {
		return err.Error()
	}
	return ""
}

func Stop() {
	web.Stop()
	settings.CloseDB()
}
