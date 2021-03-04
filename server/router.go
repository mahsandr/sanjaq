package server

import (
	"github.com/fasthttp/router"
)

func (s *Server) HandleRouter(r *router.Router) bool {
	r.POST("/post", s.postHandler.NewPost)
	r.GET("/posts/{id?}", s.postHandler.GetPosts)
	r.GET("/posts/top10", s.postHandler.Top)
	r.DELETE("/posts/{id}", s.postHandler.Del)
	return true
}
