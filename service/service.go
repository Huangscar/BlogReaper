package service

import (
	"fmt"
	"github.com/XMatrixStudio/BlogReaper/model"
	"github.com/go-sql-driver/mysql"
	"os"
)

type Service struct {
	User     UserService
	Feed     FeedService
	Category CategoryService
	Public   PublicService
}

func NewService(conf mysql.Config) *Service {
	s := &Service{}
	m, err := model.NewModel(conf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	s.User = NewUserService(s, &model.UserModel{Model: &model.Model{TableName: "user", DB: m.DB}})
	s.Feed = NewFeedService(s, &model.FeedModel{Model: &model.Model{TableName: "feed", DB: m.DB}})
	s.Category = NewCategoryService(s, &model.CategoryModel{Model: &model.Model{TableName: "category", DB: m.DB}})
	s.Public = NewPublicService(s, &model.PublicModel{Model: &model.Model{TableName: "public", DB: m.DB}})
	return s
}
