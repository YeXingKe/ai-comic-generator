package service // 用户业务层：注册、登录、Session、资料与管理员 CRUD

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/gin-contrib/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	store *store.UserStore // 用户数据库操作
}

// NewUserService 创建用户服务
func NewUserService(store *store.UserStore) *UserService {
	return &UserService{store: store} // 注入 UserStore
}

// Register 用户注册
func (s *UserService) Register(req *model.RegisterRequest) (int64, error) {
	if req.UserAccount == "" || req.UserPassword == "" || req.CheckPassword == "" { // 必填项不能为空
		return 0, common.ErrParams.WithMessage("参数为空") // 参数错误
	}
	if len(req.UserAccount) < common.MinAccountLength { // 账号长度校验
		return 0, common.ErrParams.WithMessage("账号长度过短") // 账号太短
	}
	if len(req.UserPassword) < common.MinPasswordLength || len(req.CheckPassword) < common.MinPasswordLength { // 密码长度校验
		return 0, common.ErrParams.WithMessage("密码长度过短") // 密码太短
	}
	if req.UserPassword != req.CheckPassword { // 两次密码一致性
		return 0, common.ErrParams.WithMessage("两次输入的密码不一致") // 不一致则拒绝
	}

	count, err := s.store.CountByAccount(req.UserAccount) // 查询账号是否已存在
	if err != nil { // 数据库错误
		return 0, common.ErrSystem // 系统错误
	}
	if count > 0 { // 账号已注册
		return 0, common.ErrParams.WithMessage("账号重复") // 拒绝重复注册
	}

	hashed, err := hashPassword(req.UserPassword) // bcrypt 哈希明文密码
	if err != nil { // 哈希失败
		return 0, common.ErrSystem // 系统错误
	}
	userName := "无名"
	now := time.Now()
	user := &model.User{
		UserAccount:  req.UserAccount,
		UserPassword: hashed,
		UserName:     &userName,
		UserRole:     string(model.RoleUser),
		Points:       common.DefaultPoints,
		EditTime:     &now,
	}
	if err := s.store.Create(user); err != nil {
		return 0, common.ErrOperation.WithMessage("注册失败，数据库错误")
	}
	return user.ID, nil
}

// Login 用户登录，成功后将用户 ID 写入 Session
func (s *UserService) Login(req *model.LoginRequest, session sessions.Session) (*model.LoginUser, error) {
	if req.UserAccount == "" || req.UserPassword == "" { // 账号密码不能为空
		return nil, common.ErrParams.WithMessage("参数为空") // 参数错误
	}
	if len(req.UserAccount) < common.MinAccountLength { // 账号长度校验
		return nil, common.ErrParams.WithMessage("账号长度过短") // 账号太短
	}
	if len(req.UserPassword) < common.MinPasswordLength { // 密码长度校验
		return nil, common.ErrParams.WithMessage("密码长度过短") // 密码太短
	}

	user, err := s.store.GetByAccount(req.UserAccount)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrParams.WithMessage("用户不存在或密码错误")
		}
		return nil, common.ErrSystem
	}
	if !s.verifyPassword(user, req.UserPassword) {
		return nil, common.ErrParams.WithMessage("用户不存在或密码错误")
	}
	if user.Status == 0 {
		return nil, common.ErrParams.WithMessage("账号已被禁用，请联系管理员")
	}
	session.Set(common.UserLoginState, user.ID)
	if err := session.Save(); err != nil {
		return nil, common.ErrSystem
	}
	return user.ToLoginUser(), nil
}

