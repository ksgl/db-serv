package models

// easyjson:json
type User struct {
	About    string `json:"about"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
}

// easyjson:json
type UserUpdate struct {
	About    interface{} `json:"about"`
	Email    interface{} `json:"email"`
	Fullname interface{} `json:"fullname"`
}

// easyjson:json
type UsersArr []*User
