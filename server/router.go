package server

import (
	"github.com/fasthttp/router"
)

func (s *Server) HandleRouter(r *router.Router) bool {
	r.POST("/post", s.postHandler.NewPost)
	r.GET("/posts/{id?}", s.postHandler.GetPosts)
	r.GET("/topposts/{count?}", s.postHandler.Top)
	r.DELETE("/post/{id}", s.postHandler.Del)
	return true
}
