package settings

import (
	"os"
	"path/filepath"

	"log"
)

var (
	tdb      TorrServerDB
	Path     string
	LAddr    string
	ReadOnly bool
	PubIPv4  string
	PubIPv6  string
	TorAddr  string
	MaxSize  int64
)

func InitSets(readOnly bool) {
	ReadOnly = readOnly

	bboltDB := NewTDB()
	if bboltDB == nil {
		log.Println("Error open bboltDB:", filepath.Join(Path, "config.db"))
		os.Exit(1)
	}

	jsonDB := NewJsonDB()
	if jsonDB == nil {
		log.Println("Error open jsonDB")
		os.Exit(1)
	}

	dbRouter := NewXPathDBRouter()
	// First registered DB becomes default route
	dbRouter.RegisterRoute(jsonDB, "Settings")
	dbRouter.RegisterRoute(jsonDB, "Viewed")
	dbRouter.RegisterRoute(bboltDB, "Torrents")

	tdb = NewDBReadCache(dbRouter)

	// We migrate settings here, it must be done before loadBTSets()
	if err := MigrateToJson(bboltDB, jsonDB); err != nil {
		log.Println("MigrateToJson failed")
		os.Exit(1)
	}
	loadBTSets()
	MigrateTorrents()
}

func CloseDB() {
	tdb.CloseDB()
}
