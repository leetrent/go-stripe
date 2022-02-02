package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"log"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token is the type for authenticication tokens
type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken generates and returns a token that last for tll (passed-in)
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	/////////////////////////////////////////////////////////////////////////
	// DELETE existing token for User (if any)
	/////////////////////////////////////////////////////////////////////////
	stmt := "DELETE FROM tokens WHERE user_id = ?"
	result, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	lastInsertId, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	logSnippet := "[api][tokens][DeleteToken] =>"
	fmt.Sprintf("%s (result.LastInsertId(): %d", logSnippet, lastInsertId)
	fmt.Sprintf("%s (result.RowsAffected(): %d", logSnippet, rowsAffected)

	/////////////////////////////////////////////////////////////////////////
	// INSERT new token for User (if any)
	/////////////////////////////////////////////////////////////////////////

	stmt = `INSERT INTO tokens
				(user_id, name, email, token_hash, expiry, created_at, updated_at)
				values(?, ?, ?, ?, ?, ?, ?)`

	result, err = m.DB.ExecContext(ctx, stmt, u.ID, u.LastName, u.Email, t.Hash, t.Expiry, time.Now(), time.Now())
	if err != nil {
		fmt.Println(err)
		return err
	}

	lastInsertId, _ = result.LastInsertId()
	rowsAffected, _ = result.RowsAffected()

	logSnippet = "[api][tokens][InsertToken] =>"
	fmt.Sprintf("%s (result.LastInsertId(): %d", logSnippet, lastInsertId)
	fmt.Sprintf("%s (result.RowsAffected(): %d", logSnippet, rowsAffected)

	return nil
}

func (m *DBModel) GetUserForToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	query :=
		`SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.email
		FROM
			users u
		INNER JOIN tokens t
			ON u.id = t.user_id
		WHERE
			t.token_hash = ?
		AND
			t.expiry > ?`

	err := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user, nil
}
