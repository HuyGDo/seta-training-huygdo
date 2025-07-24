package graphql

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"gorm.io/gorm"
)

// GraphQLHandler creates a gin handler for GraphQL requests.
func GraphQLHandler(db *gorm.DB) gin.HandlerFunc {
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaString, NewResolver(db), opts...)
	handler := &relay.Handler{Schema: schema}

	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler creates a gin handler for the GraphiQL playground.
func PlaygroundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the GraphQL Playground!")
	}
}
