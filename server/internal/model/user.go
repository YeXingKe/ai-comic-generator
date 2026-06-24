package model

import "time"

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserAccount string    `gorm:"uniqueIndex;size:256;not null" json:"userAccount"`
	UserPassword string   `gorm:"size:512;not null" json:"-"`
	UserName    string    `gorm:"size:256" json:"userName"`
	UserAvatar  string    `gorm:"size:1024" json:"userAvatar"`
	UserProfile string    `gorm:"size:512" json:"userProfile"`
	UserRole    string    `gorm:"size:256;default:user" json:"userRole"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

func (User) TableName() string {
	return "user"
}

type UserVO struct {
	ID          uint      `json:"id"`
	UserAccount string    `json:"userAccount"`
	UserName    string    `json:"userName"`
	UserAvatar  string    `json:"userAvatar"`
	UserProfile string    `json:"userProfile"`
	UserRole    string    `json:"userRole"`
	CreateTime  time.Time `json:"createTime"`
}

func UserToVO(user *User) UserVO {
	return UserVO{
		ID:          user.ID,
		UserAccount: user.UserAccount,
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
		CreateTime:  user.CreateTime,
	}
}

type UserRegisterRequest struct {
	UserAccount  string `json:"userAccount" binding:"required,min=4"`
	UserPassword string `json:"userPassword" binding:"required,min=8"`
	CheckPassword string `json:"checkPassword" binding:"required,min=8"`
}

type UserLoginRequest struct {
	UserAccount  string `json:"userAccount" binding:"required"`
	UserPassword string `json:"userPassword" binding:"required"`
}

type UserQueryRequest struct {
	PageNum     int    `json:"pageNum"`
	PageSize    int    `json:"pageSize"`
	UserName    string `json:"userName"`
	UserAccount string `json:"userAccount"`
	UserRole    string `json:"userRole"`
}

type DeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}
