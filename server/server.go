package server

import (
	"fmt"
	"sanjaq/post"
	postdb "sanjaq/post/db"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	post   *post.PostHandler
	config Config
}

func NewServer(logger *zap.Logger) *Server {
	config := ReadFromJSON("./config.json")
	postDB, err := postdb.NewConn(config.DataBase.MySqlConn)
	checkError(logger, err)

	err = prepareTables(postDB.DBConn())
	checkError(logger, err)

	post, err := post.NewPostHandler(postDB, logger)
	return &Server{
		config: config,
		logger: logger,
		post:   post,
	}
}
func (s *Server) Run() {
	r := router.New()
	s.HandleRouter(r)

	fmt.Println("listen on http: " + s.config.Server.Port)
	panic(fasthttp.ListenAndServe(s.config.Server.Port, r.Handler))
}
