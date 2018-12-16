package resolver

import (
	"context"
	"errors"
	"github.com/XMatrixStudio/BlogReaper/graphql"
)

func (r *userResolver) Categories(ctx context.Context, obj *graphql.User, id *string) ([]graphql.Category, error) {
	if obj == nil {
		return nil, nil
	}
	if id == nil {
		return obj.Categories, nil
	}
	for _, v := range obj.Categories {
		if v.ID == *id {
			return []graphql.Category{v}, nil
		}
	}
	return nil, errors.New("invalid_id")
}

func (r *userResolver) LaterArticles(ctx context.Context, obj *graphql.User, page, numPerPage *int) ([]graphql.Article, error) {
	userID := r.Session.GetString("id")
	if userID == "" {
		return nil, errors.New("not_login")
	}
	if obj == nil {
		return nil, nil
	}
	if (page != nil && numPerPage == nil) || (page == nil && numPerPage != nil) || (page != nil && *page <= 0) || (numPerPage != nil && *numPerPage <= 0) {
		return nil, errors.New("invalid_params")
	}
	articles, err := r.Service.Feed.GetLaterArticles(userID, page, numPerPage)
	if err != nil {
		return nil, err
	}
	return articles, nil
}
