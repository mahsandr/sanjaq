package post

import (
	"database/sql"
	"encoding/json"
	"sanjaq/post/data"
	"strconv"

	"github.com/goraz/cast"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type errorCode string

var (
	errorCodeEmptyTitle   errorCode = "EMPTY_TITLE"
	errorCodeEmptyBody    errorCode = "EMPTY_BODY"
	errorCodeEmptyPostID  errorCode = "EMPTY_POST_ID"
	errorCodeEmptyLimit   errorCode = "EMPTY_Limit"
	errorCodeServerError  errorCode = "SERVER_ERROR"
	errorCodePostNotFound errorCode = "POST_NOT_FOUND"
)

type Handler struct {
	logger *zap.Logger
	dbConn data.Conn
}

func NewHandler(dbConn data.Conn, logger *zap.Logger) (*Handler, error) {
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
func (h *Handler) NewPost(reqCtx *fasthttp.RequestCtx) {
	response := Response{}
	defer func() {
		payload, err := json.Marshal(&response)
		h.checkError("marshal new post response", err)
		_, err = reqCtx.Write(payload)
		h.checkError("write NewPost response", err)
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
	postID, err := h.dbConn.Insert(title, body)
	if err != nil {
		h.logger.Error("failed to insert post",
			zap.Error(err))
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		response.ErrorCode = errorCodeServerError
		return
	}
	response.Result = postID
}
func (h *Handler) GetPosts(reqCtx *fasthttp.RequestCtx) {
	response := Response{}
	defer func() {
		payload, err := json.Marshal(&response)
		h.checkError("marshal new post response", err)
		_, err = reqCtx.Write(payload)
		h.checkError("write GetPosts response", err)
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

	posts, err := h.dbConn.Get(postIDs, uint16(limit), offset)
	if err != nil {
		h.logger.Error("failed to get post",
			zap.Error(err))
		response.ErrorCode = errorCodeServerError
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	if len(posts) == 0 {
		response.ErrorCode = errorCodePostNotFound
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
	} else {
		go h.countViews(postIDs)
		response.Result = posts
	}
}
func (h *Handler) Del(reqCtx *fasthttp.RequestCtx) {
	response := Response{}
	defer func() {
		payload, err := json.Marshal(&response)
		h.checkError("marshal new post response", err)
		_, err = reqCtx.Write(payload)
		h.checkError("write GetPosts response", err)
	}()
	if reqCtx.UserValue("id") == nil {
		reqCtx.Response.Header.SetStatusCode(fasthttp.StatusBadRequest)
		response.ErrorCode = errorCodeEmptyPostID
		return
	}
	postID := cast.MustUint(reqCtx.UserValue("id"))
	if err := h.dbConn.Delete(postID); err != nil {
		if err == sql.ErrNoRows {
			response.ErrorCode = errorCodePostNotFound
			reqCtx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)
		} else {
			h.logger.Error("failed to delete",
				zap.Error(err))
			response.ErrorCode = errorCodeServerError
			reqCtx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}
}

func (h *Handler) Top(reqCtx *fasthttp.RequestCtx) {

}
func (h *Handler) countViews(postIDs []uint64) {
	// increase redis is atomic

	err := h.dbConn.CountPostVisits(postIDs)
	if err != nil {
		h.logger.Error("failed to count visitor",
			zap.Error(err))
	}
}
