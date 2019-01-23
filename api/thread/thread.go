package thread

import (
	"fmt"
	ut "forum/api/utility"
	"forum/database"
	"forum/models"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

var db *pgx.ConnPool

func init() {
	db = database.Connect()
}

const selectPostsFlatLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY created, p.id
	LIMIT $2
`

const selectPostsFlatLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY created DESC, p.id DESC
	LIMIT $2
`

const selectPostsFlatLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.id > $2
	ORDER BY created, p.id
	LIMIT $3
`
const selectPostsFlatLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.id < $2
	ORDER BY created DESC, p.id DESC
	LIMIT $3
`

const selectPostsTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY p.path
	LIMIT $2
`

const selectPostsTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY path DESC
	LIMIT $2
`

const selectPostsTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and (p.path > (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path
	LIMIT $3
`

const selectPostsTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and (p.path < (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path DESC
	LIMIT $3
`

const selectPostsParentTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id IS NULL
		ORDER BY p2.path
		LIMIT $3
	)
	ORDER BY path
`

const selectPostsParentTreeLimitDescByID = `
SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
FROM posts p
WHERE p.thread_id = $1 and p.path[1] IN (
    SELECT p2.path[1]
    FROM posts p2
	WHERE p2.parent_id IS NULL and p2.thread_id = $2
	ORDER BY p2.path DESC
    LIMIT $3
)
ORDER BY p.path[1] DESC, p.path
`

const selectPostsParentTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id IS NULL and p2.path[1] > (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path
		LIMIT $4
	)
	ORDER BY p.path
`

const selectPostsParentTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id IS NULL and p2.path[1] < (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path DESC
		LIMIT $4
	)
	ORDER BY p.path[1] DESC, p.path
`

const sqlGetPostsFlat = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE thread_id = $1
	`

const sqlGetPostsParentTree = `
	SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
	FROM posts p
	WHERE path[1] IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id IS NULL
`

const (
	selDescLimitSince = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
					FROM threads
					WHERE forum=$1 AND created <= $2 ORDER BY created DESC LIMIT $3;`

	selDescLimit = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
				FROM threads
				WHERE forum=$1 ORDER BY created DESC LIMIT $2;`

	selDescSince = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
			FROM threads
			WHERE forum=$1 AND created <= $2 ORDER BY created DESC;`

	selDesc = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
			FROM threads
			WHERE forum=$1 ORDER BY created DESC;`

	selAscLimitSince = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
					FROM threads
					WHERE forum=$1 AND created >= $2 ORDER BY created ASC LIMIT $3;`

	selAscLimit = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
				FROM threads
				WHERE forum=$1 ORDER BY created ASC LIMIT $2;`

	selAscSince = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
				FROM threads
				WHERE forum=$1 AND created >= $2 ORDER BY created ASC;`

	selAsc = `SELECT id,author,created,message,COALESCE(slug,''),title,votes
			FROM threads
			WHERE forum=$1 ORDER BY created ASC;`

	selectThreadId = `SELECT author,created,forum,message,COALESCE(slug,''),title,votes
			FROM threads
			WHERE id=$1`

	selectThreadSlug = `SELECT id,author,created,forum,message,COALESCE(slug,''),title,votes
			FROM threads
			WHERE slug=$1`
)

func slid(ctx *fasthttp.RequestCtx) (string, int) {
	slug := ctx.UserValue("slug_or_id").(string)
	id, _ := strconv.ParseInt(slug, 10, 32)

	return slug, int(id)
}

func CreateThread(ctx *fasthttp.RequestCtx) {
	t := &models.Thread{}
	t.UnmarshalJSON(ctx.PostBody())
	slug := ctx.UserValue("slug").(string)

	if t.Forum == "" {
		t.Forum = slug
	}

	var err error
	if t.Slug != "" {
		err = db.QueryRow(`INSERT INTO threads(author,created,forum,message,title,slug)
							VALUES((SELECT nickname
									FROM users
									WHERE nickname=$1),
									$2::timestamptz,
									(SELECT slug
									FROM forums
									WHERE slug=$3),
									$4,$5,$6)
							RETURNING id,author,created,forum,message,title,slug;`, t.Author, t.Created, t.Forum, t.Message, t.Title, t.Slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Title, &t.Slug)
	} else {
		err = db.QueryRow(`INSERT INTO threads(author,created,forum,message,title)
							VALUES((SELECT nickname
									FROM users
									WHERE nickname=$1),
									$2::timestamptz,
									(SELECT slug
									FROM forums
									WHERE slug=$3),
									$4,$5)
							RETURNING id,author,created,forum,message,title;`, t.Author, t.Created, t.Forum, t.Message, t.Title).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Title)
	}

	if err != nil {
		errStr := err.Error()
		if errStr == "ERROR: duplicate key value violates unique constraint \"idx_threads_slug\" (SQLSTATE 23505)" {
			err = db.QueryRow(`SELECT id,author,created,forum,message,slug,title,votes
			FROM threads
			WHERE slug=$1;`, t.Slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)

			p, _ := t.MarshalJSON()
			ut.Respond(ctx, fasthttp.StatusConflict, p)

			return
		}
		if errStr == "ERROR: null value in column \"author\" violates not-null constraint (SQLSTATE 23502)" {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}
		if errStr == "ERROR: null value in column \"forum\" violates not-null constraint (SQLSTATE 23502)" {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}
	}

	/* TRIGGERED-BEGIN */
	db.Exec(`UPDATE forums
				SET threads=threads+1
				WHERE slug=$1;`, t.Forum)

	db.Exec(`INSERT INTO participants(nickname,forum_slug,id)
				VALUES ($1,$2,(SELECT id FROM users WHERE nickname=$1))
				ON CONFLICT DO NOTHING;`, t.Author, t.Forum)

	/* TRIGGERED-END */

	p, _ := t.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusCreated, p)

	return
}

func Threads(ctx *fasthttp.RequestCtx) {
	slugFromURL := ctx.UserValue("slug").(string)
	slug := ""
	db.QueryRow(`SELECT slug
					FROM forums
					WHERE slug=$1;`, slugFromURL).Scan(&slug)

	if slug == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := string(ctx.QueryArgs().Peek("since"))

	var rows *pgx.Rows

	if desc {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(selDescLimitSince, slug, since, limit)
			} else if since == "" {
				rows, _ = db.Query(selDescLimit, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(selDescSince, slug, since)
			} else {
				rows, _ = db.Query(selDesc, slug)
			}
		}
	} else {
		if limit > 0 {
			if since != "" {
				rows, _ = db.Query(selAscLimitSince, slug, since, limit)
			} else {
				rows, _ = db.Query(selAscLimit, slug, limit)
			}
		} else {
			if since != "" {
				rows, _ = db.Query(selAscSince, slug, since)
			} else {
				rows, _ = db.Query(selAsc, slug)
			}
		}
	}

	threads := make(models.ThreadsArr, 0, limit)
	for rows.Next() {
		temp := models.Thread{Forum: slug}
		rows.Scan(&temp.ID, &temp.Author, &temp.Created, &temp.Message, &temp.Slug, &temp.Title, &temp.Votes)
		threads = append(threads, &temp)
	}
	rows.Close()

	p, _ := threads.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func CreatePosts(ctx *fasthttp.RequestCtx) {
	posts := models.PostsArr{}
	posts.UnmarshalJSON(ctx.PostBody())

	slug, id := slid(ctx)
	var thid int
	var forum string

	size := len(posts)

	if id != 0 {
		db.QueryRow(`SELECT id,forum
						FROM threads
						WHERE id=$1;`, id).Scan(&thid, &forum)
	} else {
		db.QueryRow(`SELECT id,forum
						FROM threads
						WHERE slug=$1;`, slug).Scan(&thid, &forum)
	}

	if thid == 0 {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	if size == 0 {
		p, _ := posts.MarshalJSON()
		ut.Respond(ctx, fasthttp.StatusCreated, p)

		return
	}

	if size != 0 {
		valueStrings := make([]string, 0, len(posts))
		valueArgs := make([]interface{}, 0, len(posts)*7)
		i := 1
		for _, post := range posts {
			if post.Parent != 0 {
				valueStrings = append(valueStrings, fmt.Sprintf(`((SELECT nextval('posts_id_seq')::integer), $%d, $%d, $%d,
					(SELECT
					(CASE WHEN EXISTS
					(SELECT 1
					FROM posts p
					WHERE p.id=$%d AND p.thread_id=$%d) THEN $%d ELSE -1 END)), array_append(
						(SELECT path FROM posts WHERE id=$%d),
						  (SELECT currval('posts_id_seq')::integer)),
						  $%d)`, i, i+1, i+2, i+3, i+4, i+5, i+5, i+6))
				valueArgs = append(valueArgs, post.Author)
				valueArgs = append(valueArgs, post.Message)
				valueArgs = append(valueArgs, thid)
				valueArgs = append(valueArgs, post.Parent)
				valueArgs = append(valueArgs, thid)
				valueArgs = append(valueArgs, post.Parent)
				valueArgs = append(valueArgs, forum)
				i += 7
			} else {
				valueStrings = append(valueStrings, fmt.Sprintf("((SELECT nextval('posts_id_seq')::integer), $%d, $%d, $%d, %s, array_append('{}', (SELECT currval('posts_id_seq')::integer)), $%d)", i, i+1, i+2, "NULL", i+3))
				valueArgs = append(valueArgs, post.Author)
				valueArgs = append(valueArgs, post.Message)
				valueArgs = append(valueArgs, thid)
				valueArgs = append(valueArgs, forum)
				i += 4
			}
		}

		var query strings.Builder
		fmt.Fprintf(&query, `INSERT INTO posts(id,author,message,thread_id,parent_id,path,forum_slug) VALUES %s`, strings.Join(valueStrings, ","))
		fmt.Fprintf(&query, ` RETURNING author,id,created,thread_id,COALESCE(parent_id,0),forum_slug,message;`)

		rows, _ := db.Query(query.String(), valueArgs...)

		postsResp := models.PostsArr{}
		for rows.Next() {
			post := models.Post{}
			rows.Scan(&post.Author, &post.ID, &post.Created, &post.Thread, &post.Parent, &post.Forum, &post.Message)
			postsResp = append(postsResp, &post)
		}

		if finalRowsErr := rows.Err(); finalRowsErr != nil {
			if pgerr, ok := finalRowsErr.(pgx.PgError); ok {
				if pgerr.ConstraintName == "posts_parent_id_fkey" {
					ut.ErrRespond(ctx, fasthttp.StatusConflict)

					return
				}
				if pgerr.ConstraintName == "posts_author_fkey" {
					ut.ErrRespond(ctx, fasthttp.StatusNotFound)

					return
				}
			}
		}

		/* TRIGGERED-BEGIN */
		//go func() {
		db.Exec(`UPDATE forums
			SET posts=posts+$1
			WHERE slug=$2;`, size, forum)

		var insertParticipants strings.Builder
		fmt.Fprintf(&insertParticipants, `INSERT INTO participants(nickname,forum_slug,id) VALUES `)
		for idx, u := range postsResp {
			if idx == size-1 {
				fmt.Fprintf(&insertParticipants, `('%s', '%s', (SELECT id FROM users WHERE nickname='%s'))`, u.Author, u.Forum, u.Author)
			} else {
				fmt.Fprintf(&insertParticipants, `('%s', '%s', (SELECT id FROM users WHERE nickname='%s')),`, u.Author, u.Forum, u.Author)
			}
		}
		fmt.Fprintf(&insertParticipants, ` ON CONFLICT DO NOTHING;`)
		db.Exec(insertParticipants.String())
		//}()

		p, _ := postsResp.MarshalJSON()
		ut.Respond(ctx, fasthttp.StatusCreated, p)

		return
	}
}

func Vote(ctx *fasthttp.RequestCtx) {
	slug, id := slid(ctx)
	vote := &models.Vote{}
	vote.UnmarshalJSON(ctx.PostBody())

	var query strings.Builder
	fmt.Fprintf(&query, `SELECT id,author,created,forum,message,COALESCE(slug,''),title,votes
						FROM threads
						WHERE`)

	t := &models.Thread{}
	var insert strings.Builder

	if id != 0 {
		fmt.Fprintf(&insert, `INSERT INTO votes(nickname, thread_id, voice)
							VALUES($1, $2, $3)
							ON CONFLICT ON CONSTRAINT votes_pkey DO
							UPDATE SET voice=$3
							WHERE votes.thread_id=$2
							AND votes.nickname=$1;`)

		_, err := db.Exec(insert.String(), vote.Nickname, id, vote.Voice)

		if err != nil {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return

		}

		fmt.Fprintf(&query, ` id=$1;`)

		err = db.QueryRow(query.String(), id).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)
	} else {
		fmt.Fprintf(&insert, `INSERT INTO votes(nickname, thread_id, voice)
							  VALUES($1,
									(SELECT id
										FROM threads
										WHERE slug=$2), $3)
							ON CONFLICT ON CONSTRAINT votes_pkey DO
							UPDATE SET voice=$3
							WHERE votes.thread_id=(SELECT id FROM threads WHERE slug=$2)
							AND votes.nickname=$1;`)
		_, err := db.Exec(insert.String(), vote.Nickname, slug, vote.Voice)

		if err != nil {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		fmt.Fprintf(&query, ` slug=$1;`)

		err = db.QueryRow(query.String(), slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)
	}

	p, _ := t.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func ThreadInfo(ctx *fasthttp.RequestCtx) {
	slug, id := slid(ctx)
	t := &models.Thread{}

	if id != 0 {
		t.ID = int32(id)
		db.QueryRow(selectThreadId, id).Scan(&t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)

	} else {
		db.QueryRow(selectThreadSlug, slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)
	}

	if t.Author == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := t.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func SortPosts(ctx *fasthttp.RequestCtx) {
	slug, idFromURL := slid(ctx)

	desc := ctx.QueryArgs().GetBool("desc")
	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := ctx.QueryArgs().GetUintOrZero("since")
	sort := string(ctx.QueryArgs().Peek("sort"))

	var id int32
	var errThr error
	if idFromURL != 0 {
		errThr = db.QueryRow(`SELECT id
						FROM threads
						WHERE id=$1;`, idFromURL).Scan(&id)
	} else {
		errThr = db.QueryRow(`SELECT id
							FROM threads
							WHERE slug=$1;`, slug).Scan(&id)
	}

	if errThr != nil {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	var rows *pgx.Rows

	switch sort {
	case "":
		fallthrough
	case "flat":
		if since != 0 {
			if desc {
				rows, _ = db.Query(selectPostsFlatLimitSinceDescByID, id,
					since, limit)
			} else {
				rows, _ = db.Query(selectPostsFlatLimitSinceByID, id,
					since, limit)
			}
		} else {
			if desc {
				rows, _ = db.Query(selectPostsFlatLimitDescByID, id, limit)
			} else {
				rows, _ = db.Query(selectPostsFlatLimitByID, id, limit)
			}
		}
	case "tree":
		if since != 0 {
			if desc {
				rows, _ = db.Query(`
							SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
							FROM posts p
							WHERE thread_id = $1 AND p.path < (SELECT path FROM posts WHERE id = $2)
							ORDER BY p.path DESC
							LIMIT $3
							`, id, since, limit)
			} else {
				rows, _ = db.Query(`
								SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
								FROM posts p
								WHERE thread_id = $1 AND p.path > (SELECT path FROM posts WHERE id = $2)
								ORDER BY p.path
								LIMIT $3
								`, id, since, limit)
			}
		} else {
			if desc {
				rows, _ = db.Query(`
								SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
								FROM posts p
								WHERE thread_id = $1
								ORDER BY p.path DESC
								LIMIT $2
								`, id, limit)
			} else {
				rows, _ = db.Query(`
									SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
									FROM posts p
									WHERE thread_id = $1
									ORDER BY p.path
									LIMIT $2
									`, id, limit)
			}
		}
	case "parent_tree":
		if since != 0 {
			if desc {
				rows, _ = db.Query(`SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
									FROM posts p
									WHERE
									path[1] IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id IS NULL
									AND p2.id < (SELECT path[1] FROM posts WHERE id=$2)
									ORDER BY p2.id DESC
									LIMIT $3)
									ORDER BY path[1] DESC, p.path`, id, since, limit)
			} else {
				rows, _ = db.Query(`SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
									FROM posts p
									WHERE
									path[1] IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id IS NULL
									AND p2.id > (SELECT path[1] FROM posts WHERE id=$2)
									ORDER BY p2.id ASC
									LIMIT $3)
									ORDER BY p.path`, id, since, limit)
			}
		} else {
			if desc {
				rows, _ = db.Query(`SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
									FROM posts p
									WHERE
									path[1] IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id IS NULL
									ORDER BY p2.id DESC
									LIMIT $2)
									ORDER BY path[1] DESC, p.path`, id, limit)
			} else {
				rows, _ = db.Query(`SELECT p.id, p.author, p.created, p.edited, p.message, COALESCE(p.parent_id,0), p.forum_slug
									FROM posts p
									WHERE
									path[1] IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id IS NULL
									ORDER BY p2.id
									LIMIT $2)
									ORDER BY p.path`, id, limit)
			}
		}
	}

	posts := make(models.PostsArr, 0, limit)
	for rows.Next() {
		temp := models.Post{Thread: id}
		rows.Scan(&temp.ID, &temp.Author, &temp.Created, &temp.IsEdited, &temp.Message, &temp.Parent, &temp.Forum)
		posts = append(posts, &temp)

	}
	rows.Close()

	p, _ := posts.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func UpdateThread(ctx *fasthttp.RequestCtx) {
	slug, id := slid(ctx)
	update := &models.ThreadUpdate{}
	t := &models.Thread{}
	update.UnmarshalJSON(ctx.PostBody())

	if update.Message == "" && update.Title == "" {
		var err error
		if id != 0 {
			err = db.QueryRow(`SELECT author,created,forum,id,message,slug,title
								FROM threads
								WHERE id=$1;`, id).Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message, &t.Slug, &t.Title)
		} else {
			err = db.QueryRow(`SELECT author,created,forum,id,message,slug,title
								FROM threads
								WHERE slug=$1;`, slug).Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message, &t.Slug, &t.Title)
		}

		if err != nil {
			ut.ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		p, _ := t.MarshalJSON()
		ut.Respond(ctx, fasthttp.StatusOK, p)

		return
	}

	valueArgs := make([]interface{}, 0, 3)
	var query strings.Builder
	query.WriteString(`UPDATE threads
						SET `)

	i := 1
	if update.Message != "" {
		if update.Title != "" {
			fmt.Fprintf(&query, `message=$%d`, i)
		} else {
			fmt.Fprintf(&query, `message=$%d`, i)
		}
		i++
		valueArgs = append(valueArgs, &update.Message)
	}

	if update.Title != "" {
		if update.Message != "" {
			fmt.Fprintf(&query, `, title=$%d`, i)
			i++
			valueArgs = append(valueArgs, &update.Title)
		} else {
			fmt.Fprintf(&query, `title=$%d`, i)
			i++
			valueArgs = append(valueArgs, &update.Title)
		}
	}

	fmt.Fprintf(&query, ` WHERE `)

	if id != 0 {
		fmt.Fprintf(&query, `id=$%d RETURNING author,created,forum,id,message,slug,title;`, i)
		valueArgs = append(valueArgs, &id)
	} else {
		fmt.Fprintf(&query, `slug=$%d RETURNING author,created,forum,id,message,slug,title;`, i)
		valueArgs = append(valueArgs, &slug)
	}

	err := db.QueryRow(query.String(), valueArgs...).Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message, &t.Slug, &t.Title)

	if err != nil {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := t.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