// GetLoginUser 从 Session 读取用户 ID 并查询完整用户实体
func (s *UserService) GetLoginUser(session sessions.Session) (*model.User, error) {
	userID := session.Get(common.UserLoginState) // 从 Session 取用户 ID
	if userID == nil { // 未登录
		return nil, common.ErrNotLogin // 返回未登录错误
	}

	id, ok := userID.(int64) // 类型断言为 int64
	if !ok { // Session 中类型异常
		return nil, common.ErrNotLogin // 视为未登录
	}

	user, err := s.store.GetByID(id) // 按 ID 查库
	if err != nil { // 查询失败
		if errors.Is(err, gorm.ErrRecordNotFound) { // 用户已被删除
			return nil, common.ErrNotLogin // 视为未登录
		}
		return nil, common.ErrSystem // 系统错误
	}

	if user.Status == 0 {
		session.Delete(common.UserLoginState)
		_ = session.Save()
		return nil, common.ErrNotLogin.WithMessage("账号已被禁用，请联系管理员")
	}

	return user, nil // 返回完整用户实体
}

// Logout 清除 Session 中的登录态
func (s *UserService) Logout(session sessions.Session) error {
	if session.Get(common.UserLoginState) == nil { // 本来就没登录
		return common.ErrOperation.WithMessage("用户未登录") // 操作错误
	}
	session.Delete(common.UserLoginState) // 删除 Session 中的用户 ID
	return session.Save()                 // 持久化变更
}

// UpdateProfile 更新当前登录用户的昵称、头像、简介
func (s *UserService) UpdateProfile(session sessions.Session, req *model.UpdateProfileRequest) (*model.LoginUser, error) {
	user, err := s.GetLoginUser(session) // 先确认已登录并取当前用户
	if err != nil { // 未登录或 Session 无效
		return nil, err // 原样返回错误
	}

	now := time.Now() // 更新编辑时间
	updateUser := &model.User{ // 只包含要更新的字段
		ID:          user.ID,          // 目标用户 ID
		UserName:    req.UserName,     // 新昵称（可为 nil 表示不改）
		UserAvatar:  req.UserAvatar,   // 新头像
		UserProfile: req.UserProfile,  // 新简介
		EditTime:    &now,             // 编辑时间
	}
	if err := s.store.Update(updateUser); err != nil { // 执行更新
		return nil, common.ErrOperation // 更新失败
	}

	updated, err := s.store.GetByID(user.ID) // 重新查询最新数据
	if err != nil { // 查询失败
		return nil, common.ErrSystem // 系统错误
	}
	return updated.ToLoginUser(), nil // 返回更新后的登录用户信息
}

// UpdatePassword 修改当前登录用户密码
func (s *UserService) UpdatePassword(session sessions.Session, req *model.UpdatePasswordRequest) error {
	if req.OldPassword == "" || req.NewPassword == "" || req.CheckPassword == "" { // 三项必填
		return common.ErrParams.WithMessage("参数为空") // 参数错误
	}
	if len(req.NewPassword) < common.MinPasswordLength || len(req.CheckPassword) < common.MinPasswordLength { // 新密码长度
		return common.ErrParams.WithMessage("密码长度过短") // 太短
	}
	if req.NewPassword != req.CheckPassword { // 两次新密码一致
		return common.ErrParams.WithMessage("两次输入的密码不一致") // 不一致
	}
	if req.OldPassword == req.NewPassword { // 新旧不能相同
		return common.ErrParams.WithMessage("新密码不能与原密码相同") // 拒绝无意义修改
	}

	user, err := s.GetLoginUser(session) // 取当前登录用户
	if err != nil { // 未登录
		return err // 返回错误
	}

	if !s.verifyPassword(user, req.OldPassword) {
		return common.ErrParams.WithMessage("原密码错误")
	}
	hashed, err := hashPassword(req.NewPassword)
	if err != nil {
		return common.ErrSystem
	}
	if err := s.store.UpdatePassword(user.ID, hashed); err != nil {
		return common.ErrOperation
	}
	return nil
	return nil // 修改成功
}

