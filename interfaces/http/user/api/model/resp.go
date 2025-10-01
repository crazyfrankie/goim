package model

type UserInfoResp struct {
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	UserUniqueName string `json:"user_unique_name"`
	Email          string `json:"email"`
	Description    string `json:"description"`
	Avatar         string `json:"avatar"`
	UserCreateTime int64  `json:"user_create_time"`
}
