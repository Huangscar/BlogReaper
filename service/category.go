package service

import (
	"errors"
	"fmt"
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/XMatrixStudio/BlogReaper/model"
	"os"
)

type CategoryService interface {
	GetModel() *model.CategoryModel
	AddCategory(userID, name string) (category graphql.Category, err error)
	GetCategories(userID string) (categories []graphql.Category, err error)
	EditCategory(userID, categoryID, newName string) (success bool, err error)
	RemoveCategory(userID, categoryID string) (success bool, err error)
}

type categoryService struct {
	Model   *model.CategoryModel
	Service *Service
}

func NewCategoryService(s *Service, m *model.CategoryModel) CategoryService {
	_, err := m.DB.Exec(`CREATE TABLE IF NOT EXISTS ` + m.TableName + ` (json JSON)`)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	return &categoryService{
		Model:   m,
		Service: s,
	}
}

func (s *categoryService) GetModel() *model.CategoryModel {
	return s.Model
}

func (s *categoryService) AddCategory(userID, name string) (category graphql.Category, err error) {
	c, err := s.Model.AddCategory(userID, name)
	if err != nil {
		return category, err
	}
	category = graphql.Category{
		ID:    c.ID,
		Name:  c.Name,
		Feeds: nil,
	}
	return
}

func (s *categoryService) GetCategories(userID string) (categories []graphql.Category, err error) {
	cs, err := s.Model.GetCategories(userID)
	if err != nil {
		return categories, err
	}
	for _, c := range cs {
		feeds, err := s.Service.Feed.GetFeedsByCategoryID(userID, c.ID)
		if err != nil && err.Error() == "not_found" {
			feeds = nil
		}
		categories = append(categories, graphql.Category{
			ID:    c.ID,
			Name:  c.Name,
			Feeds: feeds,
		})
	}
	return
}

func (s *categoryService) EditCategory(userID, categoryID, newName string) (success bool, err error) {
	category, err := s.Model.GetCategoryByName(userID, newName)
	if err != nil && err.Error() != "not_found" {
		return false, err
	}
	if category.Name != "" && category.ID != categoryID {
		return false, errors.New("repeat_name")
	}
	if category.Name != "" {
		return true, nil
	}
	return s.Model.EditCategory(userID, categoryID, newName)
}

func (s *categoryService) RemoveCategory(userID, categoryID string) (success bool, err error) {
	_, err = s.Model.GetCategoryById(userID, categoryID)
	if err != nil {
		return false, err
	}
	feeds, err := s.Service.Feed.GetModel().GetFeedsByCategoryID(userID, categoryID)
	if err != nil && err.Error() != "not_found" {
		return false, err
	}
	categories, err := s.Model.GetCategories(userID)
	if err != nil {
		return false, err
	}
	categoryMap := make(map[string]bool)
	for _, category := range categories {
		if category.ID != categoryID {
			categoryMap[category.ID] = true
		}
	}
	for _, feed := range feeds {
		categoryIDs, err := s.Service.Feed.GetModel().GetCategoryByFeedID(userID, feed.ID.Hex())
		if err != nil {
			return false, err
		}
		for i, id := range categoryIDs {
			_, exist := categoryMap[id]
			if exist {
				categoriesString := append(categoryIDs[:i],categoryIDs[i+1:]...)
				_, err := s.Service.Feed.EditFeed(userID, feed.ID.Hex(), nil, categoriesString)
				if err != nil {
					return false, err
				}
			} else {
				_, err := s.Service.Feed.RemoveFeed(userID, feed.ID.Hex())
				if err != nil {
					return false, err
				}
			}
		}
	}
	success, err = s.Model.RemoveCategory(userID, categoryID)
	return 
}
