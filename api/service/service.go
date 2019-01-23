package service

import (
	ut "forum/api/utility"
	"forum/database"
	"forum/models"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

var db *pgx.ConnPool

func init() {
	db = database.Connect()
	PStruncate, _ = db.Prepare("truncate", truncate)
	PScount, _ = db.Prepare("count", count)
}

var (
	PStruncate *pgx.PreparedStatement
	PScount    *pgx.PreparedStatement
)

const (
	truncate = `TRUNCATE forums,participants,posts,threads,users,votes;`
	count    = `SELECT (SELECT count(*) FROM users), (SELECT count(*) FROM posts), (SELECT count(*) FROM forums), (SELECT count(*) FROM threads);`
)

func Clear(ctx *fasthttp.RequestCtx) {
	db.Exec(PStruncate.Name)

	ut.Respond(ctx, fasthttp.StatusOK, []byte(`[OK]`))

	return
}

func Status(ctx *fasthttp.RequestCtx) {
	status := &models.Status{}

	db.QueryRow(PScount.Name).Scan(&status.User, &status.Post, &status.Forum, &status.Thread)

	p, _ := status.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
