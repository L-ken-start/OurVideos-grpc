package service

import (
	"ourvideos/comment-service/model"
	"ourvideos/comment-service/repository"

	"time"
)

type CommentParam struct {
	Video_id   uint
	Uid        uint
	Username   string
	Avatar     string
	Content    string
	Like_count uint
	Create_at  time.Time
}

type CommentService struct {
	Repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{Repo: repo}
}

func (s *CommentService) AddComments(param CommentParam) (*model.Comment, error) {
	comment := &model.Comment{
		VideoID:   param.Video_id,
		UID:       param.Uid,
		Username:  param.Username,
		Avatar:    param.Avatar,
		Content:   param.Content,
		LikeCount: 0,
		CreatedAt: param.Create_at,
	}
	if err := s.Repo.Create(comment); err != nil {
		return nil, InternalError
	}

	return comment, nil

}

var allowedSortBy = map[string]bool{
	"latest": true, "popular": true,
}

func (s *CommentService) UpdateAvatar(uid uint, avatar string) (int64, error) {
	return s.Repo.UpdateAvatarByUID(uid, avatar)
}

func (s *CommentService) ListComments(videoId uint, offset, limit int, sort string) ([]model.Comment, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	if !allowedSortBy[sort] {
		sort = "latest"
	}
	return s.Repo.List(videoId, offset, limit, sort)

}

func (s *CommentService) CkeckLiked(commentId uint, uid uint) (bool, error) {
	return s.Repo.CheckLiked(commentId, uid)
}

func (s *CommentService) Updatelike(videoId uint, uid uint, commentId uint) (int64, error) {
	var like_count int64
	var err error
	if videoId == 0 {
		like_count, err = s.Repo.ToggleCommentlike(uid, commentId)
		if err != nil {
			return 0, LikeFailed
		}

	}
	return like_count, nil

}
