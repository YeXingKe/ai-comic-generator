package service

import (
	"errors"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists       = errors.New("账号已存在")
	ErrUserNotFound     = errors.New("用户不存在")
	ErrPasswordMismatch = errors.New("密码错误")
	ErrInvalidParams    = errors.New("参数错误")
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Register(req *model.UserRegisterRequest) (uint, error) {
	if req.UserPassword != req.CheckPassword {
		return 0, ErrInvalidParams
	}

	var count int64
	if err := s.db.Model(&model.User{}).Where("user_account = ?", req.UserAccount).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrUserExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := model.User{
		UserAccount:  req.UserAccount,
		UserPassword: string(hashed),
		UserName:     req.UserAccount,
		UserRole:     common.UserRole,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *UserService) Login(req *model.UserLoginRequest) (*model.UserVO, error) {
	var user model.User
	if err := s.db.Where("user_account = ?", req.UserAccount).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(req.UserPassword)); err != nil {
		return nil, ErrPasswordMismatch
	}

	vo := model.UserToVO(&user)
	return &vo, nil
}

func (s *UserService) GetByID(id uint) (*model.UserVO, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	vo := model.UserToVO(&user)
	return &vo, nil
}

func (s *UserService) ListPage(req *model.UserQueryRequest) (*common.PageResult, error) {
	pageNum := req.PageNum
	pageSize := req.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := s.db.Model(&model.User{})
	if req.UserAccount != "" {
		query = query.Where("user_account LIKE ?", "%"+req.UserAccount+"%")
	}
	if req.UserName != "" {
		query = query.Where("user_name LIKE ?", "%"+req.UserName+"%")
	}
	if req.UserRole != "" {
		query = query.Where("user_role = ?", req.UserRole)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.User
	offset := (pageNum - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, err
	}

	vos := make([]model.UserVO, 0, len(users))
	for i := range users {
		vos = append(vos, model.UserToVO(&users[i]))
	}

	return &common.PageResult{
		Records:  vos,
		TotalRow: total,
		PageNum:  pageNum,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) Delete(id uint) error {
	result := s.db.Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserService) EnsureAdmin() error {
	var count int64
	if err := s.db.Model(&model.User{}).Where("user_role = ?", common.AdminRole).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		UserAccount:  "admin",
		UserPassword: string(hashed),
		UserName:     "管理员",
		UserRole:     common.AdminRole,
		UserProfile:  "系统默认管理员",
	}
	return s.db.Create(&admin).Error
}
