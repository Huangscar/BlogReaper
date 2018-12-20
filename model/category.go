package model

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/globalsign/mgo/bson"
)

type CategoryModel struct {
	*Model
}

type Category struct {
	ID   bson.ObjectId `bson:"id"`   // 订阅分类的ID
	Name string        `bson:"name"` // 订阅分类的名称
}

func (m *CategoryModel) AddCategory(userID, name string) (category Category, err error) {
	return category, m.Update(func(b *bolt.Bucket) error {
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		nub, err := ub.CreateBucketIfNotExists([]byte("key_name_value_id"))
		if err != nil {
			return err
		}
		if nub.Get([]byte(name)) != nil {
			return errors.New("repeat_name")
		}
		category = Category{
			ID:   bson.NewObjectId(),
			Name: name,
		}
		bytes, err := bson.Marshal(&category)
		if err != nil {
			return err
		}
		err = ub.Put([]byte(category.ID.Hex()), bytes)
		if err != nil {
			return err
		}
		return nub.Put([]byte(name), []byte(category.ID.Hex()))
	})
}

func (m *CategoryModel) GetCategoryById(userID, categoryID string) (category Category, err error) {
	return category, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(categoryID))
		if bytes == nil {
			return errors.New("not_found")
		}
		return bson.Unmarshal(bytes, &category)
	})
}

func (m *CategoryModel) GetCategories(userID string) (categories []Category, err error) {
	return categories, m.View(func(b *bolt.Bucket) error {
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return nil
		}
		return ub.ForEach(func(k, v []byte) error {
			if string(k) != "key_name_value_id" {
				category := Category{}
				err = bson.Unmarshal(v, &category)
				if err != nil {
					return err
				}
				categories = append(categories, category)
			}
			return nil
		})
	})
}

func (m *CategoryModel) GetCategoryByName(userID, name string) (category Category, err error) {
	return category, m.View(func(b *bolt.Bucket) error {
		category = Category{}

		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return nil
		}
		nub := ub.Bucket([]byte("key_name_value_id"))
		if nub == nil {
			return nil
		}
		categoryID := nub.Get([]byte(name))
		if categoryID == nil {
			return nil
		}
		bytes := ub.Get(categoryID)
		err = bson.Unmarshal(bytes, &category)
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *CategoryModel) EditCategory(userID, categoryID, newName string) (success bool, err error) {
	return success, m.Update(func(b *bolt.Bucket) error {
		success = false
		ub, err := b.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		nub, err := ub.CreateBucketIfNotExists([]byte("key_name_value_id"))
		if err != nil {
			return err
		}

		bytes := ub.Get([]byte(categoryID))
		if bytes == nil {
			return errors.New("invalid_id")
		}
		oldCategory := Category{}
		err = bson.Unmarshal(bytes, &oldCategory)
		if err != nil {
			return err
		}
		oldName := oldCategory.Name

		err = nub.Delete([]byte(oldName))
		if err != nil {
			return err
		}
		err = ub.Delete([]byte(categoryID))
		if err != nil {
			return err
		}

		newCategory := Category{
			ID:   bson.ObjectIdHex(categoryID),
			Name: newName,
		}
		bytes, err = bson.Marshal(&newCategory)
		if err != nil {
			return err
		}

		err = ub.Put([]byte(categoryID), bytes)
		if err != nil {
			return err
		}
		err = nub.Put([]byte(newName), []byte(categoryID))
		if err != nil {
			return err
		}
		success = true
		return nil
	})
}

func (m *CategoryModel) RemoveCategory(userID, categoryID string) (success bool, err error) {
	return success, m.Update(func(b *bolt.Bucket) error {
		success = false
		ub := b.Bucket([]byte(userID))
		if ub == nil {
			return errors.New("not_found")
		}
		nub := ub.Bucket([]byte("key_name_value_id"))
		if nub == nil {
			return errors.New("not_found")
		}
		bytes := ub.Get([]byte(categoryID))
		if bytes == nil {
			return errors.New("not_found")
		}
		category := Category{}
		bson.Unmarshal(bytes, &category)
		name := category.Name
		err = ub.Delete([]byte(categoryID))
		if err != nil {
			return err
		}
		err = nub.Delete([]byte(name))
		if err != nil {
			return err
		}
		success = true
		return nil
	})
}
