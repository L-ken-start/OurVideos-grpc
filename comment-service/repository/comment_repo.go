package repository

import (
	"errors"
	"gorm.io/gorm"
	"ourvideos/comment-service/model"
	"strings"
	"time"
)

type CommentRepository struct {
	DB *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{DB: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.DB.Create(comment).Error
}

func (r *CommentRepository) FindById(id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.DB.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) List(video_id uint, offset, limit int, sort string) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	query := r.DB.Model(&model.Comment{})

	if video_id != 0 {
		query = query.Where("video_id=?", video_id)

	}
	query.Count(&total)

	switch sort {
	case "latest":
		query = query.Order("created_at desc")
	case "popular":
		query = query.Order("created_at desc")
	}

	err := query.Offset(offset).Limit(limit).Find(&comments).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}

	return comments, total, nil

}

func (r *CommentRepository) CheckLiked(commentId uint, uid uint) (bool, error) {
	err := r.DB.Model(&model.Like{}).Where("comment_id=? and user_id=?", commentId, uid).
		First(&model.Like{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *CommentRepository) UpdateAvatarByUID(uid uint, avatar string) (int64, error) {
	result := r.DB.Model(&model.Comment{}).Where("uid = ?", uid).Update("avatar", avatar)
	return result.RowsAffected, result.Error
}

func (r *CommentRepository) UpdateCommentlike(id uint, like int) (int64, error) {

	err := r.DB.Model(&model.Comment{}).Where("id=?", id).
		UpdateColumn("like_count", gorm.Expr("like_count+?", like)).Error
	if err != nil {
		return 0, err
	}
	var newCount int64
	err = r.DB.Model(&model.Comment{}).
		Where("id=?", id).
		Select("like_count").Scan(&newCount).Error
	return newCount, err

}

func (r *CommentRepository) ToggleCommentlike(uid uint, cid uint) (int64, error) {
	var like_count int64
	var err error
	err = r.DB.Model(&model.Like{}).Create(&model.Like{
		UserID:    uid,
		VideoID:   0,
		CommentID: cid,
		CreatedAt: time.Now(),
	}).Error

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			r.DB.Model(&model.Like{}).Where("user_id=? and comment_id=?", uid, cid).Delete(&model.Like{})
			like_count, err = r.UpdateCommentlike(cid, -1)
			return like_count, nil
		}
		return 0, err
	}
	like_count, _ = r.UpdateCommentlike(cid, 1)
	return like_count, nil
}

//func (r *CommentRepository) ToggleVideolike(uid uint, vid uint) (int64, error) {
//	err := r.DB.Model(&model.Like{}).Create(&model.Like{
//		UserID:    uid,
//		VideoID:   vid,
//		CommentID: 0,
//		CreatedAt: time.Now(),
//	}).Error
//	if err != nil {
//		if strings.Contains(err.Error(), "Duplicate") {
//			r.DB.Model(&model.Like{}).Where("user_id=? and video_id=?", uid, vid).Delete(&model.Like{})
//			like_count, err := r.UpdateVideolike(vid, -1)
//			if err != nil {
//				return 0, err
//			} else {
//				return like_count, nil
//			}
//		}
//
//	}
//	like_count, err := r.UpdateVideolike(vid, -1)
//	if err != nil {
//		return 0, err
//	}
//
//	return like_count, err
//}
