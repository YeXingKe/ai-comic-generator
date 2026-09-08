package model

import "time"

type User struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`                             // 主键 ID
	UserAccount  string     `gorm:"column:userAccount;uniqueIndex:uk_userAccount" json:"userAccount"` // 登录账号（唯一）
	UserPassword string     `gorm:"column:userPassword" json:"-"`                                   // 加密后的密码（不返回前端）
	UserName     *string    `gorm:"column:userName;index:idx_userName" json:"userName"`             // 用户昵称（可为空）
	UserAvatar   *string    `gorm:"column:userAvatar" json:"userAvatar"`                            // 头像 URL（可为空）
	UserProfile  *string    `gorm:"column:userProfile" json:"userProfile"`                          // 个人简介（可为空）
	UserRole     string     `gorm:"column:userRole;default:user" json:"userRole" enums:"user,admin" example:"user"` // 用户角色：user / admin
	Status       int        `gorm:"column:status;default:1" json:"status"`                          // 用户状态：1 启用，0 禁用
	Points       int        `gorm:"column:points;default:100" json:"points"`                        // 积分（创作等业务统一扣减，无 VIP 特权）
	EditTime     *time.Time `gorm:"column:editTime" json:"editTime"`                                // 资料最后编辑时间
	CreateTime   time.Time  `gorm:"column:createTime;autoCreateTime" json:"createTime"`             // 注册时间
	UpdateTime   time.Time  `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`             // 最后更新时间
	IsDelete     int        `gorm:"column:isDelete;default:0" json:"-"`                             // 软删除标记：0 正常，1 已删除
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// LoginUser 登录用户信息（响应）
type LoginUser struct {
	ID          int64      `json:"id"`
	UserAccount string     `json:"userAccount"`
	UserName    *string    `json:"userName"`
	UserAvatar  *string    `json:"userAvatar"`
	UserProfile *string    `json:"userProfile"`
	UserRole    string     `json:"userRole" enums:"user,admin" example:"user"`
	Status      int        `json:"status"`
	Points      int        `json:"points"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
	EditTime    *time.Time `json:"editTime"`
}

// UserInfo 用户信息（响应）
type UserInfo struct {
	ID          int64      `json:"id"`
	UserAccount string     `json:"userAccount"`
	UserName    *string    `json:"userName"`
	UserAvatar  *string    `json:"userAvatar"`
	UserProfile *string    `json:"userProfile"`
	UserRole    string     `json:"userRole" enums:"user,admin" example:"user"`
	Status      int        `json:"status"`
	Points      int        `json:"points"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
	EditTime    *time.Time `json:"editTime"`
}

// ToLoginUser 转换为登录用户信息
func (u *User) ToLoginUser() *LoginUser {
	if u == nil {
		return nil
	}
	return &LoginUser{
		ID:          u.ID,
		UserAccount: u.UserAccount,
		UserName:    u.UserName,
		UserAvatar:  u.UserAvatar,
		UserProfile: u.UserProfile,
		UserRole:    u.UserRole,
		Status:      u.Status,
		Points:      u.Points,
		CreateTime:  u.CreateTime,
		UpdateTime:  u.UpdateTime,
		EditTime:    u.EditTime,
	}
}

// ToUserInfo 转换为用户信息
func (u *User) ToUserInfo() *UserInfo {
	if u == nil {
		return nil
	}
	return &UserInfo{
		ID:          u.ID,
		UserAccount: u.UserAccount,
		UserName:    u.UserName,
		UserAvatar:  u.UserAvatar,
		UserProfile: u.UserProfile,
		UserRole:    u.UserRole,
		Status:      u.Status,
		Points:      u.Points,
		CreateTime:  u.CreateTime,
		UpdateTime:  u.UpdateTime,
		EditTime:    u.EditTime,
	}
}

// UserRole 用户角色
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// IsValid 判断角色是否有效
func (r UserRole) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}
