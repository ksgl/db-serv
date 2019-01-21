package api

import (
	"forum/models"
	"strconv"
	"strings"

	"github.com/jackc/pgx/pgtype"

	"github.com/valyala/fasthttp"
)

const (
	postInfoSelect = `SELECT thread_id,message,edited,created,forum_slug,author
						FROM posts
						WHERE id=$1;`

	updatePost = `UPDATE posts
					SET message=$1,edited=message<>$1
					WHERE id=$2
					RETURNING author,created,forum_slug,edited,message,thread_id;`

	selectPost = `SELECT COALESCE(parent_id,0),thread_id,message,edited,created,forum_slug,author
					FROM posts
					WHERE id=$1;`
)

const (
	r = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author
	FROM posts p
	WHERE p.id=$1;`

	rUTF = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			u.nickname,u.about,u.fullname,u.email,
			t.author,t.created,t.votes,t.id,t.title,t.message,COALESCE(t.slug,''),t.forum,
			f.slug,f.threads,f.title,f.posts,f."user"
	FROM posts p
	LEFT JOIN users AS u
		ON u.nickname=p.author
	LEFT JOIN threads AS t
		ON t.id=p.thread_id
	LEFT JOIN forums AS f
		ON f.slug=p.forum_slug
	WHERE p.id=$1;`

	rUT = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			u.nickname,u.about,u.fullname,u.email,
			t.author,t.created,t.votes,t.id,t.title,t.message,COALESCE(t.slug,''),t.forum
	FROM posts p
	LEFT JOIN users AS u
		ON u.nickname=p.author
	LEFT JOIN threads AS t
		ON t.id=p.thread_id
	WHERE p.id=$1;`

	rUF = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			u.nickname,u.about,u.fullname,u.email,
			f.slug,f.threads,f.title,f.posts,f."user"
	FROM posts p
	LEFT JOIN users AS u
		ON u.nickname=p.author
	LEFT JOIN forums AS f
		ON f.slug=p.forum_slug
	WHERE p.id=$1;`

	rU = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			u.nickname,u.about,u.fullname,u.email
	FROM posts p
	LEFT JOIN users AS u
		ON u.nickname=p.author
	WHERE p.id=$1;`

	rT = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			t.author,t.created,t.votes,t.id,t.title,t.message,COALESCE(t.slug,''),t.forum
	FROM posts p
	LEFT JOIN threads AS t
		ON t.id=p.thread_id
	WHERE p.id=$1;`

	rTF = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			t.author,t.created,t.votes,t.id,t.title,t.message,COALESCE(t.slug,''),t.forum,
			f.slug,f.threads,f.title,f.posts,f."user"
	FROM posts p
	LEFT JOIN threads AS t
		ON t.id=p.thread_id
	LEFT JOIN forums AS f
		ON f.slug=p.forum_slug
	WHERE p.id=$1;`

	rF = `
	SELECT COALESCE(p.parent_id,0),p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,
			f.slug,f.threads,f.title,f.posts,f."user"
	FROM posts p
	LEFT JOIN forums AS f
		ON f.slug=p.forum_slug
	WHERE p.id=$1;`
)

func InfoPost(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))
	related := ctx.QueryArgs().Peek("related")
	var params []string
	if string(related) != "" {
		params = strings.Split(string(related), ",")
	}

	var (
		uRel  = false
		fRel  = false
		thRel = false
	)

	for _, param := range params {
		switch param {
		case "user":
			uRel = true
		case "forum":
			fRel = true
		case "thread":
			thRel = true
		}
	}

	if len(params) == 0 {
		post := &models.PostPost{}
		post.Post.ID = int64(id)
		db.QueryRow(selectPost, id).Scan(&post.Post.Parent, &post.Post.Thread, &post.Post.Message, &post.Post.IsEdited, &post.Post.Created, &post.Post.Forum, &post.Post.Author)

		if post.Post.Author == "" {
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		p, _ := post.MarshalJSON()
		Respond(ctx, fasthttp.StatusOK, p)

		return
	} else {
		// var query strings.Builder
		// var from strings.Builder
		// fmt.Fprintf(&query, `SELECT p.id,p.parent_id,p.thread_id,p.message,p.edited,p.created,p.forum_slug,p.author,`)
		// fmt.Fprintf(&from, ` FROM posts AS p `)
		// if uRel {
		// 	fmt.Fprintf(&query, `u.nickname,u.about,u.fullname,u.email,`)
		// 	fmt.Fprintf(&from, `LEFT JOIN users AS u
		// 					ON u.nickname=p.author AND true=$1 `)
		// } else {
		// 	fmt.Fprintf(&query, `'','','','',`)
		// 	fmt.Fprintf(&from, `LEFT JOIN users AS u
		// 					ON u.nickname=p.author AND true=$1 `)
		// }
		// if thRel {
		// 	fmt.Fprintf(&query, `t.author,t.created,t.votes,t.id,t.title,t.message,COALESCE(t.slug,''),t.forum,`)
		// 	fmt.Fprintf(&from, `LEFT JOIN threads AS t
		// 					ON t.id=p.thread_id  AND true=$2 `)
		// } else {
		// 	fmt.Fprintf(&query, `'',NULL,0,0,'','','','',`)
		// 	fmt.Fprintf(&from, `LEFT JOIN threads AS t
		// 					ON t.id=p.thread_id  AND true=$2 `)
		// }
		// if fRel {
		// 	fmt.Fprintf(&query, `f.slug,f.threads,f.title,f.posts,f."user"`)
		// 	fmt.Fprintf(&from, `LEFT JOIN forums AS f
		// 					ON f.slug=p.forum_slug AND true=$3 `)
		// } else {
		// 	fmt.Fprintf(&query, `'',0,'',0,''`)
		// 	fmt.Fprintf(&from, `LEFT JOIN forums AS f
		// 					ON f.slug=p.forum_slug AND true=$3 `)
		// }
		// fmt.Fprintf(&from, ` WHERE p.id=$4`)
		// fmt.Fprintf(&query, from.String())

		postRel := &models.Post{}
		authorRel := &models.User{}
		threadRel := &models.Thread{}
		forumRel := &models.Forum{}
		time := &pgtype.Timestamptz{}
		// db.QueryRow(query.String(), uRel, thRel, fRel, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
		// 	&authorRel.Nickname, &authorRel.About, &authorRel.Fullname, &authorRel.Email,
		// 	&threadRel.Author, time, &threadRel.Votes, &threadRel.ID, &threadRel.Title, &threadRel.Message, &threadRel.Slug, &threadRel.Forum,
		// 	&forumRel.Slug, &forumRel.Threads, &forumRel.Title, &forumRel.Posts, &forumRel.User)

		if uRel {
			if thRel {
				if fRel {
					db.QueryRow(rUTF, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&authorRel.Nickname, &authorRel.About, &authorRel.Fullname, &authorRel.Email,
						&threadRel.Author, time, &threadRel.Votes, &threadRel.ID, &threadRel.Title, &threadRel.Message, &threadRel.Slug, &threadRel.Forum,
						&forumRel.Slug, &forumRel.Threads, &forumRel.Title, &forumRel.Posts, &forumRel.User)
				} else if !fRel {
					db.QueryRow(rUT, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&authorRel.Nickname, &authorRel.About, &authorRel.Fullname, &authorRel.Email,
						&threadRel.Author, time, &threadRel.Votes, &threadRel.ID, &threadRel.Title, &threadRel.Message, &threadRel.Slug, &threadRel.Forum)
				}
			} else if !thRel {
				if fRel {
					db.QueryRow(rUF, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&authorRel.Nickname, &authorRel.About, &authorRel.Fullname, &authorRel.Email,
						&forumRel.Slug, &forumRel.Threads, &forumRel.Title, &forumRel.Posts, &forumRel.User)
				} else {
					db.QueryRow(rU, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&authorRel.Nickname, &authorRel.About, &authorRel.Fullname, &authorRel.Email)
				}
			}
		} else if !uRel {
			if thRel {
				if fRel {
					db.QueryRow(rTF, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&threadRel.Author, time, &threadRel.Votes, &threadRel.ID, &threadRel.Title, &threadRel.Message, &threadRel.Slug, &threadRel.Forum,
						&forumRel.Slug, &forumRel.Threads, &forumRel.Title, &forumRel.Posts, &forumRel.User)
				} else if !fRel {
					db.QueryRow(rT, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&threadRel.Author, time, &threadRel.Votes, &threadRel.ID, &threadRel.Title, &threadRel.Message, &threadRel.Slug, &threadRel.Forum)
				}
			} else if !thRel {
				if fRel {
					db.QueryRow(rF, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author,
						&forumRel.Slug, &forumRel.Threads, &forumRel.Title, &forumRel.Posts, &forumRel.User)
				} else {
					db.QueryRow(r, id).Scan(&postRel.Parent, &postRel.Thread, &postRel.Message, &postRel.IsEdited, &postRel.Created, &postRel.Forum, &postRel.Author)
				}
			}
		}

		postRel.ID = int64(id)

		//log.Println(err)
		if postRel.Author == "" {
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		if thRel {
			threadRel.Created = time.Time
		}

		pr := &models.PostRelated{PostRel: postRel}
		if uRel {
			pr.AuthorRel = authorRel
		}
		if thRel {
			pr.ThreadRel = threadRel
		}
		if fRel {
			pr.ForumRel = forumRel
		}

		p, _ := pr.MarshalJSON()
		Respond(ctx, fasthttp.StatusOK, p)

		return
	}

}

func UpdatePost(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))
	update := &models.PostUpdate{}
	update.UnmarshalJSON(ctx.PostBody())

	if update.Message == "" {
		post := &models.Post{}
		db.QueryRow(postInfoSelect, id).Scan(&post.Thread, &post.Message, &post.IsEdited, &post.Created, &post.Forum, &post.Author)

		if post.Author == "" {
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		post.ID = int64(id)

		p, _ := post.MarshalJSON()
		Respond(ctx, fasthttp.StatusOK, p)

		return
	}

	post := &models.Post{}

	db.QueryRow(updatePost, update.Message, id).Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Thread)
	post.ID = int64(id)
	if post.Author == "" {
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := post.MarshalJSON()
	Respond(ctx, fasthttp.StatusOK, p)

	return
}
