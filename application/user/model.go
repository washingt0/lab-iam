package user

// User represents a user for req/res
type User struct {
	Name     *string `json:"name" binding:"required,gt=4,lt=100"`
	Username *string `json:"username" binding:"required,gt=4,lt=64"`
	Password *string `json:"password" binding:"required,gte=8,lt=128"`
}
