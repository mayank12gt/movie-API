package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token" validate:"required,min=26"`
	Hash      []byte    `json:"-"`
	UserId    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userId int64, ttl time.Duration, scope string) (*Token, error) {

	token := Token{
		UserId: userId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return &token, nil
}

type TokenModel struct {
	DB *sql.DB
}

func (T *TokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {

	token, err := generateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = T.Insert(token)

	return token, err
}

func (T *TokenModel) Insert(token *Token) error {

	query := `INSERT INTO tokens(hash,user_id,expiry,scope) VALUES ($1,$2,$3,$4)`

	args := []interface{}{token.Hash, token.UserId, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := T.DB.ExecContext(ctx, query, args...)

	return err
}

func (T *TokenModel) DeleteAllForUser(userID int64, scope string) error {

	query := `DELETE FROM tokens WHERE scope=$1 AND user_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := T.DB.ExecContext(ctx, query, scope, userID)

	return err
}
