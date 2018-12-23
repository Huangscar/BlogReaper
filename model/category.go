package model

import (
	"encoding/json"
	"errors"
	"github.com/globalsign/mgo/bson"
)

type CategoryModel struct {
	*Model
}

type Category struct {
	UserID string	`json:"uid" db:"uid"`	// 用户ID
	ID   string		`json:"id" db:"id"`   	// 订阅分类的ID
	Name string     `json:"name" db:"name"` // 订阅分类的名称
}

func (m *CategoryModel) AddCategory(userID, name string) (category Category, err error) {
	if !bson.IsObjectIdHex(userID) {
		return category, errors.New("not_id")
	}

	stmt, err := m.DB.Prepare(`SELECT * FROM category WHERE JSON_EXTRACT(json, "$.name") = ?`)
	if err != nil {
		return category, err
	}
	rows, err := stmt.Query(name)
	if err != nil {
		return category, err
	}
	if rows.Next() {
		return Category{}, errors.New("repeat_name")
	}

	stmt, err = m.DB.Prepare("INSERT INTO category VALUES(?)")
	if err != nil {
		return category, err
	}
	category = Category{
		UserID:	userID,
		ID:		bson.NewObjectId().Hex(),
		Name:	name,
	}
	bytes, err := json.Marshal(&category)
	if err != nil {
		return category, err
	}
	_, err = stmt.Exec(string(bytes))
	return category, err
}

func (m *CategoryModel) GetCategoryById(userID, categoryID string) (category Category, err error) {
	if !bson.IsObjectIdHex(userID) || !bson.IsObjectIdHex(categoryID) {
		return category, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`SELECT * FROM category WHERE JSON_EXTRACT(json, "$.uid") = ? AND JSON_EXTRACT(json, "$.id") = ?`)
	if err != nil {
		return category, err
	}
	rows, err := stmt.Query(userID, categoryID)
	if err != nil {
		return category, err
	}
	if rows.Next() {
		var bytes []byte
		err = rows.Scan(&bytes)
		if err != nil {
			return category, err
		}
		err = json.Unmarshal(bytes, &category)
	}
	return category, err
}

func (m *CategoryModel) GetCategories(userID string) (categories []Category, err error) {
	if !bson.IsObjectIdHex(userID) {
		return categories, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`SELECT * FROM category WHERE JSON_EXTRACT(json, "$.uid") = ?`)
	if err != nil {
		return categories, err
	}
	rows, err := stmt.Query(userID)
	if err != nil {
		return categories, err
	}
	for {
		if rows.Next() {
			var bytes []byte
			var category Category
			err = rows.Scan(&bytes)
			if err != nil {
				return categories, err
			}
			err = json.Unmarshal(bytes, &category)
			if err != nil {
				return categories, err
			}
			categories = append(categories, category)
		} else {
			break
		}
	}
	return categories, nil
}

func (m *CategoryModel) GetCategoryByName(userID, name string) (category Category, err error) {
	if !bson.IsObjectIdHex(userID) {
		return category, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`SELECT * FROM category WHERE JSON_EXTRACT(json, "$.uid") = ? AND JSON_EXTRACT(json, "$.name") = ?`)
	if err != nil {
		return category, err
	}
	rows, err := stmt.Query(userID, name)
	if err != nil {
		return category, err
	}
	if rows.Next() {
		var bytes []byte
		err = rows.Scan(&bytes)
		if err != nil {
			return category, err
		}
		err = json.Unmarshal(bytes, &category)
	}
	return category, err
}

func (m *CategoryModel) EditCategory(userID, categoryID, newName string) (success bool, err error) {
	success = false
	if !bson.IsObjectIdHex(userID) || !bson.IsObjectIdHex(categoryID) {
		return success, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`UPDATE category SET json = JSON_SET(json, "$.name", ?) WHERE JSON_EXTRACT(json, "$.uid") = ? AND JSON_EXTRACT(json, "$.id") = ?`)
	if err != nil {
		return success, err
	}
	result, err := stmt.Exec(newName, userID, categoryID)
	if err != nil {
		return success, err
	}
	if i, _ := result.RowsAffected(); i == 0 {
		return success, errors.New("not_found")
	}
	return true, nil
}

func (m *CategoryModel) RemoveCategory(userID, categoryID string) (success bool, err error) {
	success = false
	if !bson.IsObjectIdHex(userID) || !bson.IsObjectIdHex(categoryID) {
		return success, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`DELETE FROM category WHERE JSON_EXTRACT(json, "$.uid") = ? AND JSON_EXTRACT(json, "$.id") = ?`)
	if err != nil {
		return success, err
	}
	result, err := stmt.Exec(userID, categoryID)
	if err != nil {
		return success, err
	}
	if i, _ := result.RowsAffected(); i == 0 {
		return success, errors.New("not_found")
	}
	return true, nil
}
