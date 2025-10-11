package entity

type User struct {
	UserID int64

	Name        string // nickname
	UniqueName  string // unique name
	Email       string // email
	IconURI     string // avatar URI
	IconURL     string // avatar URL
	Description string // user description
	Sex         int32  // user sex

	CreatedAt int64 // creation time
	UpdatedAt int64 // update time
}
