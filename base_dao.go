package zeta

import (
	"errors"

	"gorm.io/gorm"
)

type BaseDAO struct {
	DB *gorm.DB
}

func NewBaseDAO(db *gorm.DB) *BaseDAO {
	return &BaseDAO{DB: db}
}
func (s *BaseDAO) HandleError(err error) {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err)
	}
}

func (s *BaseDAO) GetMaxID(dbTableName string, dbIdField ...string) uint64 {
	type Result struct {
		R uint64
	}
	r := Result{}
	idField := "id"
	if len(dbIdField) > 0 {
		idField = dbIdField[0]
	}
	tx := s.DB.Raw("select max(" + idField + ") as r from " + dbTableName).Scan(&r)
	s.HandleError(tx.Error)
	return r.R
}

func (s *BaseDAO) Count(sql string, values ...interface{}) uint64 {
	var r uint64
	tx := s.DB.Raw(sql, values...).Scan(&r)
	s.HandleError(tx.Error)
	return r
}

func (s *BaseDAO) Select(result interface{}, sql string, values ...interface{}) {
	tx := s.DB.Raw(sql, values...).Scan(result)
	if tx.Error != nil {
		panic(tx.Error)
	}
}

func (s *BaseDAO) Get(resultRef interface{}, id uint64) {
	tx := s.DB.Take(resultRef, id)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) GetBy(resultRef interface{}, kvs ...interface{}) {
	s.GetByWithMap(resultRef, argsToMap(kvs))
}

func (s *BaseDAO) GetByWithMap(resultRef interface{}, kvs map[string]interface{}) {
	tx := s.DB.Where(kvs).Take(resultRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) List(resultRef interface{}, ids []uint64) {
	tx := s.DB.Where(ids).Find(resultRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) ListBy(resultRef interface{}, kvs ...interface{}) {
	s.ListByWithMap(resultRef, "", argsToMap(kvs...))
}
func (s *BaseDAO) ListByWithOrder(resultRef interface{}, orderBy string, kvs ...interface{}) {
	s.ListByWithMap(resultRef, orderBy, argsToMap(kvs...))
}

func (s *BaseDAO) ListByWithMap(resultRef interface{}, orderBy string, kvs map[string]interface{}) {
	tx := s.DB.Where(kvs).Order(orderBy).Find(resultRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) All(resultRef interface{}) {
	tx := s.DB.Find(&resultRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) Insert(valueRef interface{}) {
	tx := s.DB.Create(valueRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) InsertMany(valuesRef interface{}) {
	tx := s.DB.Create(valuesRef)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) Save(valueRef interface{}) {
	tx := s.DB.Save(valueRef)
	s.HandleError(tx.Error)
}
func (s *BaseDAO) Put(resultRef interface{}, fields ...string) {
	var tx *gorm.DB
	if len(fields) > 0 {
		tx = s.DB.Model(resultRef).Select(fields).Updates(resultRef)
	} else {
		tx = s.DB.Model(resultRef).Select("*").Updates(resultRef)
	}
	s.HandleError(tx.Error)
}

func (s *BaseDAO) Delete(resultRef interface{}, id uint64) {
	tx := s.DB.Delete(resultRef, id)
	s.HandleError(tx.Error)
}

func (s *BaseDAO) DeleteMany(valueRef interface{}, ids []uint64) {
	tx := s.DB.Delete(valueRef, ids)
	s.HandleError(tx.Error)
}
