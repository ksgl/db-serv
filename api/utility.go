package api

import (
	"forum/database"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

// type Handler struct {
// 	db *pgx.ConnPool
// }

// func New(db *pgx.ConnPool) *Handler {
// 	return &Handler{db}
// }

var db *pgx.ConnPool

func init() {
	db = database.Connect()
}

func ErrRespond(ctx *fasthttp.RequestCtx, status int) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody([]byte(`{}`))
}

func Respond(ctx *fasthttp.RequestCtx, status int, payload []byte) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(payload)
}
