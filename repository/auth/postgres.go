package auth

import (
	"lab/iam/database/types"
	"lab/iam/domain/auth"

	"github.com/jackc/pgx/v4"
	"github.com/washingt0/oops"
)

type repository struct {
	tx types.Transaction
}

func New(tx types.Transaction) auth.IAuth {
	return &repository{
		tx: tx,
	}
}

func (r *repository) GetCredentials(username *string) (password string, err error) {
	if err = r.tx.QueryRow(`
		SELECT TU.password
		FROM public.t_user TU
		WHERE TU.username = $1::TEXT
		  AND TU.active = TRUE
		  AND TU.deleted_at IS NULL
	`, username).Scan(&password); err != nil {
		if err == pgx.ErrNoRows {
			err = oops.ThrowError("Invalid username or password", err)
			return
		}

		err = oops.ThrowError("Was not possible to check user credentials", err)
		return
	}

	return
}

func (r *repository) CreateSession(username, userAgent, loginIP, loginLocation *string) (out *auth.Session, err error) {
	out = new(auth.Session)

	if err = r.tx.QueryRow(`
		INSERT INTO public.t_session (user_id, user_agent, login_ip, login_location)
		SELECT TU.id, $2::TEXT, $3::INET, $4::TEXT
		FROM public.t_user TU
		WHERE TU.username = $1::TEXT
		  AND TU.active = TRUE
		  AND TU.deleted_at IS NULL
		RETURNING id, created_at, expires_at, user_id
	`, username, userAgent, loginIP, loginLocation).Scan(&out.ID, &out.CreatedAt, &out.ExpiresAt, &out.UserID); err != nil {
		return nil, oops.ThrowError("Was not possible to create session", err)
	}

	if err = r.tx.QueryRow(`
		INSERT INTO public.t_outgoing_message(queue, payload)
		VALUES ('SESSION_CREATED', jsonb_build_object(
			'session_id', $1::UUID,
			'created_at', $2::TIMESTAMP,
			'expires_at', $3::TIMESTAMP,
			'user_id', $4::UUID,
			'name', (SELECT name FROM public.t_user WHERE id = $4::UUID),
			'username', (SELECT username FROM public.t_user WHERE id = $4::UUID)
		))
		RETURNING payload->>'name', payload->>'username'
	`, out.ID, out.CreatedAt, out.ExpiresAt, out.UserID).Scan(&out.Name, &out.Username); err != nil {
		return nil, oops.ThrowError("Was not possible to propagate changes", err)
	}

	return
}

func (r *repository) DropSession(sessionID *string) (err error) {
	if _, err = r.tx.Exec(`
		UPDATE public.t_session SET deleted_at = NOW()
		WHERE id = $1::UUID
		  AND deleted_at IS NULL
	`, sessionID); err != nil {
		return oops.ThrowError("Was not possible to perform logout", err)
	}

	if _, err = r.tx.Exec(`
		INSERT INTO public.t_outgoing_message(queue, payload)
		VALUES ('SESSION_REVOKED', jsonb_build_object(
			'session_id', $1::UUID,
		))
	`, sessionID); err != nil {
		return oops.ThrowError("Was not possible to propagate changes", err)
	}

	return
}
