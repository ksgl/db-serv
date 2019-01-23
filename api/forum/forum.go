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
	PScreateForumInsert, _ = db.Prepare("createForumInsert", createForumInsert)
	PSforumInfoShortSelect, _ = db.Prepare("forumInfoShortSelect", forumInfoShortSelect)
	PSforumInfoExtendedSelect, _ = db.Prepare("forumInfoExtendedSelect", forumInfoExtendedSelect)
	PSforumSlugSelect, _ = db.Prepare("forumSlugSelect", forumSlugSelect)
	PSdescSlugSinceLimit, _ = db.Prepare("descSlugSinceLimit", descSlugSinceLimit)
	PSdescSlugLimit, _ = db.Prepare("descSlugLimit", descSlugLimit)
	PSdescSlugSince, _ = db.Prepare("descSlugSince", descSlugSince)
	PSdescSlug, _ = db.Prepare("descSlug", descSlug)
	PSascSlugSinceLimit, _ = db.Prepare("ascSlugSinceLimit", ascSlugSinceLimit)
	PSascSlugLimit, _ = db.Prepare("ascSlugLimit", ascSlugLimit)
	PSascSlugSince, _ = db.Prepare("ascSlugSince", ascSlugSince)
	PSascSlug, _ = db.Prepare("ascSlug", ascSlug)
}

var (
	PScreateForumInsert       *pgx.PreparedStatement
	PSforumInfoShortSelect    *pgx.PreparedStatement
	PSforumInfoExtendedSelect *pgx.PreparedStatement
	PSforumSlugSelect         *pgx.PreparedStatement
	PSdescSlugSinceLimit      *pgx.PreparedStatement
	PSdescSlugLimit           *pgx.PreparedStatement
	PSdescSlugSince           *pgx.PreparedStatement
	PSdescSlug                *pgx.PreparedStatement
	PSascSlugSinceLimit       *pgx.PreparedStatement
	PSascSlugLimit            *pgx.PreparedStatement
	PSascSlugSince            *pgx.PreparedStatement
	PSascSlug                 *pgx.PreparedStatement
)

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

const (
	citext = `SELECT u.nickname,u.about,u.fullname,u.email
				FROM participants AS p
				JOIN users AS u
				ON u.nickname = p.nickname
				WHERE p.forum_slug = $1`
)

func CreateForum(ctx *fasthttp.RequestCtx) {
	f := &models.ForumTrunc{}
	f.UnmarshalJSON(ctx.PostBody())

	err := db.QueryRow(PScreateForumInsert.Name, f.Slug, f.Title, f.User).Scan(&f.Slug, &f.Title, &f.User)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"forums_pkey\" (SQLSTATE 23505)" {
			db.QueryRow(PSforumInfoShortSelect.Name, f.Slug).Scan(&f.Slug, &f.Title, &f.User)

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

	db.QueryRow(PSforumInfoExtendedSelect.Name, f.Slug).Scan(&f.Slug, &f.Title, &f.User, &f.Posts, &f.Threads)

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
	db.QueryRow(PSforumSlugSelect.Name, slug).Scan(&obtainedSlug)

	if obtainedSlug == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	var rows *pgx.Rows
	if desc {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(PSdescSlugSinceLimit.Name, slug, since, limit)
			} else {
				rows, _ = db.Query(PSdescSlugLimit.Name, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(PSdescSlugSince.Name, slug, since)
			} else {
				rows, _ = db.Query(PSdescSlug.Name, slug)
			}
		}
	} else {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(PSascSlugSinceLimit.Name, slug, since, limit)
			} else {
				rows, _ = db.Query(PSascSlugLimit.Name, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(PSascSlugSince.Name, slug, since)
			} else {
				rows, _ = db.Query(PSascSlug.Name, slug)
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

	// var query strings.Builder
	// fmt.Fprint(&query, citext)
	// if since != "" {
	// 	if desc {
	// 		fmt.Fprint(&query, " AND p.nickname < $2")
	// 	} else {
	// 		fmt.Fprint(&query, " AND p.nickname > $2 ")
	// 	}
	// } else {
	// 	fmt.Fprint(&query, " AND $2 = ''")
	// }
	// if desc {
	// 	fmt.Fprint(&query, " ORDER BY p.nickname DESC")
	// } else {
	// 	fmt.Fprint(&query, " ORDER BY p.nickname ASC")
	// }
	// if limit > 0 {
	// 	fmt.Fprint(&query, " LIMIT $3")
	// } else {
	// 	fmt.Fprint(&query, " LIMIT 100000+$3")
	// }
	// rows, _ := db.Query(query.String(), obtainedSlug, since, limit)

	// users := make(models.UsersArr, 0, limit)
	// for rows.Next() {
	// 	user := models.User{}
	// 	rows.Scan(&user.Nickname, &user.About, &user.Fullname, &user.Email)
	// 	users = append(users, &user)
	// }
	// rows.Close()

	p, _ := users.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
