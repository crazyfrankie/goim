package model

type UserRegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
