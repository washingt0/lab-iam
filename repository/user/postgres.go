package user

import (
	"lab/iam/database/types"
	"lab/iam/domain/user"

	"github.com/washingt0/oops"
)

type repository struct {
	tx types.Transaction
}

// New initializes a instance of repository
func New(tx types.Transaction) user.IUser {
	return &repository{
		tx: tx,
	}
}

// Create a user and propate its creation
func (r *repository) Create(u *user.User) (id *string, err error) {
	if err = r.tx.QueryRow(`
		INSERT INTO t_user(name, username, password)
		VALUES ($1, $2, $3)
		RETURNING id
	`, u.Name, u.Username, u.Password).Scan(&id); err != nil {
		return nil, oops.ThrowError("Was not possible to create a user", err)
	}

	if _, err = r.tx.Exec(`
		INSERT INTO t_outgoing_message(event, queue, payload)
		SELECT 'USER_CREATED', 'USER', jsonb_build_object(
			'id', id,
			'name', name,
			'username', username,
			'active', active
		)::JSONB
		FROM t_user
		WHERE id = $1
	`, id); err != nil {
		return nil, oops.ThrowError("Was not possible to propagate changes", err)
	}

	return
}
