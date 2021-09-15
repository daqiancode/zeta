package zeta_test

import (
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/daqiancode/zeta"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func GetMysql() *gorm.DB {
	if _db != nil {
		return _db
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)
	mysqlUrl, err := url.PathUnescape("root:@tcp(localhost:3306)/zeta?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(mysqlUrl), &gorm.Config{Logger: newLogger})
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
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return rdb
}

type Model struct {
	ID        uint64    `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"type:datetime not null;"`
	UpdatedAt time.Time `gorm:"type:datetime not null;"`
	// DeletedAt sql.NullTime `gorm:"index"`
}

type ModelUser struct {
	Model
	UID uint `gorm:"type:bigint unsigned not null;fk:User"`
}

type User struct {
	Model
	Name     string `gorm:"type:varchar(30) not null;"`
	Disabled bool   `gorm:"not null;default:false"`
}

type UserProject struct {
	ModelUser
	Name     string `gorm:"type:varchar(30) not null;"`
	Disabled bool   `gorm:"not null;default:false"`
}

func TestInit(t *testing.T) {
	db := GetMysql()

	models := []interface{}{&User{}, &UserProject{}}

	err := db.AutoMigrate(models...)
	assert.Nil(t, err)

	ddl := zeta.NewDDL(db)
	ddl.AddTables(models...)
	ddl.MakeFKs()

}
