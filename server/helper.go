package server

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	DataBase struct {
		MySqlConn string `json:"mysqlconn"`
	} `json:"database"`
	RedisConn struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redisconn"`
}

// ReadFromJSON read the Config from a JSON file.
func ReadFromJSON(path string) Config {
	jsonByte, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	var config Config

	err = json.Unmarshal(jsonByte, &config)

	if err != nil {
		panic(err)
	}

	return config
}
func checkError(logger *zap.Logger, err error) {
	if err != nil {
		logger.Fatal("failed to connect to ad database",
			zap.Error(err))
	}
}

func prepareTables(rowConn *pgxpool.Pool) (err error) {
	_, err = rowConn.Exec(context.Background(), `BEGIN;
    CREATE TABLE IF NOT EXISTS posts 
    (
        id serial PRIMARY KEY,
		title varchar(100) NOT NULL,
		body varchar(500)  NOT NULL,
		created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);
		COMMIT;`)
	return
}
