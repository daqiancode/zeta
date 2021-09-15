package zeta

import "gorm.io/gorm"

type TableDAO struct {
	*BaseDAO
	Model   interface{} // eg. &tables.User{}
	Table   string      // database table name of the model
	IDField string      // ID
}

func NewTableDAO(db *gorm.DB, model interface{}, table string, idField string) *TableDAO {
	return &TableDAO{BaseDAO: NewBaseDAO(db), Model: model, Table: table, IDField: "ID"}
}

func (s *TableDAO) GetMaxID() uint64 {
	return s.BaseDAO.GetMaxID(s.Table, s.IDField)
}

func (s *TableDAO) Delete(id uint64) {
	s.BaseDAO.Delete(s.Model, id)
}

func (s *TableDAO) DeleteMany(ids []uint64) {
	s.BaseDAO.DeleteMany(s.Model, ids)
}
