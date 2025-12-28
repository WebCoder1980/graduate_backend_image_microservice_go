package model

import "time"

type ImageInfo struct {
	Id                    int64     `json:"-"`
	ImageProcessorImageId int64     `json:"id"`
	Filename              string    `json:"name"`
	Format                string    `json:"format"`
	TaskId                int64     `json:"task_id"`
	Position              int       `json:"position"`
	StatusId              int64     `json:"status_id"`
	EndDT                 time.Time `json:"end_dt"`
}

type ImageRequest struct {
	ImageInfo
	Width        *int     `json:"width"`
	Height       *int     `json:"height"`
	TargetFormat *string  `json:"target_format"`
	Quality      *float64 `json:"quality"`
}

type ImageStatus struct {
	ImageProcessorImageId int64
	TaskId                int64
	Position              int
	StatusId              int64
	EndDT                 time.Time
}
