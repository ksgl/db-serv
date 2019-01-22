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
}

func Clear(ctx *fasthttp.RequestCtx) {
	db.Exec(`TRUNCATE forums,participants,posts,threads,users,votes;`)

	ut.Respond(ctx, fasthttp.StatusOK, []byte(`[OK]`))

	return
}

func Status(ctx *fasthttp.RequestCtx) {
	status := &models.Status{}

	db.QueryRow(`SELECT
					(SELECT count(*) FROM users),
					(SELECT count(*) FROM posts),
					(SELECT count(*) FROM forums),
					(SELECT count(*) FROM threads);
					`).Scan(&status.User, &status.Post, &status.Forum, &status.Thread)

	p, _ := status.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
