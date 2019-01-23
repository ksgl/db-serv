package main

import (
	"forum/api/forum"
	"forum/api/post"
	"forum/api/service"
	"forum/api/thread"
	"forum/api/user"

	"github.com/buaazp/fasthttprouter"

	"github.com/valyala/fasthttp"
)

func main() {
	r := fasthttprouter.New()
	r.POST("/api/user/:nickname/create", user.CreateUser)
	r.GET("/api/user/:nickname/profile", user.InfoUser)
	r.POST("/api/user/:nickname/profile", user.UpdateUser)

	r.GET("/api/forum/:slug/details", forum.InfoForum)
	r.GET("/api/forum/:slug/users", forum.UsersForum)

	r.POST("/api/forum/:slug/create", thread.CreateThread)
	r.GET("/api/forum/:slug/threads", thread.Threads)
	r.POST("/api/thread/:slug_or_id/create", thread.CreatePosts)
	r.POST("/api/thread/:slug_or_id/vote", thread.Vote)
	r.GET("/api/thread/:slug_or_id/details", thread.ThreadInfo)
	r.GET("/api/thread/:slug_or_id/posts", thread.SortPosts)
	r.POST("/api/thread/:slug_or_id/details", thread.UpdateThread)

	r.GET("/api/post/:id/details", post.InfoPost)
	r.POST("/api/post/:id/details", post.UpdatePost)

	r.POST("/api/service/clear", service.Clear)
	r.GET("/api/service/status", service.Status)

	fasthttp.ListenAndServe(":5000", wrapper(r))
}

func wrapper(router *fasthttprouter.Router) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if path == "/api/forum/create" {
			forum.CreateForum(ctx)

			return
		}
		router.Handler(ctx)
	}
}
