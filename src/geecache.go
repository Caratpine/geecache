package src

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once.")
	}
	g.peers = peers
}

func (g *Group) Get(key string) (bv ByteView, err error) {
	var ok bool
	if key == "" {
		err = fmt.Errorf("key is required")
		return
	}

	if bv, ok = g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return
	}
	return g.load(key)
}

func (g *Group) load(key string) (bv ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if bv, err = g.getFromPeer(peer, key); err == nil {
				return
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (bv ByteView, err error) {
	var (
		bs []byte
	)
	if bs, err = peer.Get(g.name, key); err != nil {
		return
	}
	bv.b = bs
	return
}

func (g *Group) getLocally(key string) (bv ByteView, err error) {
	var (
		bytes []byte
	)
	if bytes, err = g.getter.Get(key); err != nil {
		log.Printf("Group.getLocally(%s) error(%v)\n", key, err)
		return
	}
	bv = ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, bv)
	return
}

func (g *Group) populateCache(key string, bv ByteView) {
	g.mainCache.add(key, bv)
}
