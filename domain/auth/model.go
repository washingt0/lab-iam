package auth

type IAuth interface {
	CheckCredentials(username, password, userAgent *string) error
}
