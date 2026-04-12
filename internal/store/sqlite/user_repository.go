package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) EnsureAdmin(ctx context.Context, username string, password string) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	if strings.TrimSpace(password) == "" {
		return errors.New("数据库无用户，需设置 ADMIN_PASSWORD 以初始化管理员")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO users(username, password_hash) VALUES(?, ?)`,
		strings.TrimSpace(username),
		string(hash),
	)
	return err
}

func (r *UserRepository) CheckPassword(ctx context.Context, username string, password string) (bool, error) {
	var hash string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT password_hash FROM users WHERE username = ?`,
		strings.TrimSpace(username),
	).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return false, nil
	}

	return true, nil
}
