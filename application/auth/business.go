package auth

import (
	"context"
	"lab/iam/database"
	"lab/iam/database/types"
	"lab/iam/domain/auth"
	authRepo "lab/iam/repository/auth"
)

func TryLogin(ctx context.Context, req *Login) (err error) {
	var (
		tx   types.Transaction
		repo auth.IAuth
	)

	if tx, err = database.NewTx(ctx, false); err != nil {
		return
	}

	repo = authRepo.New(tx)

	if err = repo.CheckCredentials(req.Username, req.Password, req.UserAgent); err != nil {
		return
	}

	return
}
