package auth

import (
	"context"
	"crypto/rand"
	"lab/iam/config"
	"lab/iam/database"
	"lab/iam/database/types"
	"lab/iam/domain/auth"
	authRepo "lab/iam/repository/auth"
	"math/big"
	"time"

	"github.com/washingt0/oops"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// TryLogin tries to perform login with the given credentials
func TryLogin(ctx context.Context, req *Login) (token *string, err error) {
	var (
		tx             types.Transaction
		repo           auth.IAuth
		hashedPassword string
		sess           *auth.Session
	)

	if tx, err = database.NewTx(ctx, false); err != nil {
		return
	}

	repo = authRepo.New(tx)

	if hashedPassword, err = repo.GetCredentials(req.Username); err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(*req.Password)); err != nil {
		return nil, oops.ThrowError("Invalid username or password", err)
	}

	if sess, err = repo.CreateSession(req.Username, req.UserAgent, req.IPAddress, req.IPLocation); err != nil {
		return
	}

	if token, err = generateJWT(&session{
		SessionID:        sess.ID,
		CreatedAt:        sess.CreatedAt,
		SessionExpiresAt: sess.ExpiresAt,
		UserID:           sess.UserID,
		Name:             sess.Name,
		Username:         sess.Username,
	}); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return nil, oops.ThrowError("Was not possible to persist the data", err)
	}

	return
}

func generateJWT(out *session) (token *string, err error) {
	var (
		idx *big.Int
		sig jose.Signer
	)
	if idx, err = rand.Int(rand.Reader, big.NewInt(int64(len(config.GetConfig().JWT.Keys)))); err != nil {
		return nil, oops.ThrowError("Unable to define key", err)
	}

	if sig, err = jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key:       config.GetConfig().JWT.Keys[idx.Int64()].PrivateKey,
		}, (&jose.SignerOptions{
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				"kid": config.GetConfig().JWT.Keys[idx.Int64()].ID,
			},
		}).WithType("JWT")); err != nil {
		return nil, oops.ThrowError("Unable to prepare token signer", err)
	}

	out.Subject = *out.UserID
	out.Issuer = config.GetConfig().JWT.Issuer
	out.NotBefore = jwt.NewNumericDate(time.Now())
	out.IssuedAt = jwt.NewNumericDate(time.Now())
	out.Audience = jwt.Audience{config.GetConfig().JWT.Audience}
	out.Expiry = jwt.NewNumericDate(*out.SessionExpiresAt)
	out.ID = *out.SessionID

	token = new(string)
	if *token, err = jwt.Signed(sig).Claims(out).CompactSerialize(); err != nil {
		return nil, oops.ThrowError("Was not possible to generate JWT", err)
	}

	return
}
