package main

import (
	API "forum/api"

	"github.com/buaazp/fasthttprouter"

	"github.com/valyala/fasthttp"
)

// ./tech-db-forum func -u http://localhost:5000/api -r report.html
// pgbadger --prefix '%t [%p]: [%l-1] ' -f stderr /usr/local/var/log/postgres.log -o /users/ksenia/Desktop/postgres.html

// func  h fasthttp.RequestHandler) fasthttp.RequestHandler {
// 	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

// 		startTime := time.Now()
// 		h(ctx)
// 		duration := time.Now().Sub(startTime)
// 		fmt.Println(float64(duration)/float64(time.Millisecond), string(ctx.RequestURI()))

// 	})
// }

func main() {
	r := fasthttprouter.New()
	r.POST("/api/user/:nickname/create", API.CreateUser)
	r.GET("/api/user/:nickname/profile", API.InfoUser)
	r.POST("/api/user/:nickname/profile", API.UpdateUser)

	r.GET("/api/forum/:slug/details", API.InfoForum)
	r.GET("/api/forum/:slug/users", API.UsersForum)

	r.POST("/api/forum/:slug/create", API.CreateThread)
	r.GET("/api/forum/:slug/threads", API.Threads)
	r.POST("/api/thread/:slug_or_id/create", API.CreatePosts)
	r.POST("/api/thread/:slug_or_id/vote", API.Vote)
	r.GET("/api/thread/:slug_or_id/details", API.ThreadInfo)
	r.GET("/api/thread/:slug_or_id/posts", API.SortPosts)
	r.POST("/api/thread/:slug_or_id/details", API.UpdateThread)

	r.GET("/api/post/:id/details", API.InfoPost)
	r.POST("/api/post/:id/details", API.UpdatePost)

	r.POST("/api/service/clear", API.Clear)
	r.GET("/api/service/status", API.Status)

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
