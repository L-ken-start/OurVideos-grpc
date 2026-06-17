// video-service/model/series.go
package model

import "time"

type Series struct {
	ID           uint      `gorm:"not null"`
	UserID       uint      `gorm:"not null"`
	Title        string    `gorm:"type:varchar(256);not null"`
	Description  string    `gorm:"type:text"`
	PosterURL    string    `gorm:"column:poster_url;type:varchar(512)"`
	Category     string    `gorm:"type:varchar(64);not null;index"`
	Tags         string    `gorm:"type:varchar(512)"`
	Year         int       `gorm:"not null;default:0"`
	Rating       float64   `gorm:"type:decimal(3,1);default:0"`
	EpisodeCount int       `gorm:"column:episode_count;not null;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Series) TableName() string {
	return "series"
}
