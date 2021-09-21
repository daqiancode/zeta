package zeta

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type CachedDAO struct {
	*TableDAO
	Prefix     string
	Redis      *redis.Client
	Serializer ValueSerializer
	RedisisCtx context.Context
	TTL        int        // in minutes
	Indexes    [][]string //eg. {{Uid} , {ProjectId, ModelId}}
}

func NewCachedDAO(db *gorm.DB, model interface{}, table string, idField string, red *redis.Client, serializer ValueSerializer, prefix string, ttlMinutes int) *CachedDAO {
	return &CachedDAO{
		TableDAO:   NewTableDAO(db, model, table, idField),
		Prefix:     prefix,
		Redis:      red,
		Serializer: serializer,
		RedisisCtx: context.Background(),
		TTL:        ttlMinutes,
	}
}

func (s *CachedDAO) AddIndex(index ...string) *CachedDAO {
	s.Indexes = append(s.Indexes, index)
	return s
}

//MakeKey MakeKey make cache key. eg. project_name/users/id/1. input like: k1,v1,k2,v2
func (s *CachedDAO) MakeKey(indexValues ...interface{}) string {
	n := len(indexValues) / 2
	m := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		m[indexValues[2*i].(string)] = indexValues[2*i+1]
	}
	return s.MakeKeyByMap(m)
}

// MakeKey make cache key. eg. project_name/users/id/1
func (s *CachedDAO) MakeKeyByMap(indexValues map[string]interface{}) string {
	keys := make([]string, len(indexValues))
	values := make([]interface{}, len(indexValues))
	i := 0
	for k := range indexValues {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	i = 0
	for _, k := range keys {
		values[i] = indexValues[k]
		i++
	}
	tpl := s.Prefix + "/" + s.Table + "/"
	for _, v := range keys {
		tpl += v + "/%v"
	}
	tpl = strings.ToLower(tpl)
	return fmt.Sprintf(tpl, values...)
}

func (s *CachedDAO) GetIDValue(valueRef interface{}) uint64 {
	r := reflect.ValueOf(valueRef)
	return reflect.Indirect(r).FieldByNameFunc(func(n string) bool { return strings.EqualFold(n, s.IDField) }).Uint()
}

func (s *CachedDAO) Get(valueRef interface{}, id uint64) {
	if id > s.GetMaxID() {
		return
	}
	key := s.MakeKey(s.IDField, id)
	if s.CacheGet(key, valueRef) {
		return
	}
	s.TableDAO.Get(valueRef, id)
	if id == s.GetIDValue(valueRef) {
		s.CacheSet(key, valueRef)
	} else {
		s.CacheSet(key, struct{}{})
	}

}

func (s *CachedDAO) List(valueRef interface{}, ids []uint64) {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = s.MakeKey(s.IDField, id)
	}
	strs := s.GetCacheMany(keys)
	if !hasNil(strs) {
		n := 0
		for _, v := range strs {
			if v == "{}" {
				continue
			}
			n++
		}
		values := NewSliceRef(valueRef).New(n, n)
		i := 0
		for _, v := range strs {
			if v == "{}" {
				continue
			}
			s.Serializer.Unmarshal([]byte(v.(string)), values.GetRef(i))
			i++
		}
		return
	}
	s.TableDAO.List(valueRef, ids)
	values := NewSliceRef(valueRef)
	var key string
	var ele interface{}
	var id uint64
	var goodIds []uint64
	for i := 0; i < values.Len(); i++ {
		ele = values.Get(i)
		id = s.GetIDValue(ele)
		key = s.MakeKey(s.IDField, id)
		s.CacheSet(key, ele)
		goodIds = append(goodIds, id)
	}
	badIds := subtract(ids, goodIds)
	for _, v := range badIds {
		key = s.MakeKey(s.IDField, v)
		s.CacheSet(key, struct{}{})
	}

}

func (s *CachedDAO) CacheSet(key string, value interface{}) {
	var data []byte
	var err error
	if value == struct{}{} {
		data = []byte("{}")
	} else {
		data, err = s.Serializer.Marshal(value)
		s.HandleError(err)
	}
	setErr := s.Redis.SetEX(s.RedisisCtx, key, data, time.Duration(s.TTL)*time.Minute).Err()
	s.HandleError(setErr)
}

func (s *CachedDAO) CacheGet(key string, valueRef interface{}) bool {
	bs, err := s.Redis.Get(s.RedisisCtx, key).Bytes()
	if err == nil {
		errUnmarshal := s.Serializer.Unmarshal(bs, valueRef)
		s.HandleError(errUnmarshal)
		return true
	}
	if err != nil && err != redis.Nil {
		s.HandleError(err)
	}
	return false
}

func (s *CachedDAO) CacheGetUint64(key string) uint64 {
	var r uint64
	if s.CacheGet(key, &r) {
		return r
	}
	return 0
}

// return []string
func (s *CachedDAO) GetCacheMany(keys []string) []interface{} {
	// var r map[string][]byte
	rs, err := s.Redis.MGet(s.RedisisCtx, keys...).Result()
	s.HandleError(err)
	return rs
}
func (s *CachedDAO) ListBy(valueRef interface{}, indexes ...interface{}) {
	s.ListByWithMap(valueRef, argsToMap(indexes...))
}
func (s *CachedDAO) GetBy(valueRef interface{}, indexes ...interface{}) {
	s.GetByWithMap(valueRef, argsToMap(indexes...))
}

