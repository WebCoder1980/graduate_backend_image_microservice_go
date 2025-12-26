package model

import "time"

type TaskInfo struct {
	Id        int64
	CreatedDT time.Time
}

type TaskResponse struct {
	CommonStatusId int64       `json:"common_status_id"`
	Images         []ImageInfo `json:"images"`
	CreatedDT      time.Time   `json:"created_dt"`
}
