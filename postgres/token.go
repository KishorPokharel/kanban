package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
}

type TokenService struct {
	DB *sql.DB
}

func generateToken(userID int64, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
	}
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	h := sha256.Sum256([]byte(token.PlainText))
	token.Hash = h[:]
	return token, nil
}

func (t TokenService) New(userID int64, ttl time.Duration) (*Token, error) {
	token, err := generateToken(userID, ttl)
	if err != nil {
		return nil, err
	}
	err = t.Insert(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t TokenService) Insert(token *Token) error {
	query := `
        insert into tokens (hash, user_id, expiry)
        values ($1, $2, $3)
    `
	args := []any{token.Hash, token.UserID, token.Expiry}
	_, err := t.DB.ExecContext(context.Background(), query, args...)
	return err
}
