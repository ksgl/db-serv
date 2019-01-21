package models

// easyjson:json
type Forum struct {
	Posts   int64  `json:"posts"`
	Slug    string `json:"slug"`
	Threads int32  `json:"threads"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

// easyjson:json
type ForumTrunc struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}
