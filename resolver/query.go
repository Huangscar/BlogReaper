package resolver

import (
	"context"
	"errors"
	"github.com/XMatrixStudio/BlogReaper/graphql"
)

func (r *queryResolver) User(ctx context.Context) (*graphql.User, error) {
	userID := r.Session.GetString("id")
	if userID == "" {
		return nil, errors.New("not_login")
	}
	user, err := r.Service.User.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *queryResolver) Feeds(ctx context.Context, id *string, keyword *string) ([]graphql.Feed, error) {
	if keyword != nil && id != nil {
		return nil, errors.New("invalid_params")
	}
	if keyword != nil {
		feeds, err := r.Service.Public.GetPublicFeedByKeyword(*keyword)
		return feeds, err
	} else {
		feed, err := r.Service.Public.GetPublicFeedByID(*id)
		return []graphql.Feed{feed}, err
	}
}

func (r *queryResolver) PopularArticles(ctx context.Context, page int, numPerPage int) ([]graphql.Article, error) {
	if page <= 0 || numPerPage <= 0 {
		return nil, errors.New("invalid_params")
	}
	feeds, err := r.Service.Public.GetPopularPublicArticles(page, numPerPage)
	if err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *queryResolver) PopularFeeds(ctx context.Context, page int, numPerPage int) ([]graphql.Feed, error) {
	if page <= 0 || numPerPage <= 0 {
		return nil, errors.New("invalid_params")
	}
	feeds, err := r.Service.Public.GetPopularPublicFeeds(page, numPerPage)
	if err != nil {
		return nil, err
	}
	return feeds, nil
}
