// video-service/model/comment.go
package model

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	VideoID   uint      `gorm:"column:video_id;not null;index:idx_video_sort,priority:1"`
	UID       uint      `gorm:"column:uid;not null;index:idx_uid"`
	Username  string    `gorm:"type:varchar(64);not null;default:''"`
	Avatar    string    `gorm:"type:varchar(512);not null;default:''"`
	Content   string    `gorm:"type:text;not null"`
	LikeCount int64     `gorm:"column:like_count;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (Comment) TableName() string {
	return "comments"
}
