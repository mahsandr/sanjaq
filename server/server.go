package server

import (
	"fmt"
	"sanjaq/logger"
	"sanjaq/post"
	postdata "sanjaq/post/data"

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
	postData, err := postdata.NewConn(config.DataBase.MySQLConn,
		config.RedisConn.Addr,
		config.RedisConn.Password,
		config.RedisConn.DB)
	checkError(log, err)

	err = prepareTables(postData.DBConn())
	checkError(log, err)

	postHandler, err := post.NewHandler(postData, log)
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
