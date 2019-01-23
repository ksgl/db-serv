package forum

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

const (
	createForumInsert = `INSERT INTO forums(slug,title,"user")
						VALUES($1,$2,(SELECT nickname FROM users WHERE nickname=$3))
						RETURNING slug,title,"user";`

	forumInfoShortSelect = `SELECT slug,title,"user"
							FROM forums
							WHERE slug=$1;`

	forumInfoExtendedSelect = `SELECT slug,title,"user",posts,threads
								FROM forums
								WHERE slug=$1;`

	forumSlugSelect = `SELECT slug
						FROM forums
						WHERE slug=$1;`

	sqlSelectUserForum = `
			SELECT u.nickname, u.about, u.fullname, u.email
			FROM participants AS uf
			JOIN users AS u ON u.nickname = uf.nickname
			WHERE uf.forum_slug = $1
	`

	descSlugSinceLimit = `SELECT u.nickname,u.about,u.fullname,u.email
							FROM
							"users" u JOIN
							"participants" p ON (u.id = p.id)
							WHERE
							p.forum_slug = $1 AND
							p.nickname < $2
							ORDER BY p.nickname DESC
							LIMIT $3;`

	descSlugLimit = `SELECT u.nickname,u.about,u.fullname,u.email
						FROM
						"users" u JOIN
						"participants" p ON (u.id = p.id)
						WHERE
						p.forum_slug = $1
						ORDER BY p.nickname DESC
						LIMIT $2;`

	descSlugSince = `SELECT u.nickname,u.about,u.fullname,u.email
					FROM
					"users" u JOIN
					"participants" p ON (u.id = p.id)
					WHERE
					p.forum_slug = $1 AND
					p.nickname < $2
					ORDER BY p.nickname DESC;`

	descSlug = `SELECT u.nickname,u.about,u.fullname,u.email
				FROM
				"users" u JOIN
				"participants" p ON (u.id = p.id)
				WHERE
				p.forum_slug = $1
				ORDER BY p.nickname DESC;`

	ascSlugSinceLimit = `SELECT u.nickname,u.about,u.fullname,u.email
							FROM
							"users" u JOIN
							"participants" p ON (u.id = p.id)
							WHERE
							p.forum_slug = $1 AND
							p.nickname > $2
							ORDER BY p.nickname ASC
							LIMIT $3;`

	ascSlugLimit = `SELECT u.nickname,u.about,u.fullname,u.email
					FROM
					"users" u JOIN
					"participants" p ON (u.id = p.id)
					WHERE
					p.forum_slug = $1
					ORDER BY p.nickname ASC
					LIMIT $2;`

	ascSlugSince = `SELECT u.nickname,u.about,u.fullname,u.email
					FROM
					"users" u JOIN
					"participants" p ON (u.id = p.id)
					WHERE
					p.forum_slug = $1 AND
					p.nickname > $2
					ORDER BY p.nickname ASC;`

	ascSlug = `SELECT u.nickname,u.about,u.fullname,u.email
				FROM
				"users" u JOIN
				"participants" p ON (u.id = p.id)
				WHERE
				p.forum_slug = $1
				ORDER BY p.nickname ASC`
)

func CreateForum(ctx *fasthttp.RequestCtx) {
	f := &models.ForumTrunc{}
	f.UnmarshalJSON(ctx.PostBody())

	err := db.QueryRow(createForumInsert, f.Slug, f.Title, f.User).Scan(&f.Slug, &f.Title, &f.User)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"forums_pkey\" (SQLSTATE 23505)" {
			db.QueryRow(forumInfoShortSelect, f.Slug).Scan(&f.Slug, &f.Title, &f.User)

			p, _ := f.MarshalJSON()
			ut.Respond(ctx, fasthttp.StatusConflict, p)

			return
		}

		if err.Error() == "ERROR: null value in column \"user\" violates not-null constraint (SQLSTATE 23502)" {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}
	}

	p, _ := f.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusCreated, p)

	return
}

func InfoForum(ctx *fasthttp.RequestCtx) {
	f := models.Forum{}
	f.Slug = ctx.UserValue("slug").(string)

	db.QueryRow(forumInfoExtendedSelect, f.Slug).Scan(&f.Slug, &f.Title, &f.User, &f.Posts, &f.Threads)

	if f.Title == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := f.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func UsersForum(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := string(ctx.QueryArgs().Peek("since"))

	obtainedSlug := ""
	db.QueryRow(forumSlugSelect, slug).Scan(&obtainedSlug)

	if obtainedSlug == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	var rows *pgx.Rows
	if desc {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(descSlugSinceLimit, slug, since, limit)
			} else {
				rows, _ = db.Query(descSlugLimit, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(descSlugSince, slug, since)
			} else {
				rows, _ = db.Query(descSlug, slug)
			}
		}
	} else {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(ascSlugSinceLimit, slug, since, limit)
			} else {
				rows, _ = db.Query(ascSlugLimit, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(ascSlugSince, slug, since)
			} else {
				rows, _ = db.Query(ascSlug, slug)
			}
		}
	}

	users := make(models.UsersArr, 0, limit)
	for rows.Next() {
		user := models.User{}
		rows.Scan(&user.Nickname, &user.About, &user.Fullname, &user.Email)
		users = append(users, &user)
	}
	rows.Close()

	p, _ := users.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
