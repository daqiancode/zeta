package zeta

import (
	"reflect"
	"sync"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type _Namer struct {
	schema.NamingStrategy
}

func (ns _Namer) ColumnName(table, column string) string {
	return column
}

var Namer = &_Namer{}

func TableName(v interface{}) string {
	return Namer.TableName(reflect.Indirect(reflect.ValueOf(v)).Type().Name())
}

func Open(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
	if config.NamingStrategy == nil {
		config.NamingStrategy = Namer
	}
	db, err := gorm.Open(dialector, config)
	if err != nil {
		panic(err)
	}
	return db, err
}

var cachedDaos sync.Map
var ddl *DDL = nil

func NewCachedTableDAO(db *gorm.DB, redisClient *redis.Client, cacheTTLMinutes int, serializer ValueSerializer, appName string, entityRef interface{}) *CachedDAO {
	structName := reflect.Indirect(reflect.ValueOf(entityRef)).Type().Name()
	if r, ok := cachedDaos.Load(structName); ok {
		return r.(*CachedDAO)
	}
	if ddl == nil {
		ddl = NewDDL(db)
	}
	ddl.AddTables(entityRef)
	sch := ddl.GetSchema(entityRef)
	r := NewCachedDAO(db, entityRef, sch.Table, sch.PrimaryFields[0].Name, redisClient, serializer, appName, cacheTTLMinutes)
	if indexes, ok := cacheIndex.Load(TableName(entityRef)); ok {
		for _, index := range indexes.([][]string) {
			r.AddIndex(index...)
		}
	}
	cachedDaos.Store(structName, r)
	return r
}
