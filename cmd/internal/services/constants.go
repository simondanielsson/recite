package services

import "path"

const maxArticleContentCharLength int = 3500

var BaseOutputPath string = path.Join("data", "output", "generations", "audio")

type Status string

const (
	InProgress Status = "in progress"
	Completed  Status = "completed"
	Failed     Status = "failed"
)
