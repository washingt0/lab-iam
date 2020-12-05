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

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/washingt0/oops"
	"golang.org/x/crypto/bcrypt"
)

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
		ID:               sess.ID,
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
	)

	out.ExpiresAt = out.SessionExpiresAt.Unix()
	out.Issuer = config.GetConfig().JWT.Issuer
	out.Audience = config.GetConfig().JWT.Audience
	out.NotBefore = time.Now().Unix()
	out.Subject = *out.UserID

	if idx, err = rand.Int(rand.Reader, big.NewInt(int64(len(config.GetConfig().JWT.Keys)))); err != nil {
		return nil, oops.ThrowError("Unable to define key", err)
	}

	tk := jwt.NewWithClaims(jwt.SigningMethodRS256, out)
	tk.Header["kid"] = config.GetConfig().JWT.Keys[idx.Int64()].ID

	token = new(string)
	if *token, err = tk.
		SignedString(config.GetConfig().JWT.Keys[idx.Int64()].PrivateKey); err != nil {
		return nil, oops.ThrowError("Was not possible to generate JWT", err)
	}

	return
}
