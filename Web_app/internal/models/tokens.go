package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

type Token struct {
	PlainText string `json:"token"`
	UserID    int64  `json:"-"`
	Hash      []byte `json:"-"`
	Expiry    int64  `json:"expiry"`
	Scope     string `json:"scope"`
}

func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl).Unix(),
		Scope:  scope,
	}

	randBytes := make([]byte, 16)  // Generate 16 random bytes
	_, err := rand.Read(randBytes) // Fill the byte slice with random data
	if err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes) // Encode the random bytes to a base32 string without padding and padding for URL safety
	hash := sha256.Sum256([]byte(token.PlainText))                                               // Hash the plaintext token using SHA-256
	token.Hash = hash[:]                                                                         // Store the hash in the token struct

	return token, nil
}

// InsertToken
func (m *DBModel) InsertToken(token *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO tokens (user_id, name, email, token_hash, created_at ,updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := m.DB.ExecContext(ctx, stmt,
		token.UserID,
		u.LastName,
		u.Email,
		token.Hash,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}