// Create 管理员创建用户（默认密码 12345678）
func (s *UserService) Create(req *model.AddUserRequest) (int64, error) {
	hashed, err := hashPassword(common.DefaultPassword)
	if err != nil {
		return 0, common.ErrSystem
	}
	role := req.UserRole
	if role == "" {
		role = string(model.RoleUser)
	}
	if !model.UserRole(role).IsValid() {
		return 0, common.ErrParams.WithMessage("无效的用户角色")
	}

	now := time.Now()
	user := &model.User{
		UserAccount:  req.UserAccount,
		UserPassword: hashed,
		UserName:     req.UserName,
		UserAvatar:   req.UserAvatar,
		UserProfile:  req.UserProfile,
		UserRole:     role,
		Points:       common.DefaultPoints,
		Status:       1,
		EditTime:     &now,
	}
	if req.Points != nil {
		user.Points = *req.Points
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.store.Create(user); err != nil {
		return 0, common.ErrOperation
	}
	return user.ID, nil
}

// GetByID 根据 ID 获取用户（供中间件、Handler 等使用）
func (s *UserService) GetByID(id int64) (*model.User, error) {
	user, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrSystem
	}
	return user, nil
}

// Update 管理员更新用户信息（含角色、积分、状态）
func (s *UserService) Update(req *model.UpdateUserRequest) error {
	if req.UserRole != nil && !model.UserRole(*req.UserRole).IsValid() {
		return common.ErrParams.WithMessage("无效的用户角色")
	}

	user := &model.User{
		ID:          req.ID,
		UserName:    req.UserName,
		UserAvatar:  req.UserAvatar,
		UserProfile: req.UserProfile,
	}
	if req.UserRole != nil {
		user.UserRole = *req.UserRole
	}

	if err := s.store.Update(user); err != nil {
		return common.ErrOperation
	}
	if req.Points != nil {
		if err := s.store.UpdatePoints(req.ID, *req.Points); err != nil {
			return common.ErrOperation
		}
	}
	if req.Status != nil {
		if err := s.store.UpdateStatus(req.ID, *req.Status); err != nil {
			return common.ErrOperation
		}
	}
	return nil
}

// Delete 软删除用户
func (s *UserService) Delete(id int64) error {
	if err := s.store.Delete(id); err != nil { // 委托 Store 软删除
		return common.ErrOperation // 删除失败
	}
	return nil // 删除成功
}

// ListByPage 分页查询用户列表（管理员）
func (s *UserService) ListByPage(req *model.QueryUserRequest) (*model.PageResult, error) {
	query := s.store.BuildQuery( // 根据筛选条件构建 GORM 查询
		req.ID,          // 按 ID 筛选
		req.UserAccount, // 按账号筛选
		req.UserName,    // 按昵称筛选
		req.UserProfile, // 按简介筛选
		req.UserRole,    // 按角色筛选
		req.SortField,   // 排序字段
		req.SortOrder,   // 排序方向
	)

	users, total, err := s.store.List(query, req.PageNum, req.PageSize) // 分页查询
	if err != nil { // 数据库错误
		return nil, common.ErrSystem // 系统错误
	}

	userInfos := make([]model.UserInfo, 0, len(users)) // 预分配 API 响应切片
	for i := range users { // 遍历查询结果
		if info := users[i].ToUserInfo(); info != nil { // 实体转 UserInfo（脱敏）
			userInfos = append(userInfos, *info) // 追加到列表
		}
	}

	return &model.PageResult{ // 组装分页响应
		Total:    total,       // 总记录数
		Records:  userInfos,   // 当前页数据
		PageNum:  req.PageNum, // 页码
		PageSize: req.PageSize, // 每页条数
	}, nil
}


func encryptPassword(password, salt string) string {
	hash := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// hashPassword 新密码一律 bcrypt
func hashPassword(plain string) (string, error) {
	// DefaultCost  Go 库里定义的默认计算成本（cost）常量，数值一般是 10，越高越安全
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// verifyPassword 校验密码；旧 MD5 匹配成功则升级为 bcrypt
func (s *UserService) verifyPassword(user *model.User, plain string) bool {
	stored := user.UserPassword
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
	}
	if stored == encryptPassword(plain, common.PasswordSalt) {
		if newHash, err := hashPassword(plain); err == nil {
			if err := s.store.UpdatePassword(user.ID, newHash); err != nil {
				log.Printf("upgrade password to bcrypt failed, userId=%d: %v", user.ID, err)
			}
		}
		return true
	}
	return false
}
