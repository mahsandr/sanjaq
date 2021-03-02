package db

//nolint: golint
import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Conn interface {
	DBConn() *sql.DB
	Get(postIDs []uint64, limit uint16, offset uint64) (posts []*Post, err error)
	Insert(title, body string) (uint64, error)
	Delete(postID []uint64) error
}

type conn struct {
	dbConn *sql.DB
}

// NewConn initilise and return new database connection
func NewConn(dataSourceName string) (Conn, error) {
	dbConn, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	dbConn.SetConnMaxLifetime(time.Minute * 2)
	dbConn.SetMaxOpenConns(1000)
	return &conn{
		dbConn: dbConn,
	}, nil
}
func (c *conn) DBConn() *sql.DB {
	return c.dbConn
}
