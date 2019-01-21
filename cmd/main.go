package main

import (
	"fmt"
	API "forum/api"
	"time"

	"github.com/buaazp/fasthttprouter"

	"github.com/valyala/fasthttp"
)

// ./tech-db-forum func -u http://localhost:5000/api -r report.html
// pgbadger --prefix '%t [%p]: [%l-1] ' -f stderr /usr/local/var/log/postgres.log -o /users/ksenia/Desktop/postgres.html

func timer(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		startTime := time.Now()
		h(ctx)
		duration := time.Now().Sub(startTime)
		fmt.Println(float64(duration)/float64(time.Millisecond), string(ctx.RequestURI()))

	})
}

func main() {
	r := fasthttprouter.New()
	r.POST("/api/user/:nickname/create", timer(API.CreateUser))
	r.GET("/api/user/:nickname/profile", timer(API.InfoUser))
	r.POST("/api/user/:nickname/profile", timer(API.UpdateUser))

	r.GET("/api/forum/:slug/details", timer(API.InfoForum))
	r.GET("/api/forum/:slug/users", timer(API.UsersForum))

	r.POST("/api/forum/:slug/create", timer(API.CreateThread))
	r.GET("/api/forum/:slug/threads", timer(API.Threads))
	r.POST("/api/thread/:slug_or_id/create", timer(API.CreatePosts))
	r.POST("/api/thread/:slug_or_id/vote", timer(API.Vote))
	r.GET("/api/thread/:slug_or_id/details", timer(API.ThreadInfo))
	r.GET("/api/thread/:slug_or_id/posts", timer(API.SortPosts))
	r.POST("/api/thread/:slug_or_id/details", timer(API.UpdateThread))

	r.GET("/api/post/:id/details", timer(API.InfoPost))
	r.POST("/api/post/:id/details", timer(API.UpdatePost))

	r.POST("/api/service/clear", timer(API.Clear))
	r.GET("/api/service/status", timer(API.Status))

	fasthttp.ListenAndServe(":5000", wrapper(r))
}

func wrapper(router *fasthttprouter.Router) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if path == "/api/forum/create" {
			API.CreateForum(ctx)

			return
		}
		router.Handler(ctx)
	}
}
