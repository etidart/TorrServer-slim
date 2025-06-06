package server

import (
	"net"
	"os"
	"path/filepath"

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
	// remove old disk caches
	go cleanCache()
	// set settings http and https ports. Start web server.
	settings.Port = port
	settings.IP = ip

	web.Start()
}

func cleanCache() {
	if !settings.BTsets.UseDisk || settings.BTsets.TorrentsSavePath == "/" || settings.BTsets.TorrentsSavePath == "" {
		return
	}

	dirs, err := os.ReadDir(settings.BTsets.TorrentsSavePath)
	if err != nil {
		return
	}

	torrs := settings.ListTorrent()

	log.Println("Remove unused cache in dir:", settings.BTsets.TorrentsSavePath)
	keep := map[string]bool{}
	for _, d := range dirs {
		if len(d.Name()) != 40 {
			// Not a hash
			continue
		}

		if !settings.BTsets.RemoveCacheOnDrop {
			keep[d.Name()] = true
			for _, t := range torrs {
				if d.IsDir() && d.Name() == t.InfoHash.HexString() {
					keep[d.Name()] = false
					break
				}
			}
			for hash, del := range keep {
				if del && hash == d.Name() {
					log.Println("Remove unused cache:", d.Name())
					removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
				}
			}
		} else {
			if d.IsDir() {
				log.Println("Remove unused cache:", d.Name())
				removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
			}
		}
	}

}

func removeAllFiles(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, f := range files {
		name := filepath.Join(path, f.Name())
		os.Remove(name)
	}
	os.Remove(path)
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
