package server

import (
	"context"

	"ourvideos/proto/video"
	servicemodel "ourvideos/video-service/model"
	"ourvideos/video-service/service"
)

// VideoServer gRPC 服务端 —— 只做 proto ↔ service 翻译
type VideoServer struct {
	video.UnimplementedVideoServiceServer
	Svc *service.VideoService
}

func modelToProto(v *servicemodel.Video) *video.VideoInfo {
	return &video.VideoInfo{
		Id:          uint64(v.ID),
		Title:       v.Title,
		Description: v.Description,
		Category:    v.Category,
		PosterUrl:   v.PosterURL,
		VideoUrl:    v.VideoURL,
		Duration:    int32(v.Duration),
		Tags:        v.Tags,
		Year:        int32(v.Year),
		Rating:      v.Rating,
		PlayCount:   v.PlayCount,
		LikeCount:   v.LikeCount,
		UserId:      uint64(v.UserID),
		CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *VideoServer) CreateVideo(ctx context.Context, req *video.CreateVideoReq) (*video.CreateVideoResp, error) {
	v, err := s.Svc.CreateVideo(service.CreateVideoParams{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		PosterURL:   req.PosterUrl,
		VideoURL:    req.VideoUrl,
		Duration:    int(req.Duration),
		Tags:        req.Tags,
		Year:        int(req.Year),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &video.CreateVideoResp{Video: modelToProto(v)}, nil
}

func (s *VideoServer) GetVideo(ctx context.Context, req *video.GetVideoReq) (*video.GetVideoResp, error) {
	v, err := s.Svc.GetVideo(uint(req.Id))
	liked, err := s.Svc.CkeckLiked(uint(req.Id), uint(req.Uid))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &video.GetVideoResp{Video: modelToProto(v), IsLiked: liked}, nil
}

func (s *VideoServer) ListVideo(ctx context.Context, req *video.ListVideoReq) (*video.ListVideoResp, error) {
	videos, total, err := s.Svc.ListVideos(
		req.Category,
		sortByToString(req.SortBy),
		uint(req.UserId),
		int(req.Offset),
		int(req.Limit),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbVideos := make([]*video.VideoInfo, len(videos))
	for i, v := range videos {
		pbVideos[i] = modelToProto(&v)
	}
	return &video.ListVideoResp{Videos: pbVideos, Total: int32(total)}, nil
}

func (s *VideoServer) SearchVideo(ctx context.Context, req *video.SearchVideosReq) (*video.SearchVideosResp, error) {
	videos, total, err := s.Svc.SearchVideos(req.Query, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbVideos := make([]*video.VideoInfo, len(videos))
	for i, v := range videos {
		pbVideos[i] = modelToProto(&v)
	}
	return &video.SearchVideosResp{Videos: pbVideos, Total: int32(total)}, nil
}

func (s *VideoServer) LikeVideo(ctx context.Context, req *video.LikeVideoReq) (*video.LikeVideoResp, error) {

	like_count, is_liked, err := s.Svc.LikeVideo(uint(req.VideoId), uint(req.Uid))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &video.LikeVideoResp{
		LikeCount: like_count,
		IsLiked:   is_liked,
	}, nil
}

func sortByToString(s video.SortBy) string {
	switch s {
	case video.SortBy_SORT_BY_POPULAR:
		return "popular"
	case video.SortBy_SORT_BY_RATING:
		return "rating"
	default:
		return "latest"
	}
}
