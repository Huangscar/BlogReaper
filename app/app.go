package app

import (
	"github.com/99designs/gqlgen/handler"
	"github.com/XMatrixStudio/BlogReaper/graphql"
	"github.com/XMatrixStudio/BlogReaper/resolver"
	"github.com/XMatrixStudio/BlogReaper/service"
	"github.com/XMatrixStudio/Violet.SDK.Go"
	"github.com/go-chi/chi"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
)

func App(c Config) http.Handler {
	r := chi.NewRouter()
	r.Use(resolver.SessionHttpMiddleware)
	dr := resolver.DefaultResolver()
	dr.Service = service.NewService(mysql.Config{
		User:                 c.Database.Username,
		Passwd:               c.Database.Password,
		Net:                  "tcp",
		Addr:                 c.Database.IP + ":" + strconv.Itoa(c.Database.Port),
		DBName:               c.Database.Name,
		AllowNativePasswords: true,
	})
	dr.Service.User.InitViolet(c.Violet)
	r.Handle("/", handler.Playground("BlogReaper", "/api/graphql"))
	r.Handle("/graphql", handler.GraphQL(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: dr,
	}), handler.ResolverMiddleware(resolver.SessionResolverMiddleware)))
	return r
}

func TestApp() http.Handler {
	return App(Config{
		Violet: violetSdk.Config{},
	})
}
