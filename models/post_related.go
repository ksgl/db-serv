package models

// easyjson:json
type PostRelated struct {
	AuthorRel *User   `json:"author,omitempty"`
	ForumRel  *Forum  `json:"forum,omitempty"`
	PostRel   *Post   `json:"post,omitempty"`
	ThreadRel *Thread `json:"thread,omitempty"`
}
