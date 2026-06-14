package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ourvideos/comment-service/model"
	"ourvideos/comment-service/service"
	"ourvideos/proto/comment"
	"time"
)

type CommentServer struct {
	comment.UnimplementedCommentServiceServer
	Svc *service.CommentService
}

func NewCommentServer(svc *service.CommentService) *CommentServer {
	return &CommentServer{Svc: svc}
}

func modelToproto(m *model.Comment) *comment.Comment {
	return &comment.Comment{
		Id:        uint64(m.ID),
		VideoId:   uint64(m.VideoID),
		Uid:       uint64(m.UID),
		Username:  m.Username,
		Content:   m.Content,
		CreateAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
		Avatar:    m.Avatar,
		LikeCount: m.LikeCount,
		IsLiked:   false,
	}
}

func (s *CommentServer) AddComments(ctx context.Context, req *comment.AddCommentsReq) (*comment.AddCommentsResp, error) {
	CreateAt, err := time.Parse("2006-01-02 15:04:05", req.CreateAt)
	if err != nil {
		// 格式不对，返回错误
		return nil, status.Errorf(codes.InvalidArgument, "时间格式错误")
	}
	c, err := s.Svc.AddComments(service.CommentParam{
		Video_id:  uint(req.VideoId),
		Uid:       uint(req.Uid),
		Username:  req.Username,
		Content:   req.Content,
		Avatar:    req.Avatar,
		Create_at: CreateAt,
	})

	if err != nil {
		return nil, ErrorHandler(err)
	}
	return &comment.AddCommentsResp{Comment: modelToproto(c)}, nil

}

func (s *CommentServer) UpdateCommentAvatar(ctx context.Context, req *comment.UpdateAvatarReq) (*comment.UpdateAvatarResp, error) {
	rows, err := s.Svc.UpdateAvatar(uint(req.Uid), req.Avatar)
	if err != nil {
		return nil, ErrorHandler(err)
	}
	return &comment.UpdateAvatarResp{RowsAffected: rows}, nil
}

func (s *CommentServer) ListComments(ctx context.Context, req *comment.ListCommentsReq) (*comment.ListCommentsResp, error) {
	comments, total, err := s.Svc.ListComments(uint(req.VideoId), int(req.Offset), int(req.Limit), req.Sort)
	if err != nil {
		return nil, ErrorHandler(err)
	}

	pbComments := make([]*comment.Comment, len(comments))
	for i, comment := range comments {
		pbComments[i] = modelToproto(&comment)
		liked, err := s.Svc.CkeckLiked(uint(pbComments[i].Id), uint(pbComments[i].Uid))
		if err != nil {
			return nil, ErrorHandler(err)
		}
		pbComments[i].IsLiked = liked
	}
	return &comment.ListCommentsResp{Comment: pbComments, Total: int32(total)}, nil

}

func (s *CommentServer) LikeComment(ctx context.Context, req *comment.LikeCommentReq) (*comment.LikeCommentResp, error) {
	like_count, err := s.Svc.Updatelike(uint(req.VideoId), uint(req.Uid), uint(req.CommentId))
	if err != nil {
		return nil, ErrorHandler(err)
	}
	return &comment.LikeCommentResp{LikeCount: like_count}, nil
}
