package model

type UserRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
