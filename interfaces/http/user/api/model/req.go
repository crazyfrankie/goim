package model

type UserRegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUpdateProfileReq struct {
	Name        *string `json:"name,omitempty"`
	UniqueName  *string `json:"unique_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Sex         *int32  `json:"sex,omitempty"`
}

type UserResetPasswordReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
