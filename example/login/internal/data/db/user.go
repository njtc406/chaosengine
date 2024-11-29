// Package db
// 模块名: 模块名
// 功能描述: 描述
// 作者:  yr  2024/11/16 0016 17:42
// 最后更新:  yr  2024/11/16 0016 17:42
package db

import (
	"errors"
	"gorm.io/gorm"
)

// User 用户表(这里只是单纯的用户表,后面如果加入sdk就单独给表,将id和sdk的id绑定)
type User struct {
	gorm.Model
	ID           int32  `json:"id" gorm:"primaryKey;autoIncrement"`                                                                                                                            // 用户ID
	DeviceID     string `json:"device_id" gorm:"type:varchar(128);not null"`                                                                                                                   // 设备ID
	OSVersion    string `json:"os_version" gorm:"type:varchar(64);default:null;index:idx_os_version"`                                                                                          // 操作系统版本
	OS           string `json:"os" gorm:"type:char(64);default:null;index:idx_os"`                                                                                                             // 系统
	LocalRegion  string `json:"local_region" gorm:"type:varchar(64);default:null;index:idx_local_region"`                                                                                      // 国家
	Channel      string `json:"channel" gorm:"type:varchar(64);not null;index:idx_channel;uniqueIndex:idx_channel_childchannel_account;uniqueIndex:idx_channel_child_channel"`                 // 渠道
	ChildChannel string `json:"child_channel" gorm:"type:varchar(64);default:null;index:idx_child_channel;uniqueIndex:idx_channel_childchannel_account;uniqueIndex:idx_channel_child_channel"` // 子渠道
	Phone        string `json:"phone" gorm:"type:varchar(64);default:null"`                                                                                                                    // 手机号码
	Account      string `json:"account" gorm:"type:varchar(32);default:null;index:idx_account;uniqueIndex:idx_channel_childchannel_account"`                                                   // 账号
	Password     string `json:"password" gorm:"type:varchar(32);default:null"`                                                                                                                 // 密码
	Secret       string `json:"secret" gorm:"type:varchar(128);not null"`                                                                                                                      // 密码校验密钥
	Email        string `json:"email" gorm:"type:varchar(64);default:null"`                                                                                                                    // 邮件地址
	DeleteTime   int64  `json:"delete_time" gorm:"type:bigint(20);default:null;index:idx_delete_time"`                                                                                         // 删除时间(删除账号时标记,这个时间过期后才执行真正的删除)
}

func (u *User) TableName() string {
	return "db_user"
}

func (u *User) Insert(tx *gorm.DB, _ ...interface{}) (interface{}, error) {
	if err := tx.Create(u).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

func (u *User) Get(tx *gorm.DB, args ...interface{}) (interface{}, error) {
	from := args[0].(map[string]interface{})
	err := tx.First(u, from).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		} else {
			return false, nil
		}
	}

	return true, nil
}

func (u *User) Update(db *gorm.DB, to map[string]interface{}) error {
	return db.Model(u).Updates(to).Error
}

func (u *User) Delete(db *gorm.DB, _ ...interface{}) (interface{}, error) {
	if err := db.Where("id=?", u.ID).Unscoped().Delete(u).Error; err != nil {
		return nil, err
	}
	return nil, nil
}
