package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"user-service/internal/models"
	"user-service/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresUserRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer func() {
		_ = db.Close()
	}()

	log, _ := logger.New("info", "console")
	repo := &PostgresUserRepository{
		db:     db,
		logger: log,
	}

	ctx := context.Background()

	t.Run("Create_Success", func(t *testing.T) {
		user := &models.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashedpwd",
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now())

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.Username, user.Email, user.PasswordHash).
			WillReturnRows(rows)

		err := repo.Create(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByID_Success", func(t *testing.T) {
		id := 1
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(id, "testuser", "test@example.com", "hash", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`).
			WithArgs(id).
			WillReturnRows(rows)

		user, err := repo.GetByID(ctx, id)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, id, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		id := 99
		mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = \$1`).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(ctx, id)

		assert.NoError(t, err)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetByEmail_Success", func(t *testing.T) {
		email := "test@example.com"
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(1, "testuser", email, "hash", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(ctx, email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update_Success", func(t *testing.T) {
		user := &models.User{
			ID:           1,
			Username:     "newname",
			Email:        "new@example.com",
			PasswordHash: "newhash",
		}

		mock.ExpectExec(`UPDATE users SET username = \$1, email = \$2, password_hash = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4`).
			WithArgs(user.Username, user.Email, user.PasswordHash, user.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(ctx, user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update_NotFound", func(t *testing.T) {
		user := &models.User{ID: 99}
		mock.ExpectExec(`UPDATE users`).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, user)

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Health_Success", func(t *testing.T) {
		mock.ExpectPing()
		err := repo.Health(ctx)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
