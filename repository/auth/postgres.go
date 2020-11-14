package auth

import (
	"lab/iam/database/types"
	"lab/iam/domain/auth"
)

type repository struct {
	tx types.Transaction
}

func New(tx types.Transaction) auth.IAuth {
	return &repository{
		tx: tx,
	}
}

func (r *repository) CheckCredentials(username, password, userAgent *string) (err error) {
	// TODO:
	return
}
