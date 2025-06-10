package server

import (
	"server/settings"
	"server/web"
)

func Start(laddr string, roSets bool) {
	settings.InitSets(roSets)
	settings.LAddr = laddr
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
