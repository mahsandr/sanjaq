package post

import (
	"encoding/json"
	"sanjaq/post/db"
	"strconv"

	"github.com/goraz/cast"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type errorCode string

var (
	errorCodeEmptyTitle  errorCode = "EMPTY_TITLE"
	errorCodeEmptyBody   errorCode = "EMPTY_BODY"
	errorCodeEmptyLimit  errorCode = "EMPTY_Limit"
	errorCodeServerError errorCode = "SERVER_ERROR"
)

type Handler struct {
	logger *zap.Logger
	dbConn db.Conn
}

func NewHandler(dbConn db.Conn, logger *zap.Logger) (*Handler, error) {
	return &Handler{
		logger: logger,
		dbConn: dbConn,
	}, nil
}

// Response is a structure for responding to post requests
type Response struct {
	ErrorCode errorCode
	Result    interface{}
}

// NewPost is function that insert new post and returns the id
func (p *Handler) NewPost(reqCtx *fasthttp.RequestCtx) {
	response := Response{}
	defer func() {
		payload, err := json.Marshal(&response)
		p.checkError("marshal new post response", err)
		_, err = reqCtx.Write(payload)
		p.checkError("write NewPost response", err)
	}()
	var title string
	if title = string(reqCtx.PostArgs().Peek("title")); title == "" {
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		response.ErrorCode = errorCodeEmptyTitle
		return
	}

	var body string
	if body = string(reqCtx.PostArgs().Peek("body")); body == "" {
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		response.ErrorCode = errorCodeEmptyBody
		return
	}
	postID, err := p.dbConn.Insert(title, body)
	if err != nil {
		p.logger.Error("failed to insert post",
			zap.Error(err))
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		response.ErrorCode = errorCodeServerError
		return
	}
	response.Result = postID
}
func (p *Handler) GetPosts(reqCtx *fasthttp.RequestCtx) {
	response := Response{}
	defer func() {
		payload, err := json.Marshal(&response)
		p.checkError("marshal new post response", err)
		_, err = reqCtx.Write(payload)
		p.checkError("write GetPosts response", err)
	}()
	var (
		postIDs []uint64
		limit   int
		offset  uint64
	)
	if reqCtx.UserValue("id") != nil {
		postIDs = []uint64{cast.MustUint(reqCtx.UserValue("id"))}
	} else {
		var limitStr, offsetStr string

		if limitStr = string(reqCtx.URI().QueryArgs().Peek("limit")); limitStr == "" {
			reqCtx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
			response.ErrorCode = errorCodeEmptyLimit
			return
		}
		limit, _ = strconv.Atoi(limitStr)
		if offsetStr = string(reqCtx.URI().QueryArgs().Peek("offset")); offsetStr != "" {
			offset, _ = strconv.ParseUint(offsetStr, 10, 64)
		}
	}

	posts, err := p.dbConn.Get(postIDs, uint16(limit), offset)
	if err != nil {
		p.logger.Error("failed to get post",
			zap.Error(err))
		response.ErrorCode = errorCodeServerError
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	response.Result = posts
}
func (p *Handler) Top(reqCtx *fasthttp.RequestCtx) {

}
func (p *Handler) Del(reqCtx *fasthttp.RequestCtx) {

}
