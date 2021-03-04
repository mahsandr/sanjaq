package data

//nolint: golint
import (
	"context"
	"database/sql"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

type Conn interface {
	DBConn() *sql.DB
	Get(postIDs []uint64, limit uint16, offset uint64) (posts []*Post, err error)
	Insert(title, body string) (uint64, error)
	Delete(postID uint64) error
	CountPostVisits(postIDs []uint64) error
}

type conn struct {
	dbConn      *sql.DB
	redisClient *redis.Client
}

// NewConn initilise and return new database connection
func NewConn(dataSourceName string, addr, password string, db int) (Conn, error) {
	dbConn, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	dbConn.SetConnMaxLifetime(time.Minute * 2)
	dbConn.SetMaxOpenConns(1000)

	redisConn := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err = redisConn.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &conn{
		dbConn:      dbConn,
		redisClient: redisConn,
	}, nil
}
func (c *conn) DBConn() *sql.DB {
	return c.dbConn
}
