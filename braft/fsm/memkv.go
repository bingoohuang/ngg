package fsm

import (
	"log"
	"reflect"
	"sync"

	"github.com/bingoohuang/ngg/braft/marshal"
)

type KvOperate string

const (
	KvGet KvOperate = "get"
	KvSet KvOperate = "set"
	KvDel KvOperate = "del"
)

type KvRequest struct {
	Value     any
	KvOperate KvOperate
	MapName   string
	Key       string
}

type KvExecutable interface {
	Exec(req KvRequest) any
}

var _ interface {
	Service
	KvExecutable
} = (*MemKvService)(nil)

type MemKvService struct {
	lock *sync.Mutex
	Maps map[string]*Map
}

func NewMemKvService() *MemKvService {
	return &MemKvService{Maps: map[string]*Map{}, lock: &sync.Mutex{}}
}

// RegisterMarshalTypes registers the types for marshaling and unmarshaling.
func (m *MemKvService) RegisterMarshalTypes(t *marshal.TypeRegister) {
	t.RegisterType(reflect.TypeOf(KvRequest{}))
}

func (m *MemKvService) ApplySnapshot(nodeID string, input any) error {
	log.Printf("MemKvService ApplySnapshot req: %+v", input)
	service := input.(*MemKvService)
	m.Maps = service.Maps
	return nil
}

func (m *MemKvService) NewLog(nodeID string, req any) any {
	log.Printf("MemKvService NewLog req: %+v", req)
	return m.Exec(req.(KvRequest))
}

func (m *MemKvService) Exec(req KvRequest) any {
	m.lock.Lock()
	defer m.lock.Unlock()

	switch req.KvOperate {
	case KvSet:
		m.put(req.MapName, req.Key, req.Value)
	case KvGet:
		return m.get(req.MapName, req.Key)
	case KvDel:
		m.del(req.MapName, req.Key)
	}

	return nil
}

func (m *MemKvService) GetReqDataType() any { return KvRequest{} }

func (m *MemKvService) put(mapName string, key string, value any) {
	fMap, found := m.Maps[mapName]
	if !found {
		fMap = &Map{Data: map[string]any{}, lock: &sync.RWMutex{}}
		m.Maps[mapName] = fMap
	}

	fMap.put(key, value)
}

func (m *MemKvService) get(mapName string, key string) any {
	fMap, found := m.Maps[mapName]
	if !found {
		return nil
	}

	return fMap.get(key)
}

func (m *MemKvService) del(mapName string, key string) {
	fMap, found := m.Maps[mapName]
	if !found {
		return
	}

	fMap.del(key)
}

type Map struct {
	lock *sync.RWMutex
	Data map[string]any
}

func (m *Map) del(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.Data, k)
}

func (m *Map) get(k string) any {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.Data[k]
}

func (m *Map) put(k string, v any) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.Data[k] = v
}
