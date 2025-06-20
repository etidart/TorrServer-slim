package settings

import (
	"path/filepath"
	"strings"
	"time"

	"log"

	bolt "go.etcd.io/bbolt"
)

type TDB struct {
	Path string
	db   *bolt.DB
}

func NewTDB() TorrServerDB {
	db, err := bolt.Open(filepath.Join(Path, "config.db"), 0o666, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		log.Println(err)
		return nil
	}

	tdb := new(TDB)
	tdb.db = db
	tdb.Path = Path
	return tdb
}

func (v *TDB) CloseDB() {
	if v.db != nil {
		v.db.Close()
		v.db = nil
	}
}

func (v *TDB) Get(xpath, name string) []byte {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return nil
	}
	var ret []byte
	err := v.db.View(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		ret = buckt.Get([]byte(name))
		return nil
	})
	if err != nil {
		log.Println("Error get sets", xpath+"/"+name, ", error:", err)
	}

	return ret
}

func (v *TDB) Set(xpath, name string, value []byte) {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return
	}
	err := v.db.Update(func(tx *bolt.Tx) error {
		buckt, err := tx.CreateBucketIfNotExists([]byte(spath[0]))
		if err != nil {
			return err
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt, err = buckt.CreateBucketIfNotExists([]byte(p))
			if err != nil {
				return err
			}
		}

		return buckt.Put([]byte(name), value)
	})
	if err != nil {
		log.Println("Error put sets", xpath+"/"+name, ", error:", err)
		log.Println("value:", value)
	}
}

func (v *TDB) List(xpath string) []string {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return nil
	}
	var ret []string
	err := v.db.View(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		buckt.ForEach(func(k, _ []byte) error {
			if len(k) > 0 {
				ret = append(ret, string(k))
			}
			return nil
		})

		return nil
	})
	if err != nil {
		log.Println("Error list sets", xpath, ", error:", err)
	}

	return ret
}

func (v *TDB) Rem(xpath, name string) {
	spath := strings.Split(xpath, "/")
	if len(spath) == 0 {
		return
	}
	err := v.db.Update(func(tx *bolt.Tx) error {
		buckt := tx.Bucket([]byte(spath[0]))
		if buckt == nil {
			return nil
		}

		for i, p := range spath {
			if i == 0 {
				continue
			}
			buckt = buckt.Bucket([]byte(p))
			if buckt == nil {
				return nil
			}
		}

		return buckt.Delete([]byte(name))
	})
	if err != nil {
		log.Println("Error rem sets", xpath+"/"+name, ", error:", err)
	}
}
