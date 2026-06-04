package repository

import (
	"gorm.io/gorm"
	"ourvideos/video-service/model"
	_ "ourvideos/video-service/model"
)

// VideoRepository 视频数据访问层
// 职责：纯数据库 CRUD，不包含任何业务逻辑
type VideoRepository struct {
	DB *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{DB: db}
}

// 创建视频记录
func (r *VideoRepository) Create(video *model.Video) error {
	return r.DB.Create(video).Error
}

func (r *VideoRepository) FindByID(id uint) (*model.Video, error) {
	var video model.Video
	err := r.DB.First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}
