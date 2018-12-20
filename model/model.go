package model

import (
	"errors"
	"github.com/boltdb/bolt"
	"time"
)

type Model struct {
	DB         *bolt.DB
	BucketName string
}

func (m *Model) View(fn func(b *bolt.Bucket) error) error {
	err := m.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.BucketName))
		if b == nil {
			return errors.New("bucket_not_found")
		}
		return fn(b)
	})
	if err != nil && err.Error() == "bucket_not_found" {
		return m.Update(fn)
	}
	return err
}

func (m *Model) Update(fn func(b *bolt.Bucket) error) error {
	return m.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(m.BucketName))
		if err != nil {
			return err
		}
		return fn(b)
	})
}

func NewModel() (*Model, error) {
	db, err := bolt.Open("config/BlogReaper.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return &Model{DB: db}, nil
}