// redis key eg. projectusers/pid/2/uid/1 -> id
func (s *CachedDAO) GetByWithMap(valueRef interface{}, indexes map[string]interface{}) {
	key := s.MakeKeyByMap(indexes)
	id := s.CacheGetUint64(key)
	if id > 0 { // hit
		s.Get(valueRef, id)
		return
	}
	// miss
	s.TableDAO.GetByWithMap(valueRef, indexes)
	id = s.GetIDValue(valueRef)
	if id == 0 {
		return
	}
	s.CacheSet(key, id)
}

// redis key eg. projectusers/pid/2/uid/1 -> [id1,id2]
func (s *CachedDAO) ListByWithMap(valueRef interface{}, indexes map[string]interface{}) {
	key := s.MakeKeyByMap(indexes)
	var ids []uint64
	if ok := s.CacheGet(key, &ids); ok { //hit
		s.List(valueRef, ids)
		return
	}

	// miss
	s.TableDAO.ListByWithMap(valueRef, "", indexes)
	sr := NewSliceRef(valueRef)
	n := sr.Len()
	newIds := make([]uint64, n)
	for i := 0; i < n; i++ {
		newIds[i] = s.GetIDValue(sr.Get(i))

	}
	s.CacheSet(key, newIds)
}
func (s *CachedDAO) ClearCacheForTable() {
	s.ClearCacheWithPrefix(s.Table)
}

func (s *CachedDAO) ClearCacheWithPrefix(prefix string) {
	err := s.Redis.Eval(s.RedisisCtx, "return redis.call('del',unpack({'',unpack(redis.call('keys', KEYS[1]))}))", []string{s.Prefix + "/" + prefix + "*"}).Err()
	s.HandleError(err)
}

func (s *CachedDAO) GetKeyID(v interface{}) string {
	return s.MakeKey(s.IDField, s.GetIDValue(v))
}
func (s *CachedDAO) GetKeyMax() string {
	return s.MakeKeyByMap(map[string]interface{}{"__max": ""})
}

func (s *CachedDAO) GetMaxID() uint64 {
	key := s.GetKeyMax()
	var maxId uint64
	if ok := s.CacheGet(key, &maxId); ok {
		return maxId
	}
	maxId = s.TableDAO.GetMaxID()
	s.CacheSet(key, maxId)
	return maxId
}

// ClearCache clear cache after insert and delete .eg ClearCache(User{1},User{2})
func (s *CachedDAO) ClearCache(objs ...interface{}) {
	if len(objs) == 0 {
		return
	}
	sr := NewSliceRef(&objs)
	n := sr.Len()
	m := n*(len(s.Indexes)+1) + 1
	var keySet map[string]bool = make(map[string]bool, m)
	keySet[s.GetKeyMax()] = true
	for i := 0; i < n; i++ {
		v := sr.Get(i)
		keySet[s.GetKeyID(v)] = true
		for _, pairs := range s.Indexes {
			d := pick(v, pairs...)
			keySet[s.MakeKeyByMap(d)] = true
		}
	}

	rkeys := make([]string, m)
	i := 0
	for k := range keySet {
		rkeys[i] = k
		i++
	}
	err := s.Redis.Del(s.RedisisCtx, rkeys...).Err()
	s.HandleError(err)
}

func (s *CachedDAO) ClearCacheWithMaps(objs ...map[string]interface{}) {
	if len(objs) == 0 {
		return
	}
	n := len(objs)
	m := n*(len(s.Indexes)+1) + 1
	var keySet map[string]bool = make(map[string]bool, m)
	keySet[s.GetKeyMax()] = true
	for i := 0; i < n; i++ {
		v := objs[i]
		keySet[s.MakeKeyByMap(pickFromMap(v, s.IDField))] = true
		for _, pairs := range s.Indexes {
			m := pickFromMap(v, pairs...)
			keySet[s.MakeKeyByMap(m)] = true
		}
	}

	rkeys := make([]string, m)
	i := 0
	for k := range keySet {
		rkeys[i] = k
		i++
	}
	err := s.Redis.Del(s.RedisisCtx, rkeys...).Err()
	s.HandleError(err)
}

func (s *CachedDAO) Insert(valueRef interface{}) {
	s.TableDAO.Insert(valueRef)
	s.ClearCache(valueRef)
}
func (s *CachedDAO) InsertMany(valuesRef interface{}) {
	s.TableDAO.InsertMany(valuesRef)
	sr := NewSliceRef(valuesRef)
	n := sr.Len()
	var objs []interface{}
	for i := 0; i < n; i++ {
		objs = append(objs, sr.Get(i))
	}
	s.ClearCache(objs...)
}
func (s *CachedDAO) Delete(id uint64) {
	old := make(map[string]interface{})
	tx := s.DB.Model(s.Model).Take(&old, id)
	s.HandleError(tx.Error)
	s.TableDAO.Delete(id)
	s.ClearCacheWithMaps(old)
}
func (s *CachedDAO) DeleteMany(ids []uint64) {
	var old []map[string]interface{}
	tx := s.DB.Model(s.Model).Where(ids).Find(&old)
	s.HandleError(tx.Error)
	s.TableDAO.DeleteMany(ids)
	s.ClearCacheWithMaps(old...)
}
func (s *CachedDAO) Put(resultRef interface{}, fields ...string) {
	v1 := make(map[string]interface{})
	tx := s.DB.Model(s.Model).Take(&v1, s.GetIDValue(resultRef))
	s.HandleError(tx.Error)

	s.TableDAO.Put(resultRef, fields...)
	v2 := make(map[string]interface{})
	tx2 := s.DB.Model(s.Model).Take(&v2, s.GetIDValue(resultRef))
	s.HandleError(tx2.Error)
	s.ClearCacheWithMaps(v1, v2)
}

func (s *CachedDAO) Save(valueRef interface{}) {
	s.TableDAO.Save(valueRef)
	s.ClearCache(valueRef)
}
