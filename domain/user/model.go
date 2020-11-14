package user

// IUser defines all methods for the repository
type IUser interface {
	Create(user *User) (id *string, err error)
}

// User represents a user
type User struct {
	ID       *string
	Name     *string
	Username *string
	Password *string
}
