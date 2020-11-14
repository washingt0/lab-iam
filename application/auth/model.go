package auth

// Login represents a login request data
type Login struct {
	Username  *string `json:"username" binding:"required"`
	Password  *string `json:"password" binding:"required"`
	UserAgent *string `json:"-"`
}
