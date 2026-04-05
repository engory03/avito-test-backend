package dummy

import (
	"context"

	"github.com/internships-backend/test-backend-engory03/internal/shared/utils"
	usersmodel "github.com/internships-backend/test-backend-engory03/internal/users/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EnsureUsersInDB создаёт записи dummy-пользователей при отсутствии
func EnsureUsersInDB(ctx context.Context, db *gorm.DB) error {
	hash, err := utils.HashPassword("dummy-login-not-used")
	if err != nil {
		return err
	}
	rows := []usersmodel.User{
		{ID: DummyAdminID, Email: "admin@mail.ru", PasswordHash: hash, Role: "admin"},
		{ID: DummyUserID, Email: "user@mail.ru", PasswordHash: hash, Role: "user"},
	}
	for i := range rows {
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
