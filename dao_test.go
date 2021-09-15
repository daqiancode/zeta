package zeta_test

import (
	"github.com/daqiancode/zeta"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

var userDAOInstance *userDAO = nil

type userDAO struct {
	*zeta.CachedDAO
}

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

func (s *userDAO) ListByNameNoCache(name string) []User {
	var r []User
	s.TableDAO.ListBy(&r, "name", name)
	return r
}

type UserDAOTest struct {
	suite.Suite
	dao *userDAO
}

func (s *UserDAOTest) SetupTest() {
	s.dao = NewUserDAO(GetMysql(), GetRedis())
}

func (s *UserDAOTest) TestGet() {
	userInsert := User{Name: "test1"}
	s.dao.Insert(&userInsert)
	userGet := s.dao.Get(userInsert.ID)
	s.Equal(userInsert.ID, userGet.ID)

}
