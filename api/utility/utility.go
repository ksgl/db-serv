package utility

import (
	"github.com/valyala/fasthttp"
)

// type Handler struct {
// 	db *pgx.ConnPool
// }

// func New(db *pgx.ConnPool) *Handler {
// 	return &Handler{db}
// }

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
