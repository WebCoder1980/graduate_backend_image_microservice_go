package model

import "time"

type ImageInfo struct {
	Id       int64     `json:"id"`
	Filename string    `json:"name"`
	Format   string    `json:"format"`
	TaskId   int64     `json:"task_id"`
	Position int       `json:"position"`
	StatusId int64     `json:"status_id"`
	EndDT    time.Time `json:"end_dt"`
}

type ImageStatus struct {
	TaskId   int64
	Position int
	StatusId int64
	EndDT    time.Time
}
