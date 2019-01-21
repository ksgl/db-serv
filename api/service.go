package api

import (
	"forum/models"

	"github.com/valyala/fasthttp"
)

func Clear(ctx *fasthttp.RequestCtx) {
	db.Exec(`TRUNCATE forums,participants,posts,threads,users,votes;`)

	Respond(ctx, fasthttp.StatusOK, []byte(`[OK]`))

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
	Respond(ctx, fasthttp.StatusOK, p)

	return
}
