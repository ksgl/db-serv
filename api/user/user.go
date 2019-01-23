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
	PScreateUserInsert, _ = db.Prepare("createUserInsert", createUserInsert)
	PSuserByNicknameOrEmailSelect, _ = db.Prepare("userByNicknameOrEmailSelect", userByNicknameOrEmailSelect)
	PSuserByNicknameExtendedSelect, _ = db.Prepare("userByNicknameExtendedSelect", userByNicknameExtendedSelect)
	PSuserByNicknameShortSelect, _ = db.Prepare("userByNicknameShortSelect", userByNicknameShortSelect)
	PSupdateUsers, _ = db.Prepare("updateUsers", updateUsers)
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

var (
	PScreateUserInsert             *pgx.PreparedStatement
	PSuserByNicknameOrEmailSelect  *pgx.PreparedStatement
	PSuserByNicknameExtendedSelect *pgx.PreparedStatement
	PSuserByNicknameShortSelect    *pgx.PreparedStatement
	PSupdateUsers                  *pgx.PreparedStatement
)

func CreateUser(ctx *fasthttp.RequestCtx) {
	u := &models.User{}
	u.UnmarshalJSON(ctx.PostBody())
	u.Nickname = ctx.UserValue("nickname").(string)

	_, err := db.Exec(PScreateUserInsert.Name, u.About, u.Email, u.Fullname, u.Nickname)

	if err != nil {
		rows, _ := db.Query(PSuserByNicknameOrEmailSelect.Name, u.Email, u.Nickname)

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

	db.QueryRow(PSuserByNicknameExtendedSelect.Name, u.Nickname).Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)

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

	res, err := db.Exec(PSupdateUsers.Name, u.About, u.Email, u.Fullname, finalUser.Nickname)

	if err != nil {
		ut.ErrRespond(ctx, fasthttp.StatusConflict)

		return

	}

	if res.RowsAffected() == 0 {
		ut.ErrRespond(ctx, fasthttp.StatusNotFound)

		return
	}

	db.QueryRow(PSuserByNicknameShortSelect.Name, finalUser.Nickname).Scan(&finalUser.About, &finalUser.Email, &finalUser.Fullname)

	p, _ := finalUser.MarshalJSON()
	ut.Respond(ctx, fasthttp.StatusOK, p)

	return
}
