package models

import "time"

// easyjson:json
type Thread struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Forum   string    `json:"forum"`
	ID      int32     `json:"id"`
	Message string    `json:"message"`
	Slug    string    `json:"slug"`
	Title   string    `json:"title"`
	Votes   int32     `json:"votes"`
}

// easyjson:json
type ThreadTrunc struct {
	Author  string `json:"author"`
	Created string `json:"created"`
	Forum   string `json:"forum"`
	ID      int32  `json:"id"`
	Message string `json:"message"`
	Title   string `json:"title"`
}

// easyjson:json
type ThreadUpdate struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

// easyjson:json
type ThreadsArr []*Thread
