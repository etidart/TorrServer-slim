package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/alexflint/go-arg"

	"server"
	"server/settings"
	"server/torr"
	"server/version"
)

type args struct {
	Port        string `arg:"-p" help:"web server port (default 8090)"`
	IP          string `arg:"-i" help:"web server addr (default empty)"`
	Path        string `arg:"-d" help:"database and config dir path"`
	RDB         bool   `arg:"-r" help:"start in read-only DB mode"`
	TorrentsDir string `arg:"-t" help:"autoload torrents from dir"`
	TorrentAddr string `help:"Torrent client address, like 127.0.0.1:1337 (default :PeersListenPort)"`
	PubIPv4     string `arg:"-4" help:"set public IPv4 addr"`
	PubIPv6     string `arg:"-6" help:"set public IPv6 addr"`
	MaxSize     string `arg:"-m" help:"max allowed stream size (in Bytes)"`
}

func (args) Version() string {
	return "TorrServer " + version.Version
}

var params args

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	arg.MustParse(&params)

	if params.Path == "" {
		params.Path, _ = os.Getwd()
	}

	if params.Port == "" {
		params.Port = "8090"
	}

	settings.Path = params.Path
	fmt.Println("=========== START ===========")
	fmt.Println("TorrServer", version.Version+",", runtime.Version()+",", "CPU Num:", runtime.NumCPU())

	if params.RDB {
		log.Println("Running in Read-only DB mode!")
	}

	dnsResolve()

	if params.TorrentAddr != "" {
		settings.TorAddr = params.TorrentAddr
	}

	if params.PubIPv4 != "" {
		settings.PubIPv4 = params.PubIPv4
	}

	if params.PubIPv6 != "" {
		settings.PubIPv6 = params.PubIPv6
	}

	if params.TorrentsDir != "" {
		go watchTDir(params.TorrentsDir)
	}

	if params.MaxSize != "" {
		maxSize, err := strconv.ParseInt(params.MaxSize, 10, 64)
		if err == nil {
			settings.MaxSize = maxSize
		}
	}

	server.Start(params.Port, params.IP, params.RDB)
	log.Println(server.WaitServer())
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func dnsResolve() {
	addrs, err := net.LookupHost("www.google.com")
	if len(addrs) == 0 {
		log.Println("Check dns failed", addrs, err)

		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		}

		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}

		addrs, err = net.LookupHost("www.google.com")
		log.Println("Check cloudflare dns", addrs, err)
	} else {
		log.Println("Check dns OK", addrs, err)
	}
}

func watchTDir(dir string) {
	time.Sleep(5 * time.Second)
	path, err := filepath.Abs(dir)
	if err != nil {
		path = dir
	}
	for {
		files, err := os.ReadDir(path)
		if err == nil {
			for _, file := range files {
				filename := filepath.Join(path, file.Name())
				if strings.ToLower(filepath.Ext(file.Name())) == ".torrent" {
					sp, err := openFile(filename)
					if err == nil {
						tor, err := torr.AddTorrent(sp, "", "", "", "")
						if err == nil {
							if tor.GotInfo() {
								if tor.Title == "" {
									tor.Title = tor.Name()
								}
								torr.SaveTorrentToDB(tor)
								tor.Drop()
								os.Remove(filename)
								time.Sleep(time.Second)
							} else {
								log.Println("Error get info from torrent")
							}
						} else {
							log.Println("Error parse torrent file:", err)
						}
					} else {
						log.Println("Error parse file name:", err)
					}
				}
			}
		} else {
			log.Println("Error read dir:", err)
		}
		time.Sleep(time.Second * 5)
	}
}

func openFile(path string) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	mag, err := minfo.MagnetV2()
	if err != nil {
		return nil, err
	}

	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: mag.DisplayName,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}
