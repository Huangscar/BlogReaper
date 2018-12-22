package resolver

import (
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/XMatrixStudio/BlogReaper/service"
	"github.com/kataras/go-sessions"
)

type Resolver struct {
	Service *service.Service
	Session *sessions.Session
}

var resolver *Resolver

func DefaultResolver() *Resolver {
	if resolver == nil {
		resolver = &Resolver{
			Session: nil,
		}
	}
	return resolver
}

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
type categoryResolver struct{ *Resolver }
type feedResolver struct{ *Resolver }

func (r *Resolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) User() graphql.UserResolver {
	return &userResolver{r}
}

func (r *Resolver) Category() graphql.CategoryResolver {
	return &categoryResolver{r}
}

func (r *Resolver) Feed() graphql.FeedResolver {
	return &feedResolver{r}
}
