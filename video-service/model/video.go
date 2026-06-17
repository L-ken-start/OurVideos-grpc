// video-service/model/video.go
package model

import "time"

type Video struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"type:varchar(256);not null;index"`
	Description string    `gorm:"type:text"`
	Category    string    `gorm:"type:varchar(64);not null;index"`
	PosterURL   string    `gorm:"column:poster_url;type:varchar(512)"`
	VideoURL    string    `gorm:"column:video_url;type:varchar(512)"`
	Duration    int       `gorm:"not null;default:0"`
	Tags        string    `gorm:"type:varchar(512)"`
	Year        int       `gorm:"not null;default:0"`
	Rating      float64   `gorm:"type:decimal(3,1);default:0"`
	PlayCount   int64     `gorm:"column:play_count;not null;default:0"`
	LikeCount   int64     `gorm:"column:like_count;not null;default:0"`
	UserID      uint      `gorm:"column:user_id;not null;index"`
	SeriesID    uint      `gorm:"column:series_id;not null;default:0"`
	Episode     uint      `gorm:"column:episode;not null;default:1"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Video) TableName() string {
	return "videos"
}
