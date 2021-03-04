//nolint: golint
package data

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInsert(t *testing.T) {
	dbConn, err := NewConn(getTestDBAddress(), "127.0.0.1:6379", "", 1)
	if err != nil {
		panic(err)
	}
	prepareTestTables(dbConn.DBConn())
	if err != nil {
		t.Fatal(err)
	}
	var tests = map[string]struct {
		errWant error
		ctx     func() context.Context
		title   string
		body    string
	}{
		"success": {
			title: "test",
			body:  "test body",
		},
		"error": {
			errWant: errors.New("Error 1054: Unknown column 'title' in 'field list'"),
		},
	}
	for name, test := range tests {
		Convey(name, t, func() {
			err = truncateTable(dbConn.DBConn(), "posts")
			So(err, ShouldBeNil)

			if name == "error" {
				_, err = dbConn.DBConn().Exec("ALTER TABLE posts RENAME COLUMN title TO title1;")
				So(err, ShouldBeNil)
				defer func() {
					dbConn.DBConn().Exec("ALTER TABLE posts RENAME COLUMN title1 TO title;")
					So(err, ShouldBeNil)
				}()
			}
			postID, err := dbConn.Insert(test.title, test.body)
			if test.errWant != nil {
				So(err, ShouldBeError, test.errWant)
			} else {
				So(err, ShouldBeNil)
				post, err := dbConn.Get([]uint64{postID}, 0, 0)
				So(err, ShouldBeNil)
				So(len(post), ShouldBeGreaterThan, 0)
				So(post[0].ID, ShouldEqual, postID)
			}
		})
	}
}
func TestGet(t *testing.T) {
	dbConn, err := NewConn(getTestDBAddress(), "127.0.0.1:6379", "", 1)
	if err != nil {
		panic(err)
	}
	prepareTestTables(dbConn.DBConn())
	if err != nil {
		t.Fatal(err)
	}
	var tests = map[string]struct {
		errWant     error
		ctx         func() context.Context
		title       string
		insertPosts []*Post
		getPosts    []*Post
		postIds     []uint64
		limit       uint16
		offset      uint64
	}{
		"success": {
			insertPosts: []*Post{
				{
					Title: "test1",
					Body:  "test1 body",
				}, {
					Title: "test2",
					Body:  "test2 body",
				},
			},
			getPosts: []*Post{
				{
					Title:     "test1",
					Body:      "test1 body",
					CreatedAt: time.Now().Unix(),
				}, {
					Title:     "test2",
					Body:      "test2 body",
					CreatedAt: time.Now().Unix(),
				},
			},
			limit:  0,
			offset: 0,
		},
		"limit": {
			insertPosts: []*Post{
				{
					Title: "test1",
					Body:  "test1 body",
				}, {
					Title: "test2",
					Body:  "test2 body",
				},
			},
			getPosts: []*Post{
				{
					Title:     "test2",
					Body:      "test2 body",
					CreatedAt: time.Now().Unix(),
				},
			},
			limit:  1,
			offset: 0,
		},
		"not found": {
			insertPosts: []*Post{
				{
					Title: "test3",
					Body:  "test3 body",
				}, {
					Title: "test4",
					Body:  "test4 body",
				},
			},
			postIds: []uint64{12},
		},
		"postIds & limit are zero": {
			limit:   0,
			offset:  0,
			errWant: errInvalidInputs,
		},
		"error": {
			postIds: []uint64{1},
			errWant: errors.New("Error 1054: Unknown column 'title' in 'field list'"),
		},
	}
	for testName, test := range tests {
		Convey(testName, t, func() {
			err = truncateTable(dbConn.DBConn(), "posts")
			So(err, ShouldBeNil)
			var postID uint64
			for _, post := range test.insertPosts {
				postID, err = dbConn.Insert(post.Title, post.Body)
				So(err, ShouldBeNil)
				for j, p := range test.getPosts {
					if p.Title == post.Title {
						test.getPosts[j].ID = postID
						test.postIds = append(test.postIds, postID)
					}
				}
			}
			if testName == "error" {
				_, err = dbConn.DBConn().Exec("ALTER TABLE posts RENAME COLUMN title TO title1;")
				So(err, ShouldBeNil)
				defer func() {
					dbConn.DBConn().Exec("ALTER TABLE posts RENAME COLUMN title1 TO title;")
					So(err, ShouldBeNil)
				}()
			}
			posts, err := dbConn.Get(test.postIds, test.limit, test.offset)
			if test.errWant != nil {
				So(err, ShouldBeError, test.errWant)
			} else {
				So(err, ShouldBeNil)
				So(posts, ShouldResemble, test.getPosts)
			}
		})
	}
}
func TestDelete(t *testing.T) {
	dbConn, err := NewConn(getTestDBAddress(), "127.0.0.1:6379", "", 1)
	if err != nil {
		panic(err)
	}
	prepareTestTables(dbConn.DBConn())
	if err != nil {
		t.Fatal(err)
	}
	var tests = map[string]struct {
		errWant    error
		ctx        func() context.Context
		title      string
		insertPost *Post
	}{
		"success": {
			insertPost: &Post{
				Title: "test1",
				Body:  "test1 body",
			},
		},
		"invalid_input": {
			errWant: errInvalidInputs,
		},
	}
	for testName, test := range tests {
		Convey(testName, t, func() {
			err = truncateTable(dbConn.DBConn(), "posts")
			So(err, ShouldBeNil)

			var (
				postIDs []uint64
				postID  uint64
			)

			if test.insertPost != nil {
				postID, err = dbConn.Insert(test.insertPost.Title, test.insertPost.Body)
				So(err, ShouldBeNil)
				postIDs = append(postIDs, postID)
			}

			err = dbConn.Delete(postIDs)
			if test.errWant != nil {
				So(err, ShouldBeError, test.errWant)
			} else {
				So(err, ShouldBeNil)
				post, err := dbConn.Get(postIDs, 0, 0)
				So(err, ShouldBeNil)
				So(post, ShouldBeNil)
			}
		})
	}
}
