package model

type TaskResponse struct {
	CommonStatusId int64       `json:"common_status_id"`
	Images         []ImageInfo `json:"images"`
}
