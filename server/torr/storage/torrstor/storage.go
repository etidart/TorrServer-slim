package torrstor

import (
	"context"
	"sync"

	"server/torr/storage"

	"github.com/anacrolix/torrent/metainfo"
	ts "github.com/anacrolix/torrent/storage"
)

type Storage struct {
	storage.Storage

	caches   map[metainfo.Hash]*Cache
	capacity int64
	mu       sync.Mutex
}

func NewStorage(capacity int64) *Storage {
	stor := new(Storage)
	stor.capacity = capacity
	stor.caches = make(map[metainfo.Hash]*Cache)
	return stor
}

func (s *Storage) OpenTorrent(contx context.Context, info *metainfo.Info, infoHash metainfo.Hash) (ts.TorrentImpl, error) {
	capFunc := func() (int64, bool) { //	NE
	 	return s.capacity, true //	NE
	} //	NE
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := NewCache(s.capacity, s)
	ch.Init(info, infoHash)
	s.caches[infoHash] = ch
	// return ch, nil //	OE
	return ts.TorrentImpl{ //	NE
	 	Piece:    ch.Piece, //	NE
	 	Close:    ch.Close, //	NE
	 	Capacity: &capFunc, //	NE
	}, nil //	NE
}

func (s *Storage) CloseHash(hash metainfo.Hash) {
	if s.caches == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.caches[hash]; ok {
		ch.Close()
		delete(s.caches, hash)
	}
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.caches {
		ch.Close()
	}
	return nil
}

func (s *Storage) GetCache(hash metainfo.Hash) *Cache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cache, ok := s.caches[hash]; ok {
		return cache
	}
	return nil
}
