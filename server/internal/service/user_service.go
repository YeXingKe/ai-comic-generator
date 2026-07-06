package service // 用户业务层：注册、登录、Session、资料与管理员 CRUD

import (
	"crypto/md5"   // 密码 MD5 哈希
	"encoding/hex" // 将哈希字节转为十六进制字符串
	"errors"       // 判断 gorm.ErrRecordNotFound
	"time"         // 记录编辑时间、VIP 时间

	"github.com/ai-comic-generator/server/internal/common" // 常量、业务错误
	"github.com/ai-comic-generator/server/internal/model"  // 用户实体与请求模型
	"github.com/ai-comic-generator/server/internal/store"  // 用户数据访问层
	"github.com/gin-contrib/sessions"                    // Session 读写（登录态）
	"gorm.io/gorm"                                         // 判断记录不存在错误
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

	userName := "无名" // 默认昵称
	now := time.Now()  // 当前时间作为编辑时间
	user := &model.User{ // 组装新用户实体
		UserAccount:  req.UserAccount,                                    // 登录账号
		UserPassword: encryptPassword(req.UserPassword, common.PasswordSalt), // MD5 加密密码
		UserName:     &userName,                                          // 默认昵称
		UserRole:     string(model.RoleUser),                             // 默认普通用户角色
		EditTime:     &now,                                               // 编辑时间
	}

	if err := s.store.Create(user); err != nil { // 写入数据库
		return 0, common.ErrOperation.WithMessage("注册失败，数据库错误") // 创建失败
	}

	return user.ID, nil // 返回新用户 ID
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

	user, err := s.store.GetByAccountAndPassword(req.UserAccount, encryptPassword(req.UserPassword, common.PasswordSalt)) // 按账号+加密密码查询
	if err != nil { // 查询失败
		if errors.Is(err, gorm.ErrRecordNotFound) { // 无匹配记录
			return nil, common.ErrParams.WithMessage("用户不存在或密码错误") // 统一提示，不暴露具体原因
		}
		return nil, common.ErrSystem // 其他数据库错误
	}

	session.Set(common.UserLoginState, user.ID) // 将用户 ID 写入 Session
	if err := session.Save(); err != nil { // 持久化 Session 到 Redis/Cookie
		return nil, common.ErrSystem // 保存失败
	}

	return user.ToLoginUser(), nil // 返回登录用户信息（不含密码）
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

	_, err = s.store.GetByAccountAndPassword(user.UserAccount, encryptPassword(req.OldPassword, common.PasswordSalt)) // 校验原密码
	if err != nil { // 校验失败
		if errors.Is(err, gorm.ErrRecordNotFound) { // 原密码不匹配
			return common.ErrParams.WithMessage("原密码错误") // 提示原密码错误
		}
		return common.ErrSystem // 系统错误
	}

	if err := s.store.UpdatePassword(user.ID, encryptPassword(req.NewPassword, common.PasswordSalt)); err != nil { // 写入新密码
		return common.ErrOperation // 更新失败
	}
	return nil // 修改成功
}

// EncryptPassword 根据明文密码与盐值生成数据库存储用的密码哈希（管理员工具接口）
func (s *UserService) EncryptPassword(password, salt string) (*model.EncryptPasswordResponse, error) {
	if password == "" { // 密码不能为空
		return nil, common.ErrParams.WithMessage("密码不能为空") // 参数错误
	}
	if salt == "" { // 未传盐值
		salt = common.PasswordSalt // 使用系统默认盐值
	}

	return &model.EncryptPasswordResponse{ // 返回加密结果
		EncryptedPassword: encryptPassword(password, salt), // MD5 哈希值
		Salt:              salt,                            // 实际使用的盐值
	}, nil
}

// Create 管理员创建用户（默认密码 12345678）
func (s *UserService) Create(req *model.AddUserRequest) (int64, error) {
	now := time.Now() // 当前时间
	user := &model.User{ // 组装用户实体
		UserAccount:  req.UserAccount,                                              // 登录账号
		UserPassword: encryptPassword(common.DefaultPassword, common.PasswordSalt), // 默认密码加密
		UserName:     req.UserName,                                                 // 昵称
		UserAvatar:   req.UserAvatar,                                               // 头像
		UserProfile:  req.UserProfile,                                              // 简介
		UserRole:     req.UserRole,                                                 // 角色
		EditTime:     &now,                                                         // 编辑时间
	}
	if req.Quota != nil { // 指定了额度
		user.Quota = *req.Quota // 写入额度
	}
	if req.UserRole == string(model.RoleVIP) { // 创建为 VIP 用户
		if req.VipTime != nil { // 指定了 VIP 到期时间
			user.VipTime = req.VipTime // 使用指定时间
		} else { // 未指定
			user.VipTime = &now // 默认从当前时间起算
		}
	}

	if err := s.store.Create(user); err != nil { // 写入数据库
		return 0, common.ErrOperation // 创建失败
	}
	return user.ID, nil // 返回新用户 ID
}

// GetByID 根据 ID 获取用户（供中间件、Handler 等使用）
func (s *UserService) GetByID(id int64) (*model.User, error) {
	user, err := s.store.GetByID(id) // 查库
	if err != nil { // 查询失败
		if errors.Is(err, gorm.ErrRecordNotFound) { // 不存在
			return nil, common.ErrNotFound // 未找到
		}
		return nil, common.ErrSystem // 系统错误
	}
	return user, nil // 返回用户实体
}

// Update 管理员更新用户信息（含角色、额度、VIP 时间）
func (s *UserService) Update(req *model.UpdateUserRequest) error {
	user := &model.User{ // 基础更新字段
		ID:          req.ID,          // 目标用户 ID
		UserName:    req.UserName,    // 昵称
		UserAvatar:  req.UserAvatar,  // 头像
		UserProfile: req.UserProfile, // 简介
	}
	if req.UserRole != nil { // 要改角色
		user.UserRole = *req.UserRole // 写入新角色
	}

	if err := s.store.Update(user); err != nil { // 更新基础字段
		return common.ErrOperation // 失败
	}
	if req.Quota != nil { // 要改额度
		if err := s.store.UpdateQuota(req.ID, *req.Quota); err != nil { // 单独更新额度列
			return common.ErrOperation // 失败
		}
	}
	if req.UserRole != nil && *req.UserRole != string(model.RoleVIP) { // 改为非 VIP
		if err := s.store.UpdateVipTime(req.ID, nil); err != nil { // 清空 VIP 时间
			return common.ErrOperation // 失败
		}
	} else if req.VipTime != nil { // 指定了 VIP 时间
		if err := s.store.UpdateVipTime(req.ID, req.VipTime); err != nil { // 更新 VIP 时间
			return common.ErrOperation // 失败
		}
	}
	return nil // 全部更新成功
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

// encryptPassword 使用盐值加密密码：MD5(密码 + 盐值)
func encryptPassword(password, salt string) string {
	hash := md5.Sum([]byte(password + salt)) // 计算 MD5 字节数组
	return hex.EncodeToString(hash[:])       // 转为 32 位十六进制字符串
}
