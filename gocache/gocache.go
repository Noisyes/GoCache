package gocache

import (
	"GoCache/singlefilght"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte,error)
}

type GetterFunc func(key string) ([]byte,error)

func (f GetterFunc) Get(key string)([]byte,error){
	return f(key)
}

type Group struct{
	name string //命名空间
	getter Getter //没有对应缓存时的回调函数
	mainCache cache // 并发缓存
	peers PeerPicker

	loader *singlefilght.Group
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group{
	if getter == nil{
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:name,
		getter: getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader: &singlefilght.Group{},
	}
	groups[name] = g
	return g
}

func (g *Group) RegisterPeers(peers PeerPicker){
	if g.peers !=nil{
		panic("regeisterPeerPicker called more than once")
	}
	g.peers = peers
}

func GetGroup(name string)*Group{
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string)(ByteView,error){
	if key ==""{
		return ByteView{},fmt.Errorf("key is required")
	}
	if v,ok := g.mainCache.get(key);ok{
		log.Println("[GoCache] hit")
		return v,nil
	}
	return g.load(key)
}

//缓存不存在，调用该方法查询关键字
func (g *Group) load(key string)(value ByteView,err error){
	data,err := g.loader.Do(key,func()(interface{},error){
		if g.peers!=nil{
			if peer,ok:= g.peers.PickPeer(key);ok{
				if value,err:= g.getFromPeer(peer,key);err==nil{
					return value,nil
				}
				log.Println("[GoCache] Failed to get from peer",err)
			}
		}
		return g.getLocally(key)
	})
	if err==nil{
		return data.(ByteView),nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter,key string)(ByteView,error){
	bytes,err := peer.Get(g.name,key)
	if err != nil{
		return ByteView{},err
	}
	return ByteView{b:bytes},nil
}


//从本地节点查询缓存
func (g *Group) getLocally(key string)(ByteView,error){
	bytes,err := g.getter.Get(key)
	if err!=nil{
		return ByteView{},err
	}
	value := ByteView{b:cloneBytes(bytes)}
	g.populateCache(key,value)
	return value, nil
}


//添加缓存
func (g *Group) populateCache(key string,value ByteView){
	g.mainCache.add(key,value)
}


