package model

import (
	"encoding/json"
	"errors"
	"github.com/globalsign/mgo/bson"
)

type UserModel struct {
	*Model
}

type User struct {
	VioletID string   `json:"vid" db:"vid"`     // VioletID
	Token    string   `json:"token" db:"token"` // Violet 访问令牌
	Email    string   `json:"email" db:"email"` // 用户唯一邮箱
	Info     UserInfo `json:"info" db:"info"`   // 用户个性信息
}

// UserInfo 用户个性信息
type UserInfo struct {
	Name   string `json:"name" db:"name"`     // 用户昵称
	Avatar string `json:"avatar" db:"avatar"` // 头像URL
	Bio    string `json:"bio" db:"bio"`       // 个人简介
	Gender int    `json:"gender" db:"gender"` // 性别
}

// GetUserByID 根据ID查询用户
func (m *UserModel) GetUserByID(id string) (user User, err error) {
	if !bson.IsObjectIdHex(id) {
		return user, errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`SELECT * FROM user WHERE JSON_EXTRACT(json, "$.vid") = ?`)
	if err != nil {
		return user, err
	}
	rows, err := stmt.Query(id)
	if err != nil {
		return user, err
	}
	if rows.Next() {
		var bytes []byte
		err = rows.Scan(&bytes)
		if err != nil {
			return user, err
		}
		err = json.Unmarshal(bytes, &user)
	} else {
		return user, errors.New("not_found")
	}
	return user, err
}

// AddUser 添加用户
func (m *UserModel) AddUser(id, token, email, name, avatar, bio string, gender int) error {
	if !bson.IsObjectIdHex(id) {
		return errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`INSERT INTO user VALUES(?)`)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(&User{
		VioletID: id,
		Token:    token,
		Email:    email,
		Info: UserInfo{
			Name:   name,
			Avatar: avatar,
			Bio:    bio,
			Gender: gender,
		},
	})
	if err != nil {
		return err
	}
	_, err = stmt.Exec(string(bytes))
	return err
}

// SetUserToken 设置Token
func (m *UserModel) SetUserToken(id, token string) error {
	if !bson.IsObjectIdHex(id) {
		return errors.New("not_id")
	}
	stmt, err := m.DB.Prepare(`UPDATE user SET json = JSON_SET(json, "$.token", ?) WHERE JSON_EXTRACT(json, "$.vid") = ?`)
	if err != nil {
		return err
	}
	result, err := stmt.Exec(token, id)
	if err != nil {
		return err
	}
	if i, _ := result.RowsAffected(); i == 0 {
		return errors.New("not_found")
	}
	return nil
}
