package model

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

type UserRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
