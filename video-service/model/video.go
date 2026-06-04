// video-service/model/video.go
package model

import "time"

// Video 视频数据库实体
//
// 设计决策：
//   - Tags 存为逗号分隔字符串而非 JSON 数组：
//     小规模项目用 LIKE 搜索标签就够了，JSON 字段在 MySQL 里索引效果差。
//     体量上来后可以拆成 video_tags 关联表。
//   - Category 用 varchar 而非 ENUM：
//     ENUM 改起来要 ALTER TABLE，varchar 加个校验逻辑就行。
//   - 播放量/点赞数用 int64：
//     MySQL BIGINT，够你用到地老天荒。
type Video struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"type:varchar(256);not null;index"`     // 标题索引：搜索必备
	Description string    `gorm:"type:text"`                            // TEXT 类型，不限长度
	Category    string    `gorm:"type:varchar(64);not null;index"`      // 分类索引：首页分类筛选
	PosterURL   string    `gorm:"column:poster_url;type:varchar(512)"`  // 封面图 URL
	VideoURL    string    `gorm:"column:video_url;type:varchar(512)"`   // 视频文件 URL
	Duration    int       `gorm:"not null;default:0"`                   // 时长（秒）
	Tags        string    `gorm:"type:varchar(512)"`                    // "科幻,冒险,史诗"
	Year        int       `gorm:"not null;default:0"`                   // 发行年份
	Rating      float64   `gorm:"type:decimal(3,1);default:0"`          // 评分 0.0-10.0，精确到 1 位小数
	PlayCount   int64     `gorm:"column:play_count;not null;default:0"` // 播放数
	LikeCount   int64     `gorm:"column:like_count;not null;default:0"` // 点赞数
	UserID      uint      `gorm:"column:user_id;not null;index"`        // 上传者 ID，索引用于"我的上传"
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名（GORM 默认用复数 videos，显式指定更清晰）
func (Video) TableName() string {
	return "videos"
}
