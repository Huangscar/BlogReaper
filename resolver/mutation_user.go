package resolver

import (
	"context"
	"errors"
	"github.com/XMatrixStudio/BlogReaper/graphql"
)

func (r *mutationResolver) CreateLoginURL(ctx context.Context, backUrl string) (string, error) {
	url, state := r.Service.User.GetLoginURL(backUrl)
	r.Session.Set("state", state)
	return url, nil
}

func (r *mutationResolver) Login(ctx context.Context, code string, state string) (*graphql.User, error) {
	if r.Session.GetString("state") != state {
		return nil, errors.New("error_state")
	}
	userID, err := r.Service.User.LoginByCode(code)
	if err != nil {
		return nil, err
	}
	r.Session.Set("id", userID)
	user, err := r.Service.User.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	r.Session.Delete("id")
	return true, nil
}
