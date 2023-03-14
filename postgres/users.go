package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const round = 12

var (
	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type UserService struct {
	DB *sql.DB
}

type User struct {
	ID        int64
	Username  string
	Email     string
	Password  password
	CreatedAt time.Time
}

type password struct {
	plainText *string
	hash      []byte
}

func (p *password) Set(pwd string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), round)
	if err != nil {
		return err
	}
	p.plainText = &pwd
	p.hash = hash
	return nil
}

func (p *password) Matches(pwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(pwd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (us UserService) Create(user *User) error {
	queryInsertUser := `
        insert into users (username, email, password)
        values ($1, $2, $3)
        returning id, created_at
    `
	argsInsertUser := []any{user.Username, user.Email, user.Password.hash}

	tx, err := us.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userRow := tx.QueryRowContext(context.Background(), queryInsertUser, argsInsertUser...)
	err = userRow.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if e.Code == "23505" {
				return ErrDuplicateEmail
			}
		default:
			return err
		}
		return err
	}

	queryInsertList := `
        insert into taskorder
        (user_id, category, value)
        values
        ($1, 'TODO', array[]::bigint[]),
        ($1, 'IN PROGRESS', array[]::bigint[]),
        ($1, 'TESTING', array[]::bigint[]),
        ($1, 'DONE', array[]::bigint[])
    `
	_, err = tx.ExecContext(context.Background(), queryInsertList, user.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (us UserService) GetByEmail(email string) (*User, error) {
	query := `
		select id, username, email, password, created_at
		from users
		where email = $1
	`
	args := []any{email}
	row := us.DB.QueryRowContext(context.Background(), query, args...)
	user := User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u UserService) GetForToken(token string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(token))
	query := `
        select users.id, users.username, users.email, users.created_at
        from users
        inner join tokens
        on tokens.user_id = users.id
        where tokens.hash = $1
        and expiry > $2
    `
	args := []any{tokenHash[:], time.Now()}
	row := u.DB.QueryRowContext(context.Background(), query, args...)
	user := User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
