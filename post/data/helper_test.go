package data

import (
	"database/sql"
)

const testConn = "root:Salam1234#@/test"

func getTestDBAddress() string {
	return testConn
}

func prepareTestTables(rowConn *sql.DB) {
	_, err := rowConn.Exec(`
    CREATE TABLE IF NOT EXISTS posts 
    (
        id serial PRIMARY KEY,
		title varchar(100) NOT NULL,
		body varchar(500)  NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`)
	if err != nil {
		panic(err)
	}
}

func truncateTable(rowConn *sql.DB, tableName string) error {
	_, err := rowConn.Exec("TRUNCATE " + tableName)
	return err
}
