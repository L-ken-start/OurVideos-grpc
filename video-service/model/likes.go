// video-service/model/like.go
package model

import "time"

type Like struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"column:user_id;not null;uniqueIndex:uk_user_target,priority:1"`
	VideoID   uint      `gorm:"column:video_id;not null;default:0;uniqueIndex:uk_user_target,priority:2"`
	CommentID uint      `gorm:"column:comment_id;not null;default:0;uniqueIndex:uk_user_target,priority:3"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Like) TableName() string {
	return "likes"
}
