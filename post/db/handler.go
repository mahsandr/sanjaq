package db

import (
	"errors"
	"fmt"
)

type Post struct {
	ID        uint64
	Title     string
	Body      string
	CreatedAt int64
}

var errInvalidInputs = errors.New("INVALID_INPUT")

const insertQuery = `
INSERT INTO posts(id,title,body,created_at)
	VALUES(DEFAULT,?,?,DEFAULT)`

// Insert a new post
func (c *conn) Insert(title, body string) (uint64, error) {
	result, err := c.DBConn().Exec(insertQuery, title, body)
	if err != nil {
		return 0, err
	}
	postIDInt, _ := result.LastInsertId()
	return uint64(postIDInt), nil
}

const (
	minLimit = 200
	maxLimit = 1500
)

const selectQuery = `
SELECT id,title,body,UNIX_TIMESTAMP(created_at) 
		FROM posts %s
		order by created_at desc 
		 limit ? offset ?
		`
const fileterPostQuery = " WHERE id in (%s) "

// Get is function that get posts by id
func (c *conn) Get(postIDs []uint64, limit uint16, offset uint64) (posts []*Post, err error) {
	if len(postIDs) == 0 && limit == 0 {
		return nil, errInvalidInputs
	}
	query := selectQuery
	whereCondition := ""
	if len(postIDs) > 0 {
		whereCondition = fmt.Sprintf(fileterPostQuery, convertFormatUintAppend(postIDs))
	}

	params := []interface{}{}

	if limit > maxLimit {
		limit = maxLimit
	} else if limit == 0 {
		limit = minLimit
	}
	params = append(params, limit, offset)

	rows, err := c.DBConn().Query(fmt.Sprintf(query, whereCondition), params...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		post := Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Body, &post.CreatedAt)
		if err != nil {
			return
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

const deleteQuery = "DELETE FROM posts WHERE id in (%s)"

// Insert a new post
func (c *conn) Delete(postIDs []uint64) (err error) {
	if len(postIDs) == 0 {
		return errInvalidInputs
	}

	_, err = c.DBConn().Exec(fmt.Sprintf(deleteQuery, convertFormatUintAppend(postIDs)))
	return err
}
