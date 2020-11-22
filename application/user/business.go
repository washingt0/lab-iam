package user

import (
	"context"
	"lab/iam/database"
	"lab/iam/database/types"
	"lab/iam/domain/user"
	userRepo "lab/iam/repository/user"
	"lab/iam/utils"

	"github.com/washingt0/oops"
	"golang.org/x/crypto/bcrypt"
)

// Create performs all business logic to create a user
func Create(ctx context.Context, u *User) (id *string, err error) {
	var (
		tx   types.Transaction
		repo user.IUser
	)

	if tx, err = database.NewTx(ctx, false); err != nil {
		return
	}

	if u.Password, err = hashPassword(u.Password); err != nil {
		return
	}

	repo = userRepo.New(tx)

	if id, err = repo.Create(&user.User{
		Name:     u.Name,
		Username: u.Username,
		Password: u.Password,
	}); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return
}

func hashPassword(plain *string) (hashed *string, err error) {
	var (
		xs []byte
	)

	if xs, err = bcrypt.GenerateFromPassword([]byte(*plain), bcrypt.DefaultCost); err != nil {
		return nil, oops.ThrowError("Was not possible to hash the password", err)
	}

	hashed = utils.GetStringPointer(string(xs))

	return
}
