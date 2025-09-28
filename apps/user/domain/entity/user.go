package entity

type User struct {
	ID int64

	Name       string // nickname
	UniqueName string // unique name
	Email      string // email
	IconURI    string // avatar URI
	IconURL    string // avatar URL

	CreatedAt int64 // creation time
	UpdatedAt int64 // update time
}
