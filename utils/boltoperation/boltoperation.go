package boltoperation

import (
	"log"

	kv "github.com/ben-han-cn/kvzoo"
	"github.com/ben-han-cn/kvzoo/backend/bolt"
)

type BoltHandler struct {
	db     kv.DB
	dBPath string
}

func NewBoltHandler(dbPath string, dbName string) *BoltHandler {
	log.Println("in NewBoltHandler, dbPath: ", dbPath)
	log.Println("in NewBoltHandler, dbName: ", dbName)
	var tmpDBPath string
	if dbPath[len(dbPath)-1] != '/' {
		tmpDBPath = dbPath + "/"
	} else {
		tmpDBPath = dbPath
	}

	instance := &BoltHandler{dBPath: tmpDBPath}
	pbolt, err := bolt.New(tmpDBPath + dbName)
	if err != nil {
		panic(err)
	}
	instance.db = pbolt
	return instance
}

func (handler *BoltHandler) TableKVs(table string) (map[string][]byte, error) {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(table))
	if err != nil {
		return nil, err
	}
	var ts kv.Transaction
	if ts, err = tb.Begin(); err != nil {
		return nil, err
	}
	defer ts.Rollback()
	kvs, err := ts.List()
	if err != nil {
		return nil, err
	}
	return kvs, nil
}

func (handler *BoltHandler) Tables(table string) ([]string, error) {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(table))
	if err != nil {
		return nil, err
	}
	var ts kv.Transaction
	if ts, err = tb.Begin(); err != nil {
		return nil, err
	}
	defer ts.Rollback()
	tables, err := ts.Tables()
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (handler *BoltHandler) AddKVs(tableName string, values map[string][]byte) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for k, value := range values {
		if err := ts.Add(k, value); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}

func (handler *BoltHandler) UpdateKVs(tableName string, values map[string][]byte) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for k, value := range values {
		if err := ts.Update(k, value); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}

func (handler *BoltHandler) DeleteKVs(tableName string, keys []string) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for _, key := range keys {
		if err := ts.Delete(key); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}
