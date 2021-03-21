package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct{
	hash Hash
	replicas int //虚拟节点倍数
	keys []int //哈希环
	hashMap map[int] string
}

func New(replicas int, fn Hash) *Map{
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int] string),
	}
	if m.hash == nil{
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string){
	for _,key := range keys{
		for i:=0;i<m.replicas;i++{
			hashValue := int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.keys = append(m.keys,hashValue)
			m.hashMap[hashValue] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string{
	if len(m.keys) == 0{
		return ""
	}
	hashValue := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.keys),func(i int)bool{
		return m.keys[i] >= hashValue
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
