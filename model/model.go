package model

import (
	"database/sql"
	"github.com/boltdb/bolt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

type Model struct {
	DB        *sql.DB
	TableName string
}

func (m *Model) View(fn func(b *bolt.Bucket) error) error {
	panic("deprecated method")
}

func (m *Model) Update(fn func(b *bolt.Bucket) error) error {
	panic("deprecated method")
}

func NewModel(conf mysql.Config) (*Model, error) {
	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}
	return &Model{DB: db}, nil
}
