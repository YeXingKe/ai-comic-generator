package store

import (
	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/gorm"
)

// 数据访问层

// NotDeleted 软删除过滤 Scope
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("isDelete = ?", 0)
}

// UserStore 用户数据存储
type UserStore struct {
	db *gorm.DB
}

// NewUserStore 创建用户存储实例
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

// Create 创建用户
func (s *UserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}

// GetByID 根据 ID 获取用户
func (s *UserStore) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := s.db.Scopes(NotDeleted).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListByIDs 按 ID 列表批量查询用户（跳过已软删除）
func (s *UserStore) ListByIDs(ids []int64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	err := s.db.Scopes(NotDeleted).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetByAccount 根据账号获取用户
func (s *UserStore) GetByAccount(account string) (*model.User, error) {
	var user model.User
	err := s.db.Scopes(NotDeleted).Where("userAccount = ?", account).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByAccountAndPassword 根据账号和密码获取用户
func (s *UserStore) GetByAccountAndPassword(account, password string) (*model.User, error) {
	var user model.User
	err := s.db.Scopes(NotDeleted).
		Where("userAccount = ? AND userPassword = ?", account, password).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePoints 更新用户积分
func (s *UserStore) UpdatePoints(id int64, points int) error {
	return s.db.Model(&model.User{}).Scopes(NotDeleted).Where("id = ?", id).Update("points", points).Error
}

// UpdateStatus 更新用户启用状态
func (s *UserStore) UpdateStatus(id int64, status int) error {
	return s.db.Model(&model.User{}).Scopes(NotDeleted).Where("id = ?", id).Update("status", status).Error
}

// Update 更新用户
func (s *UserStore) Update(user *model.User) error {
	return s.db.Scopes(NotDeleted).Where("id = ?", user.ID).Updates(user).Error
}

// UpdatePassword 更新用户密码
func (s *UserStore) UpdatePassword(id int64, password string) error {
	return s.db.Model(&model.User{}).Scopes(NotDeleted).Where("id = ?", id).Update("userPassword", password).Error
}

// Delete 删除用户（逻辑删除）
func (s *UserStore) Delete(id int64) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("isDelete", 1).Error
}

// CountByAccount 根据账号统计用户数
func (s *UserStore) CountByAccount(account string) (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Scopes(NotDeleted).
		Where("userAccount = ?", account).Count(&count).Error
	return count, err
}

// CountByRole 根据角色统计用户数
func (s *UserStore) CountByRole(role string) (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Scopes(NotDeleted).
		Where("userRole = ?", role).Count(&count).Error
	return count, err
}

// List 分页查询用户列表
func (s *UserStore) List(query *gorm.DB, pageNum, pageSize int64) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 统计总数
	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// BuildQuery 构建查询条件
func (s *UserStore) BuildQuery(id *int64, userAccount, userName, userProfile, userRole *string, sortField, sortOrder *string) *gorm.DB {
	query := s.db.Scopes(NotDeleted)

	if id != nil {
		query = query.Where("id = ?", *id)
	}
	if userRole != nil && *userRole != "" {
		query = query.Where("userRole = ?", *userRole)
	}
	if userAccount != nil && *userAccount != "" {
		query = query.Where("userAccount LIKE ?", "%"+*userAccount+"%")
	}
	if userName != nil && *userName != "" {
		query = query.Where("userName LIKE ?", "%"+*userName+"%")
	}
	if userProfile != nil && *userProfile != "" {
		query = query.Where("userProfile LIKE ?", "%"+*userProfile+"%")
	}

	// 排序
	if sortField != nil && *sortField != "" {
		order := "ASC"
		if sortOrder != nil && *sortOrder == "descend" {
			order = "DESC"
		}
		query = query.Order(*sortField + " " + order)
	}

	return query
}

// DecrementPoints 原子扣减用户积分（默认扣 1）
// 使用 points > 0 条件确保并发安全，避免超扣
// 返回影响行数：1 表示成功，0 表示积分不足
func (s *UserStore) DecrementPoints(userID int64) (int64, error) {
	return s.DecrementPointsBy(userID, 1)
}

// DecrementPointsBy 原子扣减指定积分
func (s *UserStore) DecrementPointsBy(userID int64, amount int) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}
	result := s.db.Exec("UPDATE user SET points = points - ? WHERE id = ? AND points >= ?", amount, userID, amount)
	return result.RowsAffected, result.Error
}

// AddPoints 原子增加用户积分
func (s *UserStore) AddPoints(userID int64, amount int) error {
	if amount == 0 {
		return nil
	}
	return s.db.Exec("UPDATE user SET points = points + ? WHERE id = ?", amount, userID).Error
}
