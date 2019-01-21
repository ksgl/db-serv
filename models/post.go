package models

import "time"

// easyjson:json
type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum"`
	ID       int64     `json:"id"`
	IsEdited bool      `json:"isEdited"`
	Message  string    `json:"message"`
	Parent   int32     `json:"parent"`
	Thread   int32     `json:"thread"`
	Path     []int32   `json:"-"`
}

// easyjson:json
type PostAuthor struct {
	Author User `json:"author"`
}

// easyjson:json
type PostForum struct {
	Forum Forum `json:"forum"`
}

// easyjson:json
type PostPost struct {
	Post Post `json:"post"`
}

// easyjson:json
type PostThread struct {
	Thread Thread `json:"thread"`
}

// easyjson:json
type PostUpdate struct {
	Message string `json:"message"`
}

// easyjson:json
type PostsArr []*Post
