package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ourvideos/video-service/model"
	"ourvideos/video-service/repository"
)

var allowedCategories = map[string]bool{
	"movie": true, "series": true, "anime": true, "documentary": true, "variety": true,
}

var allowedSortBy = map[string]bool{
	"latest": true, "popular": true, "rating": true,
}

type VideoService struct {
	Repo *repository.VideoRepository
}

func NewVideoService(repo *repository.VideoRepository) *VideoService {
	return &VideoService{Repo: repo}
}

type CreateVideoParams struct {
	Title       string
	Description string
	Category    string
	PosterURL   string
	VideoURL    string
	Duration    int
	Tags        string
	Year        int
	SeriesID    uint // 新增：0=电影，>0=归属某个系列
	Episode     uint // 新增：电影=0，电视剧/动漫=第几集
	UserID      uint
}

func (s *VideoService) CreateVideo(params CreateVideoParams) (*model.Video, error) {
	if !allowedCategories[params.Category] {
		return nil, ErrInvalidParam
	}
	fmt.Println("s")
	if strings.TrimSpace(params.Title) == "" {
		return nil, ErrInvalidParam
	}
	var video *model.Video
	Category := SwitchCategory(params.Category)
	if Category != "电影" {
		video = &model.Video{
			Title:     params.Title,
			Category:  Category,
			VideoURL:  params.VideoURL,
			PosterURL: params.PosterURL,
			Year:      params.Year,
			SeriesID:  0, // 新增
			Episode:   1, // 新增
		}
		if err := s.Repo.Create(video); err != nil {
			return nil, ErrInternal
		}
		fmt.Println("series_id:", video.SeriesID)
		series := &model.Series{
			ID:          video.SeriesID,
			UserID:      params.UserID,
			Title:       params.Title,
			Description: params.Description,
			PosterURL:   params.PosterURL,
			Category:    Category,
			Tags:        params.Tags,
			Year:        params.Year,
			Rating:      10.0,
		}
		if err := s.Repo.CreateSeries(series); err != nil {
			return nil, ErrInternal
		}
	} else {
		video = &model.Video{
			Title:       params.Title,
			Description: params.Description,
			Category:    Category,
			PosterURL:   params.PosterURL,
			VideoURL:    params.VideoURL,
			Duration:    params.Duration,
			Tags:        params.Tags,
			Year:        params.Year,
			SeriesID:    0, // 新增
			Episode:     1, // 新增
		}
		if err := s.Repo.Create(video); err != nil {
			return nil, ErrInternal
		}
	}

	return video, nil
}

func (s *VideoService) UploadVideo(params CreateVideoParams) (*model.Video, error) {
	video := &model.Video{
		Title:     params.Title,
		Category:  params.Category,
		PosterURL: params.PosterURL,
		VideoURL:  params.VideoURL,
		SeriesID:  params.SeriesID,
		Year:      params.Year,
		Episode:   params.Episode,
	}
	fmt.Println(params.SeriesID)
	if err := s.Repo.UploadVideo(video); err != nil {
		return nil, ErrInternal
	}
	return video, nil
}

func (s *VideoService) GetVideo(id uint) (*model.Video, error) {
	video, err := s.Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, ErrInternal
	}
	if video.SeriesID > 0 {
		if series, err := s.Repo.FindSeriesByID(video.SeriesID); err == nil {
			video.Description = series.Description
			video.PosterURL = series.PosterURL
			video.Category = series.Category
			video.Tags = series.Tags
			video.Year = series.Year
			video.Rating = series.Rating
		}
	}
	_ = s.Repo.IncrementPlayCount(id)
	video.PlayCount++

	return video, nil
}

func (s *VideoService) ListVideos(category string, sortBy string, userID uint, offset, limit int, tag string) ([]model.Video, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if !allowedSortBy[sortBy] {
		sortBy = "latest"
	}
	category = SwitchCategory(category)
	return s.Repo.List(category, sortBy, userID, offset, limit, tag)
}

func (s *VideoService) ListSeriesEpisodes(seriesID uint) ([]model.Video, error) {
	return s.Repo.FindBySeries(seriesID)
}

func (s *VideoService) SearchVideos(query string, offset, limit int) ([]model.Video, int64, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, 0, ErrInvalidParam
	}
	if len(query) > 100 {
		query = query[:100]
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit >= 50 {
		limit = 20
	}
	return s.Repo.Search(query, offset, limit)
}

// 转船前端发来的分类请求
func SwitchCategory(category string) string {
	switch category {
	case "movie":
		return "电影"
	case "series":
		return "电视剧"
	case "anime":
		return "动画"

	}
	return ""
}

func (s *VideoService) LikeVideo(vid uint, uid uint) (int64, bool, error) {
	like_count, is_liked, err := s.Repo.ToggleVideolike(vid, uid)
	if err != nil {
		return 0, is_liked, ErrLike
	}
	return like_count, is_liked, nil
}

func (s *VideoService) CkeckLiked(videoId uint, uid uint) (bool, error) {
	return s.Repo.CheckLiked(videoId, uid)
}
