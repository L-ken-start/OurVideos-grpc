package repository

import (
	"gorm.io/gorm"
	"ourvideos/user-service/model"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Create 创建新用户
func (r *UserRepository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据 ID 查找用户
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername 检查用户名是否已存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Update 更新用户字段
func (r *UserRepository) Update(user *model.User, updates map[string]interface{}) error {
	return r.DB.Model(user).Updates(updates).Error
}

func (r *UserRepository) UpdateProfile(id uint, user *model.User) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Updates(user).
		Select("username", "email", "avatar", "bio", "gender", "phone").
		Updates(user).Error

}

func (r *UserRepository) UpdatePassword(id uint, hashpassword string) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).
		Update("password", hashpassword).Error
}
