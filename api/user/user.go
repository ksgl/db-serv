package user

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
	createUserInsert = `INSERT INTO users(about,email,fullname,nickname)
						VALUES($1,$2,$3,$4);`

	userByNicknameOrEmailSelect = `SELECT about,email,fullname,nickname
									FROM users
									WHERE email=$1 OR nickname=$2;`

	userByNicknameExtendedSelect = `SELECT about,email,fullname,nickname
								FROM users
								WHERE nickname=$1;`

	userByNicknameShortSelect = `SELECT about,email,fullname
								FROM users
								WHERE nickname=$1;`

	updateUsers = `UPDATE users
					SET about=COALESCE($1,about),email=COALESCE($2,email),fullname=COALESCE($3,fullname)
					WHERE nickname=$4;`
)

func CreateUser(ctx *fasthttp.RequestCtx) {
	u := &models.User{}
	u.UnmarshalJSON(ctx.PostBody())
	u.Nickname = ctx.UserValue("nickname").(string)

	_, err := db.Exec(createUserInsert, u.About, u.Email, u.Fullname, u.Nickname)

	if err != nil {
		rows, _ := db.Query(userByNicknameOrEmailSelect, u.Email, u.Nickname)

		var users models.UsersArr
		for rows.Next() {
			temp := models.User{}
			rows.Scan(&temp.About, &temp.Email, &temp.Fullname, &temp.Nickname)
			users = append(users, &temp)
		}

		rows.Close()
		p, _ := users.MarshalJSON()
		ut.Respond(ctx, fasthttp.StatusConflict, p)

		return

	}

	p, _ := u.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusCreated, p)

	return
}

func InfoUser(ctx *fasthttp.RequestCtx) {
	u := &models.User{}
	u.Nickname = ctx.UserValue("nickname").(string)

	db.QueryRow(userByNicknameExtendedSelect, u.Nickname).Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)

	if u.Email == "" {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	p, _ := u.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}

func UpdateUser(ctx *fasthttp.RequestCtx) {
	u := &models.UserUpdate{}
	u.UnmarshalJSON(ctx.PostBody())
	finalUser := &models.User{}
	finalUser.Nickname = ctx.UserValue("nickname").(string)

	res, err := db.Exec(updateUsers, u.About, u.Email, u.Fullname, finalUser.Nickname)

	if err != nil {
		ut.ErrRespond(ctx, fasthttp.StatusConflict)

		return

	}

	if res.RowsAffected() == 0 {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	db.QueryRow(userByNicknameShortSelect, finalUser.Nickname).Scan(&finalUser.About, &finalUser.Email, &finalUser.Fullname)

	p, _ := finalUser.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
