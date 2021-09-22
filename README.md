# zeta
[![Build Status](https://app.travis-ci.com/daqiancode/zeta.svg?branch=main)](https://app.travis-ci.com/daqiancode/zeta)   
integrate gorm and redis

## Installation
```shell
go get https://github.com/daqiancode/zeta
```

## Cached DAO Step:
1. Define models and migrate
2. Define NewCachedDAO
3. Add `Indexes` for each table's NewCachedDAO. Clearing cache action will use those indexes. Indexes must cover `GetBy` and `ListBy` fields
4. `ClearCache` manually after transaction or using `non-NewCachedDAO` updating methods

## Examples:
- Define models
```go 
package tables

import (
	"time"
)

type Model struct {
	ID        uint64    `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"type:datetime not null;"`
	UpdatedAt time.Time `gorm:"type:datetime not null;"`
	// DeletedAt sql.NullTime `gorm:"index"`
}

type ModelUser struct {
	Model
	UID uint64 `gorm:"type:bigint unsigned not null;fk:User"`
}

type User struct {
	Model
	Name     string `gorm:"type:varchar(30) not null;"`
	Disabled bool   `gorm:"not null;default:false"`
}
type UserPrivate struct {
	Model
	UID      uint64 `gorm:"type:bigint unsigned not null;fk:User;uniqueIndex"`
	Email    string `gorm:"type:varchar(100);uniqueIndex"`
	Mobile   string `gorm:"type:varchar(20);uniqueIndex"`
	Password string `gorm:"type:varchar(32) not null;"`
}
```
- Define database connections
```go
package dao

import (
	"iam/config"
	"log"
	"net/url"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/daqiancode/zeta"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Namer struct {
	schema.NamingStrategy
}

func (ns Namer) ColumnName(table, column string) string {
	return column
}

var _db *gorm.DB
func GetMysql() *gorm.DB {
	if _db != nil {
		return _db
	}
	mysqlUrl, err := url.PathUnescape(config.Config.MysqlUrl)
	if err != nil {
		panic(err)
	}
	dbConfig := &gorm.Config{}
	namer := &Namer{}
	dbConfig.NamingStrategy = namer
	db, err := gorm.Open(mysql.Open(mysqlUrl), dbConfig)
	if err != nil {
		panic(err)
	}
	_db = db
	return db

}

var _rdb *redis.Client

func GetRedis() *redis.Client {
	if _rdb != nil {
		return _rdb
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
		DB:       config.Config.Redis.DB,
	})
	return rdb
}

var tableIndexes map[string][][]string = map[string][][]string{
	"UserPrivate": {{"UID"}},
	"UserToken":   {{"UID"}, {"AccessToken"}},
}
var ddl *zeta.DDL = nil
var cachedDaos sync.Map

func NewCachedDAO(entityRef interface{}) *zeta.CachedDAO {
	structName := reflect.Indirect(reflect.ValueOf(entityRef)).Type().Name()
	if r, ok := cachedDaos.Load(structName); ok {
		return r.(*zeta.CachedDAO)
	}
	db := GetMysql()
	if ddl == nil {
		ddl = zeta.NewDDL(db)
	}
	ddl.AddTables(entityRef)
	sch := ddl.GetSchema(entityRef)
	r := zeta.NewCachedDAO(db, entityRef, sch.Table, sch.PrimaryFields[0].Name, GetRedis(), &zeta.JSONValueSerializer{}, "user_center", config.Config.Redis.TTL)
	if indexes, ok := tableIndexes[structName]; ok {
		for _, index := range indexes {
			r.AddIndex(index...)
		}
	}
	cachedDaos.Store(structName, r)
	return r
}

func NewBaseDAO() *zeta.BaseDAO {
	return zeta.NewBaseDAO(GetMysql())
}

```

- Migrate models into database
```go
package dao
func Migrate(t *testing.T) {
	db := GetMysql()

	models := []interface{}{&tables.User{}, &tables.UserPrivate{}, &tables.UserToken{}}

	err := db.AutoMigrate(models...)
	if err != nil {
		panic(err)
	}

	ddl := zeta.NewDDL(db)
	ddl.AddTables(models...)
	ddl.MakeFKs() // handle foreign key like "UID uint64 "`gorm:"fk:User"`"
}
```

- Use CachedDAO and BaseDAO in service
```go
package service

type Users struct {
	userDao *zeta.CachedDAO
}

var _users *Users = nil

func NewUsers() *Users {
	if _users != nil {
		return _users
	}
	_users := &Users{
		userDao: dao.NewCachedDAO(&tables.User{}),
	}
	return _users
}

func (s *Users) Get(id uint64) tables.User {
	var r tables.User
	s.userDao.Get(&r, id)
	return r
}

// Create should be run in transaction
func (s *Users) Create(name string) tables.User {
	var r tables.User
	r.Name = name
	service.userDao.Insert(&r)
	return r
}
```
- Transaction(Cache will be disabled in transaction)
```go
func Signup() {
    err := s.userPrivateDao.Transaction(func(tx *zeta.BaseDAO) error {
		newUser := NewUsers().Create(tx, info.Name)
		up := tables.UserPrivate{UID: newUser.ID, Email: info.Email, Mobile: info.Mobile, Password: makepassword(info.Password)}
		tx.Insert(&up)
		
		token = s.userTokenDao.Create(tx, newUser.ID)
        // Clear cache manually
        s.userDao.ClearCache(newUser)
		s.userTokenDao.ClearCache(token)
		s.userPrivateDao.ClearCache(up)
		return nil
	})

	return token, err
}
```