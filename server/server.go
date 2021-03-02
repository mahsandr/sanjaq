package server

import (
	"fmt"
	"sanjaq/logger"
	"sanjaq/post"
	postdb "sanjaq/post/db"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	logger      *zap.Logger
	postHandler *post.Handler
	config      Config
}

func NewServer() *Server {
	log := logger.InitLog()

	config := ReadFromJSON("./config.json")
	postDB, err := postdb.NewConn(config.DataBase.MySQLConn)
	checkError(log, err)

	err = prepareTables(postDB.DBConn())
	checkError(log, err)

	postHandler, err := post.NewHandler(postDB, log)
	checkError(log, err)
	return &Server{
		config:      config,
		logger:      log,
		postHandler: postHandler,
	}
}
func (s *Server) Run() {
	r := router.New()
	s.HandleRouter(r)

	fmt.Println("listen on http: " + s.config.Server.Port)
	panic(fasthttp.ListenAndServe(s.config.Server.Port, r.Handler))
}
