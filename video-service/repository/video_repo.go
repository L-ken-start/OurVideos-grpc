package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"ourvideos/video-service/model"
	_ "ourvideos/video-service/model"
	"strconv"
	"strings"
	"time"
)

// VideoRepository 视频数据访问层
// 职责：纯数据库 CRUD，不包含任何业务逻辑
type VideoRepository struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func NewVideoRepository(db *gorm.DB, rdb *redis.Client) *VideoRepository {
	return &VideoRepository{DB: db, RDB: rdb}

}

// IncrPlayCount 播放量 +1（只写 Redis）
func (r *VideoRepository) IncrPlayCount(videoID uint) (int64, error) {
	//本质上是将关系型数据库中“表名 + 字段名 + 主键ID”的概念
	return r.RDB.Incr(context.Background(), fmt.Sprintf("video:play:%d", videoID)).Result()
}

func hotkey() string {
	return "video:hot:" + time.Now().Format("2006-01-02")
}

// AddHeat —— 播放+1分，点赞+2分，取消点赞-2分

func (r *VideoRepository) AddHeat(videoID uint, seriseID uint, delta int64) error {
	memberID := videoID
	if seriseID > 0 {
		memberID = seriseID
	}
	member := strconv.FormatUint(uint64(memberID), 10)

	key := "video:hot:" + time.Now().Format("2006-01-02")

	err := r.RDB.ZIncrBy(context.Background(), key, float64(delta), member).Err()
	if err != nil {
		return err
	}

	// 48 小时后自动清理旧 key
	r.RDB.Expire(context.Background(), key, 48*time.Hour)
	return nil
}

// GetHotVideoIDs —— 拿热度 Top N 的视频 ID
func (r *VideoRepository) GetHotVideos(limit int, category string) ([]model.Video, error) {
	key := "video:hot:" + time.Now().Format("2006-01-02")
	strs, err := r.RDB.ZRangeArgs(context.Background(), redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  int64(limit - 1),
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, err
	}
	var ids []uint
	for _, s := range strs {
		var id uint
		fmt.Sscanf(s, "%d", &id)
		ids = append(ids, id)
	}

	var videos []model.Video
	for _, id := range ids {

		var v model.Video
		query := r.DB.Where("id= ? or (series_id=? and episode=1)", id, id)
		if category != "" {
			query = query.Where("category = ?", category)
		}
		err := query.Order("series_id = 0 DESC").First(&v).Error
		if err == nil {
			videos = append(videos, v) // 查到了才加
		}
	}
	return videos, nil

}

// redis控制批量添加播放量
func (r *VideoRepository) SyncPlayCounts() error {
	keys, _ := r.RDB.Keys(context.Background(), "video:play:*").Result()
	for _, key := range keys {
		id := strings.TrimPrefix(key, "video:play:")
		count, _ := r.RDB.Get(context.Background(), key).Int64()
		r.RDB.Del(context.Background(), key)
		result := r.DB.Exec("UPDATE videos SET play_count = play_count + ? WHERE id = ?", count, id)
		if result.Error != nil {
			fmt.Printf("更新播放量失败 id=%s count=%d err=%v\n", id, count, result.Error)
		}
	}
	return nil

}

// 创建视频记录
func (r *VideoRepository) Create(video *model.Video) error {
	err := r.DB.Create(video).Error
	if err != nil {
		return err
	}
	if video.Category != "电影" && video.SeriesID == 0 {
		err = r.DB.Model(video).Update("series_id", gorm.Expr("id")).Error

	}
	return err
}

func (r *VideoRepository) UploadVideo(video *model.Video) error {
	err := r.DB.Create(video).Error
	if err != nil {
		return err
	}
	r.DB.Model(video).Update("series_id", video.SeriesID)
	return err
}

func (r *VideoRepository) CreateSeries(series *model.Series) error {
	err := r.DB.Create(series).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *VideoRepository) FindByID(id uint) (*model.Video, error) {
	var video model.Video
	err := r.DB.First(&video, id).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// 返回值：视频切片 + 总数（用于分页组件显示"共 42 条，第 1/3 页"）
func (r *VideoRepository) List(category string, sortBy string, userID uint, offset, limit int, tag string) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	query := r.DB.Table("videos").Select("*").Joins("LEFT JOIN series ON series.id = videos.series_id")

	if category != "" {

		query = query.Where("videos.category = ?", category)
	}
	if tag != "" {
		tag = "%" + tag + "%"
		query = query.Where("videos.tags like ? or series.tags like ?", tag, tag)
	}
	if userID != 0 {
		query = query.Where("series.user_id=?", userID)
	}
	//过滤
	query = query.Where("videos.series_id = 0 or videos.id=videos.series_id")

	query.Count(&total)

	switch sortBy {
	case "rating":
		query = query.Order("COALESCE(series.rating, videos.rating) DESC")
	default:
		query = query.Order("videos.created_at DESC")
	}
	err := query.Offset(offset).Limit(limit).Find(&videos).Error
	return videos, total, err

}

func (r *VideoRepository) FindBySeries(sid uint) ([]model.Video, error) {
	var videos []model.Video
	err := r.DB.Where("series_id=?", sid).Order("episode asc").Find(&videos).Error
	return videos, err
}

func (r *VideoRepository) FindSeriesByID(id uint) (*model.Series, error) {
	var s model.Series
	//var v model.Video
	//var playCount int64
	err := r.DB.First(&s, id).Error
	//r.DB.Model(v).Where("id=?", id).Select("play_count").Scan(&playCount)
	//fmt.Println("播放量为:", playCount)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ============================================================
// Search — 标题 + 标签模糊搜索
// ============================================================
func (r *VideoRepository) Search(query string, offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	like := "%" + query + "%"
	base := r.DB.Model(&model.Video{}).
		Where("title LIKE ?", like)

	base.Count(&total)

	err := base.Order("play_count desc").
		Offset(offset).
		Limit(limit).Find(&videos).Error
	return videos, total, err
}

// 增加播放量
func (r *VideoRepository) IncrementPlayCount(id uint) error {
	return r.DB.Model(&model.Video{}).Where("id=?", id).
		UpdateColumn("play_count", gorm.Expr("play_count+1")).Error
}

func (r *VideoRepository) CheckLiked(videoId uint, uid uint) (bool, error) {
	err := r.DB.Model(&model.Like{}).Where("video_id=? and user_id=?", videoId, uid).
		First(&model.Like{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *VideoRepository) UpdateVideoLike(videoId uint, like int64) (int64, error) {
	err := r.DB.Model(&model.Video{}).Where("id=?", videoId).
		UpdateColumn("like_count", gorm.Expr("like_count+?", like)).Error
	if err != nil {
		return 0, err
	}
	var newCount int64
	err = r.DB.Model(&model.Video{}).
		Where("id=?", videoId).Select("like_count").Scan(&newCount).Error
	if err != nil {
		return 0, err
	}
	return newCount, nil
}

func (r *VideoRepository) ToggleVideolike(vid uint, uid uint) (int64, bool, error) {
	var like_count int64
	var err error
	err = r.DB.Model(&model.Like{}).Create(&model.Like{
		UserID:    uid,
		VideoID:   vid,
		CommentID: 0,
		CreatedAt: time.Now(),
	}).Error

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			r.DB.Model(&model.Like{}).Where("user_id=? and video_id=?", uid, vid).Delete(&model.Like{})
			like_count, err = r.UpdateVideoLike(vid, -1)
			return like_count, false, nil
		}
		return 0, false, err
	}
	like_count, _ = r.UpdateVideoLike(vid, 1)
	return like_count, true, nil
}
