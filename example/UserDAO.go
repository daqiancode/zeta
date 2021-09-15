package example

import (
	"github.com/daqiancode/zeta"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type User struct {
	ID uint64
}
type userDAO struct {
	*zeta.CachedDAO
}

var userDAOInstance *userDAO = nil

func NewUserDAO(db *gorm.DB, red *redis.Client) *userDAO {
	if userDAOInstance != nil {
		return userDAOInstance
	}
	r := &userDAO{CachedDAO: zeta.NewCachedDAO(db, &User{}, "users", "ID", red, &zeta.JSONValueSerializer{}, "iam", 30)}
	r.AddIndex("Name")
	userDAOInstance = r
	return r
}

func (s *userDAO) Get(id uint64) User {
	r := User{}
	s.CachedDAO.Get(&r, id)
	return r
}

func (s *userDAO) List(ids []uint64) []User {
	var r []User
	s.CachedDAO.List(&r, ids)
	return r
}

func (s *userDAO) All() []User {
	var r []User
	s.CachedDAO.All(&r)
	return r
}
func (s *userDAO) ListByName(name string) []User {
	var r []User
	s.CachedDAO.ListBy(&r, "name", name)
	return r
}
