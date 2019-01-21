package api

import (
	"fmt"
	"forum/models"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

// const sqlGetPostsFlat = `
// 	SELECT p.author, p.created, p.forum, p.isedited, p.message, p.parent, p.thread, p.id
// 	FROM posts p
// 	WHERE thread = $1
// `

const selectPostsFlatLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY p.id
	LIMIT $2
`

const selectPostsFlatLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY p.id DESC
	LIMIT $2
`

const selectPostsFlatLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.id > $2
	ORDER BY p.id
	LIMIT $3
`
const selectPostsFlatLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.id < $2
	ORDER BY p.id DESC
	LIMIT $3
`

const selectPostsTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY p.path
	LIMIT $2
`

const selectPostsTreeLimitDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1
	ORDER BY path DESC
	LIMIT $2
`

const selectPostsTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and (p.path > (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path
	LIMIT $3
`

const selectPostsTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and (p.path < (SELECT p2.path from posts p2 where p2.id = $2))
	ORDER BY p.path DESC
	LIMIT $3
`

const selectPostsParentTreeLimitByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id = 0
		ORDER BY p2.path
		LIMIT $3
	)
	ORDER BY path
`

const selectPostsParentTreeLimitDescByID = `
SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
FROM posts p
WHERE p.thread_id = $1 and p.path[1] IN (
    SELECT p2.path[1]
    FROM posts p2
	WHERE p2.parent_id = 0 and p2.thread_id = $2
	ORDER BY p2.path DESC
    LIMIT $3
)
ORDER BY p.path[1] DESC, p.path
`

const selectPostsParentTreeLimitSinceByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id = 0 and p2.path[1] > (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path
		LIMIT $4
	)
	ORDER BY p.path
`

const selectPostsParentTreeLimitSinceDescByID = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE p.thread_id = $1 and p.path[1] IN (
		SELECT p2.path[1]
		FROM posts p2
		WHERE p2.thread_id = $2 AND p2.parent_id = 0 and p2.path[1] < (SELECT p3.path[1] from posts p3 where p3.id = $3)
		ORDER BY p2.path DESC
		LIMIT $4
	)
	ORDER BY p.path[1] DESC, p.path
`

const sqlGetPostsFlat = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE thread_id = $1
	`

const sqlGetPostsParentTree = `
	SELECT p.id, p.author, p.created, p.edited, p.message, p.parent_id, p.forum_slug
	FROM posts p
	WHERE root IN (SELECT id FROM posts p2 WHERE p2.thread_id=$1 AND p2.parent_id=0
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

	// log.Println(err)
	if err != nil {
		errStr := err.Error()
		if errStr == "ERROR: duplicate key value violates unique constraint \"idx_threads_slug\" (SQLSTATE 23505)" {
			err = db.QueryRow(`SELECT id,author,created,forum,message,slug,title,votes
			FROM threads
			WHERE slug=$1;`, t.Slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)

			p, _ := t.MarshalJSON()
			Respond(ctx, fasthttp.StatusConflict, p)

			return
		}
		if errStr == "ERROR: null value in column \"author\" violates not-null constraint (SQLSTATE 23502)" {
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}
		if errStr == "ERROR: null value in column \"forum\" violates not-null constraint (SQLSTATE 23502)" {
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}
	}

	/* TRIGGERED-BEGIN */
	db.Exec(`UPDATE forums
				SET threads=threads+1
				WHERE slug=$1;`, t.Forum)

	db.Exec(`INSERT INTO participants(nickname,forum_slug)
				VALUES ($1,$2)
				ON CONFLICT DO NOTHING;`, t.Author, t.Forum)
	/* TRIGGERED-END */

	p, _ := t.MarshalJSON()
	Respond(ctx, fasthttp.StatusCreated, p)

	return
}

func Threads(ctx *fasthttp.RequestCtx) {
	slugFromURL := ctx.UserValue("slug").(string)
	slug := ""
	db.QueryRow(`SELECT slug
					FROM forums
					WHERE slug=$1;`, slugFromURL).Scan(&slug)

	// if err != nil {
	// 	ErrRespond(ctx, fasthttp.StatusNotFound)

	// 	return
	// }

	if slug == "" {
		ErrRespond(ctx, fasthttp.StatusNotFound)

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
	Respond(ctx, fasthttp.StatusOK, p)

	return
}

// func CreatePosts(ctx *fasthttp.RequestCtx) {
// 	posts := models.PostsArr{}
// 	posts.UnmarshalJSON(ctx.PostBody())

// 	slug, id := slid(ctx)
// 	var thid int
// 	var forum string

// 	size := len(posts)

// 	if id != 0 {
// 		db.QueryRow(`SELECT id,forum
// 						FROM threads
// 						WHERE id=$1;`, id).Scan(&thid, &forum)
// 	} else {
// 		db.QueryRow(`SELECT id,forum
// 						FROM threads
// 						WHERE slug=$1;`, slug).Scan(&thid, &forum)
// 	}

// 	// if err != nil {
// 	// 	ErrRespond(ctx, fasthttp.StatusNotFound)

// 	// 	return
// 	// }

// 	if thid == 0 {
// 		ErrRespond(ctx, fasthttp.StatusNotFound)

// 		return
// 	}

// 	if size == 0 {
// 		p, _ := posts.MarshalJSON()
// 		Respond(ctx, fasthttp.StatusCreated, p)

// 		return
// 	}

// 	if size != 0 {
// 		valueStrings := make([]string, 0, len(posts))
// 		valueArgs := make([]interface{}, 0, len(posts)*7)
// 		i := 1
// 		for _, post := range posts {
// 			if post.Parent != 0 {
// 				valueStrings = append(valueStrings, fmt.Sprintf(`($%d, $%d, $%d,
// 					(SELECT
// 					(CASE WHEN EXISTS
// 					(SELECT 1
// 					FROM posts p
// 					WHERE p.id=$%d AND p.thread_id=$%d) THEN $%d ELSE -1 END)), $%d)`, i, i+1, i+2, i+3, i+4, i+5, i+6))
// 				valueArgs = append(valueArgs, post.Author)
// 				valueArgs = append(valueArgs, post.Message)
// 				valueArgs = append(valueArgs, thid)
// 				valueArgs = append(valueArgs, post.Parent)
// 				valueArgs = append(valueArgs, thid)
// 				valueArgs = append(valueArgs, post.Parent)
// 				valueArgs = append(valueArgs, forum)
// 				i += 7
// 			} else {
// 				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, %s, $%d)", i, i+1, i+2, "0", i+3))
// 				valueArgs = append(valueArgs, post.Author)
// 				valueArgs = append(valueArgs, post.Message)
// 				valueArgs = append(valueArgs, thid)
// 				valueArgs = append(valueArgs, forum)
// 				i += 4
// 			}
// 		}

// 		var query strings.Builder
// 		fmt.Fprintf(&query, `INSERT INTO posts(author,message,thread_id,parent_id,forum_slug) VALUES %s`, strings.Join(valueStrings, ","))
// 		fmt.Fprintf(&query, ` RETURNING author,id,created,thread_id,COALESCE(parent_id,0),forum_slug,message;`)

// 		rows, _ := db.Query(query.String(), valueArgs...)

// 		postsResp := models.PostsArr{}
// 		for rows.Next() {
// 			post := models.Post{}
// 			rows.Scan(&post.Author, &post.ID, &post.Created, &post.Thread, &post.Parent, &post.Forum, &post.Message)
// 			postsResp = append(postsResp, &post)
// 		}
// 		err := rows.Err()

// 		// if finalRowsErr := rows.Err(); finalRowsErr != nil {
// 		// 	if pgerr, ok := finalRowsErr.(pgx.PgError); ok {
// 		// 		if pgerr.ConstraintName == "posts_parent_id_fkey" {
// 		// 			ErrRespond(ctx, fasthttp.StatusConflict)

// 		// 			return
// 		// 		}
// 		// 		if pgerr.ConstraintName == "posts_author_fkey" {
// 		// 			ErrRespond(ctx, fasthttp.StatusNotFound)

// 		// 			return
// 		// 		}
// 		// 	}
// 		// }

// 		// log.Println(err)
// 		if err != nil {
// 			if err.Error() == "ERROR: null value in column \"parent_id\" violates not-null constraint (SQLSTATE 23502)" {
// 				ErrRespond(ctx, fasthttp.StatusConflict)

// 				return
// 			}
// 			if err.Error() == "ERROR: insert or update on table \"posts\" violates foreign key constraint \"posts_author_fkey\" (SQLSTATE 23503)" {
// 				ErrRespond(ctx, fasthttp.StatusNotFound)

// 				return
// 			}
// 		}

// 		/* TRIGGERED-BEGIN */
// 		//go func() {
// 		db.Exec(`UPDATE forums
// 			SET posts=posts+$1
// 			WHERE slug=$2;`, size, forum)

// 		var insertParticipants strings.Builder
// 		fmt.Fprintf(&insertParticipants, `INSERT INTO participants(nickname,forum_slug) VALUES `)
// 		for idx, u := range postsResp {
// 			if idx == size-1 {
// 				fmt.Fprintf(&insertParticipants, `('%s', '%s')`, u.Author, forum)
// 			} else {
// 				fmt.Fprintf(&insertParticipants, `('%s', '%s'),`, u.Author, forum)
// 			}
// 		}
// 		fmt.Fprintf(&insertParticipants, ` ON CONFLICT DO NOTHING;`)

// 		db.Exec(insertParticipants.String())
// 		//}()

// 		p, _ := postsResp.MarshalJSON()
// 		Respond(ctx, fasthttp.StatusCreated, p)

// 		return
// 	}
// }

func CreatePosts(ctx *fasthttp.RequestCtx) {
	//log.Println("ya tut")
	// not found
	slug, id := slid(ctx)
	var obtainedID int
	var forum string

	// tx, _ := db.Begin()
	if id != 0 {
		db.QueryRow(`SELECT id,forum
						FROM threads
						WHERE id=$1;`, id).Scan(&obtainedID, &forum)
	} else {
		db.QueryRow(`SELECT id,forum
						FROM threads
						WHERE slug=$1;`, slug).Scan(&obtainedID, &forum)
	}

	if obtainedID == 0 {
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	posts := models.PostsArr{}
	posts.UnmarshalJSON(ctx.PostBody())

	size := len(posts)

	if size == 0 {
		p, _ := posts.MarshalJSON()
		Respond(ctx, fasthttp.StatusCreated, p)

		return
	}

	parentsSet := map[int32]bool{}
	authorsSet := map[string]bool{}
	for _, p := range posts {
		if p.Parent != 0 {
			parentsSet[p.Parent] = true
		}
		authorsSet[p.Author] = true
	}

	parents := map[int32]*models.Post{}
	//log.Println("aa")
	for pid := range parentsSet {
		tmp := &models.Post{}
		err := db.QueryRow(`SELECT author,created,edited,message,parent_id,forum_slug,thread_id
							FROM posts
							WHERE id=$1`, pid).Scan(&tmp.Author, &tmp.Created, &tmp.IsEdited, &tmp.Message, &tmp.Parent, &tmp.Forum, &tmp.Thread)

		if err != nil {
			ErrRespond(ctx, fasthttp.StatusConflict)

			return
		}
		if tmp.Thread != int32(obtainedID) {
			ErrRespond(ctx, fasthttp.StatusConflict)

			return
		}
		parents[pid] = tmp
	}

	uids, ok := UsersCheck(authorsSet)
	if !ok {
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	for _, p := range posts {
		p.Thread = int32(obtainedID)
		p.Forum = forum
		if r := parents[p.Parent]; r != nil {
			p.Path = r.Path
		} else {
			p.Path = []int32{}
		}
	}
	created := time.Now().Format("2006-01-02T15:04:05.999999999Z07:00")
	b, a := postQueryBuilder(posts, obtainedID, created)

	tx, _ := db.Begin()
	defer tx.Rollback()

	rows, err := tx.Query(b.String(), a...)
	//log.Println(err)
	if err != nil {
		ErrRespond(ctx, fasthttp.StatusConflict)

		return
	}
	defer rows.Close()

	for _, p := range posts {
		rows.Next()
		if err := rows.Scan(&p.Created, &p.ID); err != nil {
			ErrRespond(ctx, fasthttp.StatusConflict)

			return
		}
	}
	rows.Close()

	if _, err := tx.Exec(`UPDATE forums
						SET posts=posts+$1
						WHERE slug=$2;`, size, forum); err != nil {

		ErrRespond(ctx, fasthttp.StatusConflict)

		return
	}

	var insertParticipants strings.Builder
	fmt.Fprintf(&insertParticipants, `INSERT INTO participants(nickname,forum_slug) VALUES `)
	for nickname := range uids {
		// if idx == size-1 {
		// 	fmt.Fprintf(&insertParticipants, `('%s', '%s')`, nickname, forum)
		// } else {
		fmt.Fprintf(&insertParticipants, `('%s', '%s'),`, nickname, forum)
		// }
	}
	fmt.Fprintf(&insertParticipants, `('', '')`)
	fmt.Fprintf(&insertParticipants, ` ON CONFLICT DO NOTHING;`)

	tx.Exec(insertParticipants.String())

	tx.Commit()

	p, _ := posts.MarshalJSON()
	Respond(ctx, fasthttp.StatusCreated, p)

	return
}

func postQueryBuilder(ps models.PostsArr, tid int, created string) (*strings.Builder, []interface{}) {
	b := &strings.Builder{}
	a := []interface{}{}

	b.WriteString("INSERT INTO posts(author, thread_id, message, parent_id, edited, created, path, forum_slug) VALUES ")
	for i, p := range ps {
		if i != 0 {
			b.WriteString(", ")
		}

		c := 7 * i
		b.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, false, $%d, $%d, $%d)",
			c+1, c+2, c+3, c+4, c+5, c+6, c+7))
		a = append(a, p.Author, tid, p.Message, p.Parent, created, pq.Array(p.Path), p.Forum)
	}
	b.WriteString(" RETURNING created, id")

	return b, a
}

func UsersCheck(nicks map[string]bool) (map[string]string, bool) {
	if len(nicks) == 0 {
		return nil, true
	}

	arr := make([]string, 0, len(nicks))
	for n := range nicks {
		arr = append(arr, n)
	}
	rows, err := db.Query(`
		SELECT email
		FROM users
		WHERE nickname = ANY (ARRAY['` + strings.Join(arr, "', '") + `'])
	`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	r := map[string]string{}
	for _, n := range arr {
		if !rows.Next() {
			return nil, false
		}
		var t string
		err = rows.Scan(&t)

		if err != nil {
			panic(err)
		}
		r[n] = t
	}
	return r, true
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
			ErrRespond(ctx, fasthttp.StatusNotFound)

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
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		fmt.Fprintf(&query, ` slug=$1;`)

		err = db.QueryRow(query.String(), slug).Scan(&t.ID, &t.Author, &t.Created, &t.Forum, &t.Message, &t.Slug, &t.Title, &t.Votes)
	}

	p, _ := t.MarshalJSON()
	Respond(ctx, fasthttp.StatusOK, p)

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
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := t.MarshalJSON()
	Respond(ctx, fasthttp.StatusOK, p)

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
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	var rows *pgx.Rows
	//var err error

	switch sort {
	case "":
		fallthrough
	case "flat":
		if since != 0 {
			if desc {
				// log.Println(selectPostsFlatLimitSinceDescByID)
				// log.Println(id)
				// log.Println(since)
				// log.Println(limit)
				rows, _ = db.Query(selectPostsFlatLimitSinceDescByID, id,
					since, limit)
			} else {
				// log.Println(selectPostsFlatLimitSinceByID)
				// log.Println(id)
				// log.Println(since)
				// log.Println(limit)
				rows, _ = db.Query(selectPostsFlatLimitSinceByID, id,
					since, limit)
			}
		} else {
			if desc {
				// log.Println(selectPostsFlatLimitDescByID)
				// log.Println(id)
				// log.Println(since)
				// log.Println(limit)
				rows, _ = db.Query(selectPostsFlatLimitDescByID, id, limit)
			} else {
				// log.Println(selectPostsFlatLimitByID)
				// log.Println(id)
				// log.Println(since)
				// log.Println(limit)
				rows, _ = db.Query(selectPostsFlatLimitByID, id, limit)
			}
		}
	case "tree":
		var query strings.Builder
		fmt.Fprint(&query, sqlGetPostsFlat)
		if since != 0 {
			if desc {
				fmt.Fprint(&query, " AND p.path < (SELECT path FROM posts WHERE id = $2)")
			} else {
				fmt.Fprint(&query, " AND p.path > (SELECT path FROM posts WHERE id = $2)")
			}
		} else {
			fmt.Fprint(&query, " AND $2 = 0")
		}
		if desc {
			fmt.Fprint(&query, " ORDER BY p.path DESC")
		} else {
			fmt.Fprint(&query, " ORDER BY p.path")
		}
		fmt.Fprint(&query, " LIMIT $3")

		// log.Println(query.String())
		// log.Println(id)
		// log.Println(since)
		// log.Println(limit)
		rows, _ = db.Query(query.String(), id, since, limit)
	case "parent_tree":
		// if since != 0 {
		// 	if desc {
		// 		rows, _ = db.Query(selectPostsParentTreeLimitSinceDescByID, id, id,
		// 			since, limit)
		// 	} else {
		// 		rows, _ = db.Query(selectPostsParentTreeLimitSinceByID, id, id,
		// 			since, limit)
		// 	}
		// } else {
		// 	if desc {
		// 		rows, _ = db.Query(selectPostsParentTreeLimitDescByID, id, id,
		// 			limit)
		// 	} else {
		// 		rows, _ = db.Query(selectPostsParentTreeLimitByID, id, id,
		// 			limit)
		// 	}
		// }

		// var descFlag string
		// var sortAdd string
		// var sinceAdd string
		// var limitAdd string
		// if limit != 0 {
		// 	limitAdd = " WHERE rank <= " + strconv.Itoa(limit)
		// }

		// if desc {
		// 	descFlag = " desc "
		// 	sortAdd = "ORDER BY path[1] DESC, path "
		// 	if since != 0 {
		// 		sinceAdd = " AND path[1] < (SELECT path[1] FROM posts WHERE id = " + strconv.Itoa(since) + " ) "
		// 	}
		// } else {
		// 	descFlag = " ASC "
		// 	sortAdd = " ORDER BY path[1], path ASC"
		// 	if since != 0 {
		// 		sinceAdd = " AND path[1] > (SELECT path[1] FROM posts WHERE id = " + strconv.Itoa(since) + " ) "
		// 	}
		// }

		// q = "SELECT id, author, created, edited, message, parent_id, forum_slug FROM (" +
		// 	" SELECT id, author, created, edited, message, parent_id, forum_slug, " +
		// 	" dense_rank() over (ORDER BY path[1] " + descFlag + " ) AS rank " +
		// 	" FROM posts WHERE thread_id=$1 " + sinceAdd + " ) AS tree " + limitAdd + " " + sortAdd

		// rows, err = db.Query(q, id)
		var query strings.Builder
		fmt.Fprint(&query, sqlGetPostsParentTree)
		if since != 0 {
			if desc {
				fmt.Fprint(&query, " AND p2.id < (SELECT root FROM posts WHERE id=$2)")
			} else {
				fmt.Fprint(&query, " AND p2.id > (SELECT root FROM posts WHERE id=$2)")
			}
		} else {
			fmt.Fprint(&query, " AND $2 = 0")
		}
		if desc {
			fmt.Fprint(&query, " ORDER BY p2.id DESC")
		} else {
			fmt.Fprint(&query, " ORDER BY p2.id")
		}
		fmt.Fprint(&query, " LIMIT $3)")
		if desc {
			fmt.Fprint(&query, " ORDER BY root DESC, p.path")
		} else {
			fmt.Fprint(&query, " ORDER BY p.path")
		}

		rows, _ = db.Query(query.String(), id, since, limit)
	}
	// log.Println(err)
	// log.Println(q)
	posts := make(models.PostsArr, 0, limit)
	for rows.Next() {
		temp := models.Post{Thread: id}
		rows.Scan(&temp.ID, &temp.Author, &temp.Created, &temp.IsEdited, &temp.Message, &temp.Parent, &temp.Forum)
		posts = append(posts, &temp)

	}
	// log.Println(err)
	rows.Close()

	p, _ := posts.MarshalJSON()
	Respond(ctx, fasthttp.StatusOK, p)

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
			ErrRespond(ctx, fasthttp.StatusNotFound)

			return
		}

		p, _ := t.MarshalJSON()
		Respond(ctx, fasthttp.StatusOK, p)

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
		ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := t.MarshalJSON()
	Respond(ctx, fasthttp.StatusOK, p)

	return
}
