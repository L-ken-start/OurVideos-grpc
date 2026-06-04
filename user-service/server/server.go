package server

import (
	"context"

	"ourvideos/proto/user"
	"ourvideos/user-service/service"
)

type UserServer struct {
	user.UnimplementedUserServiceServer
	Svc *service.UserService
}

// Register 注册
func (s *UserServer) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	u, token, err := s.Svc.Register(req.Username, req.Password, req.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &user.RegisterResp{
		Id:    uint64(u.ID),
		Token: token,
	}, nil
}

// Login 登录
func (s *UserServer) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	u, token, err := s.Svc.Login(req.Username, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &user.LoginResp{
		Token:  token,
		UserId: uint64(u.ID),
	}, nil
}

// GetUser 获取用户信息
func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserReq) (*user.GetUserResp, error) {
	u, err := s.Svc.GetUser(uint(req.Id))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &user.GetUserResp{
		Id:       uint64(u.ID),
		Username: u.Username,
		Email:    u.Email,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Bio:      u.Bio,
		Gender:   int32(u.Gender),
		Phone:    u.Phone,
	}, nil
}

// UserUpdate 更新用户信息
func (s *UserServer) UserUpdate(ctx context.Context, req *user.UserUpdateReq) (*user.UserUpdateResp, error) {
	u, err := s.Svc.UpdateUser(uint(req.Id), service.UserUpdateParams{
		Username: req.Username,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Bio:      req.Bio,
		Gender:   int8(req.Gender),
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &user.UserUpdateResp{
		Id:       uint64(u.ID),
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
		Bio:      u.Bio,
		Gender:   int32(u.Gender),
		Phone:    u.Phone,
	}, nil
}

func (s *UserServer) PswUpdate(ctx context.Context, req *user.PswUpdateReq) (*user.PswUpdateResp, error) {
	uid, err := s.Svc.PswUpdate(uint(req.Id), req.Oldpsw, req.Newpsw)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &user.PswUpdateResp{
		Id: uint32(uid),
	}, nil
}
