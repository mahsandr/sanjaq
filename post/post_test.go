package post

import (
	"encoding/json"
	"errors"
	"sanjaq/post/db"
	mockdb "sanjaq/post/mockdb"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewPost(t *testing.T) {
	var tests = map[string]struct {
		initMock   func(mockDB *mockdb.MockConn)
		statusCode int
		logEntry   []*zapcore.Entry
		logContext []map[string]interface{}
		title      string
		body       string
		response   Response
	}{
		"success": {
			title:      "title1",
			body:       "test body 1",
			statusCode: fasthttp.StatusOK,
			initMock: func(mockDB *mockdb.MockConn) {
				postID := uint64(1)
				mockDB.EXPECT().
					Insert("title1", "test body 1").
					Return(postID, nil)
			},
			response: Response{
				Result: uint64(1),
			},
		},
		"title is empty": {
			body:       "test body 1",
			statusCode: fasthttp.StatusBadRequest,
			response: Response{
				ErrorCode: errorCodeEmptyTitle,
			},
		},
		"body is empty": {
			title:      "title1",
			statusCode: fasthttp.StatusBadRequest,
			response: Response{
				ErrorCode: errorCodeEmptyBody,
			},
		},
		"error insert": {
			title:      "title1",
			body:       "test body 1",
			statusCode: fasthttp.StatusInternalServerError,
			initMock: func(mockDB *mockdb.MockConn) {
				mockDB.EXPECT().
					Insert("title1", "test body 1").
					Return(uint64(0), errors.New("db error"))
			},
			response: Response{
				ErrorCode: errorCodeServerError,
			},
			logEntry: []*zapcore.Entry{
				{
					Level:   zap.ErrorLevel,
					Message: "failed to insert post",
				},
			},
			logContext: []map[string]interface{}{
				{
					"error": "db error",
				},
			},
		},
	}
	for name, test := range tests {
		Convey(name, t, func() {
			observerlog, out := observer.New(zap.InfoLevel)

			ctrlAccount := gomock.NewController(t)
			defer ctrlAccount.Finish()
			mockDB := mockdb.NewMockConn(ctrlAccount)

			if test.initMock != nil {
				test.initMock(mockDB)
			}

			reqCtx := &fasthttp.RequestCtx{}

			if test.title != "" {
				reqCtx.PostArgs().Set("title", test.title)
			}

			if test.body != "" {
				reqCtx.PostArgs().Set("body", test.body)
			}
			handler, err := NewPostHandler(mockDB, zap.New(observerlog))
			So(err, ShouldBeNil)

			handler.NewPost(reqCtx)
			logs := out.TakeAll()

			resp := Response{}
			err = json.Unmarshal(reqCtx.Response.Body(), &resp)
			So(err, ShouldBeNil)

			So(reqCtx.Response.StatusCode(), ShouldEqual, test.statusCode)
			So(resp.ErrorCode, ShouldEqual, test.response.ErrorCode)
			So(resp.Result, ShouldEqual, test.response.Result)

			if len(test.logEntry) > 0 {
				So(len(logs), ShouldEqual, len(test.logEntry))
				for i, log := range logs {
					So(log.Level, ShouldEqual, test.logEntry[i].Level)
					So(log.Message, ShouldEqual, test.logEntry[i].Message)
					So(log.ContextMap(), ShouldResemble, test.logContext[i])
				}
			}

		})
	}
}
func TestGet(t *testing.T) {
	var tests = map[string]struct {
		initMock   func(mockDB *mockdb.MockConn)
		statusCode int
		logEntry   []*zapcore.Entry
		logContext []map[string]interface{}
		postID     uint64
		limit      string
		offset     string
		response   Response
	}{
		"success": {
			postID:     1,
			statusCode: fasthttp.StatusOK,
			initMock: func(mockDB *mockdb.MockConn) {
				postID := uint64(1)
				mockDB.EXPECT().
					Get([]uint64{postID}, uint16(0), uint64(0)).
					Return([]*db.Post{
						{
							ID:        1,
							Title:     "test1",
							Body:      "test",
							CreatedAt: 19119919,
						}}, nil)
			},
			response: Response{
				Result: []interface{}{map[string]interface{}{"Body": "test", "CreatedAt": 1.9119919e+07, "ID": float64(1), "Title": "test1"}},
			},
		},
		"id is null": {
			statusCode: fasthttp.StatusBadRequest,
			response: Response{
				ErrorCode: errorCodeEmptyLimit,
			},
		},
		"limit": {
			limit:      "10",
			offset:     "5",
			statusCode: fasthttp.StatusOK,
			initMock: func(mockDB *mockdb.MockConn) {
				limit := uint16(10)
				offset := uint64(5)
				var postIDs []uint64
				mockDB.EXPECT().
					Get(postIDs, limit, offset).
					Return([]*db.Post{
						{
							ID:        1,
							Title:     "test1",
							Body:      "test",
							CreatedAt: 19119919,
						}}, nil)
			},
			response: Response{
				Result: []interface{}{map[string]interface{}{"Body": "test", "CreatedAt": 1.9119919e+07, "ID": float64(1), "Title": "test1"}},
			},
		},
		"db error": {
			postID:     1,
			statusCode: fasthttp.StatusInternalServerError,
			initMock: func(mockDB *mockdb.MockConn) {
				limit := uint16(0)
				offset := uint64(0)
				var postIDs = []uint64{1}
				mockDB.EXPECT().
					Get(postIDs, limit, offset).
					Return(nil, errors.New("db error"))
			},
			response: Response{
				ErrorCode: errorCodeServerError,
			},
		},
	}
	for name, test := range tests {
		Convey(name, t, func() {
			observerlog, out := observer.New(zap.InfoLevel)

			ctrlAccount := gomock.NewController(t)
			defer ctrlAccount.Finish()
			mockDB := mockdb.NewMockConn(ctrlAccount)

			if test.initMock != nil {
				test.initMock(mockDB)
			}

			reqCtx := &fasthttp.RequestCtx{}

			if test.postID != 0 {
				reqCtx.SetUserValue("id", test.postID)
			}

			if test.limit != "" {
				reqCtx.URI().QueryArgs().Set("limit", test.limit)
			}
			if test.offset != "" {
				reqCtx.URI().QueryArgs().Set("offset", test.offset)
			}

			handler, err := NewPostHandler(mockDB, zap.New(observerlog))
			So(err, ShouldBeNil)

			handler.Get(reqCtx)
			logs := out.TakeAll()

			resp := Response{}
			err = json.Unmarshal(reqCtx.Response.Body(), &resp)
			So(err, ShouldBeNil)

			So(reqCtx.Response.StatusCode(), ShouldEqual, test.statusCode)
			So(resp.ErrorCode, ShouldEqual, test.response.ErrorCode)
			So(resp.Result, ShouldResemble, test.response.Result)

			if len(test.logEntry) > 0 {
				So(len(logs), ShouldEqual, len(test.logEntry))
				for i, log := range logs {
					So(log.Level, ShouldEqual, test.logEntry[i].Level)
					So(log.Message, ShouldEqual, test.logEntry[i].Message)
					So(log.ContextMap(), ShouldResemble, test.logContext[i])
				}
			}

		})
	}
}
