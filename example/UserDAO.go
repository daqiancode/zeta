package example

type userDAO struct {
	*CachedDAO
}

var userDAOInstance *userDAO = nil

func NewUserDAO(db *gorm.DB, red *redis.Client) *userDAO {
	if userDAOInstance != nil {
		return userDAOInstance
	}
	r := &userDAO{CachedDAO: NewCachedDAO(db, &tables.User{}, "users", "ID", red, &JSONValueSerializer{}, "iam", 30)}
	r.AddIndex("Name")
	userDAOInstance = r
	return r
}

func (s *userDAO) Get(id uint64) tables.User {
	r := tables.User{}
	s.CachedDAO.Get(&r, id)
	return r
}

func (s *userDAO) List(ids []uint64) []tables.User {
	var r []tables.User
	s.CachedDAO.List(&r, ids)
	return r
}

func (s *userDAO) All() []tables.User {
	var r []tables.User
	s.CachedDAO.All(&r)
	return r
}
func (s *userDAO) ListByName(name string) []tables.User {
	var r []tables.User
	s.CachedDAO.ListBy(&r, "name", name)
	return r
}

func (s *userDAO) ListByNameNoCache(name string) []tables.User {
	var r []tables.User
	s.TableDAO.ListBy(&r, "name", name)
	return r
}
