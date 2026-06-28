package store

import (
	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/gorm"
)

// UserStore 用户数据访问层
type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *UserStore) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetByAccount(account string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("user_account = ?", account).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) CountByAccount(account string) (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("user_account = ?", account).Count(&count).Error
	return count, err
}

func (s *UserStore) CountByRole(role string) (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("user_role = ?", role).Count(&count).Error
	return count, err
}

func (s *UserStore) Delete(id uint) (int64, error) {
	result := s.db.Delete(&model.User{}, id)
	return result.RowsAffected, result.Error
}

func (s *UserStore) BuildQuery(userAccount, userName, userRole string) *gorm.DB {
	query := s.db.Model(&model.User{})
	if userAccount != "" {
		query = query.Where("user_account LIKE ?", "%"+userAccount+"%")
	}
	if userName != "" {
		query = query.Where("user_name LIKE ?", "%"+userName+"%")
	}
	if userRole != "" {
		query = query.Where("user_role = ?", userRole)
	}
	return query
}

func (s *UserStore) List(query *gorm.DB, pageNum, pageSize int) ([]model.User, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	offset := (pageNum - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
