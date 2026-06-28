package service

import (
	"errors"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
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
	store *store.UserStore
}

func NewUserService(userStore *store.UserStore) *UserService {
	return &UserService{store: userStore}
}

func (s *UserService) Register(req *model.UserRegisterRequest) (uint, error) {
	if req.UserPassword != req.CheckPassword {
		return 0, ErrInvalidParams
	}

	count, err := s.store.CountByAccount(req.UserAccount)
	if err != nil {
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

	if err := s.store.Create(&user); err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *UserService) Login(req *model.UserLoginRequest) (*model.UserVO, error) {
	user, err := s.store.GetByAccount(req.UserAccount)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(req.UserPassword)); err != nil {
		return nil, ErrPasswordMismatch
	}

	vo := model.UserToVO(user)
	return &vo, nil
}

func (s *UserService) GetByID(id uint) (*model.UserVO, error) {
	user, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	vo := model.UserToVO(user)
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

	query := s.store.BuildQuery(req.UserAccount, req.UserName, req.UserRole)
	users, total, err := s.store.List(query, pageNum, pageSize)
	if err != nil {
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
	affected, err := s.store.Delete(id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserService) EnsureAdmin() error {
	count, err := s.store.CountByRole(common.AdminRole)
	if err != nil {
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
	return s.store.Create(&admin)
}
