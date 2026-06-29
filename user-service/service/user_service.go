package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"ourvideos/user-service/model"
	"ourvideos/user-service/repository"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

// generateJWT 生成 JWT token
func generateJWT(userId uint, username string) (string, error) {
	secret := []byte("ourvideos-secret-key")
	claims := jwt.MapClaims{
		"user_id":  userId,
		"username": username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// Register 注册新用户，返回用户对象和 token
func (s *UserService) Register(username, password, email string) (*model.User, string, error) {
	//   ↑ 方法接收者：告诉 Go 这个函数"挂"在 UserService 上
	exists, err := s.Repo.ExistsByUsername(username)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "query user failed")
	}
	if exists {
		return nil, "", ErrUserAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "bcrypt hash failed")
	}

	u := &model.User{
		Username: username,
		Password: string(hashed),
		Email:    email,
	}
	if err := s.Repo.Create(u); err != nil {
		return nil, "", status.Errorf(codes.Internal, "create user failed")
	}

	token, err := generateJWT(u.ID, u.Username)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "generate token failed")
	}

	return u, token, nil
}

// Login 登录，返回用户对象和 token
func (s *UserService) Login(username, password string) (*model.User, string, error) {
	u, err := s.Repo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", status.Errorf(codes.NotFound, "user not found")
		}
		return nil, "", status.Errorf(codes.Internal, "query user failed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", status.Errorf(codes.Unauthenticated, "invalid password")
	}

	token, err := generateJWT(u.ID, u.Username)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "generate token failed")
	}

	return u, token, nil
}

func (s *UserService) OAuthLogin(provider, openID, nickname, avatar string) (*model.User, string, error) {
	// 查数据库：这个 openID 来过没有
	u, err := s.Repo.FindByOpenID(openID)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "查询失败")

	}

	// 没来过 → 自动注册
	if u == nil {
		// 用 provider + openID 尾 8 位拼成唯一用户名，比如 "github_c4e8f2a1"
		u = &model.User{
			Username: provider + "_" + openID[len(openID)-8:],
			Nickname: nickname,
			Avatar:   avatar,
			OpenID:   openID,
			Provider: provider,
		}
		if err = s.Repo.Create(u); err != nil {
			return nil, "", status.Errorf(codes.Internal, "create user failed")
		}
	}

	// 签发 JWT（和密码登录用的同一个 generateJWT）
	token, err := generateJWT(u.ID, u.Username)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "generate token failed")
	}
	return u, token, nil

}

// GetUser 根据 ID 获取用户信息
func (s *UserService) GetUser(id uint) (*model.User, error) {
	u, err := s.Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "query user failed")
	}
	return u, nil
}

// UserUpdateParams 可更新的用户字段
type UserUpdateParams struct {
	Username string
	Email    string
	Avatar   string
	Bio      string
	Gender   int8
	Phone    string
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, params UserUpdateParams) (*model.User, error) {
	u, err := s.Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "query user failed")
	}

	updates := map[string]interface{}{
		"username": params.Username,
		"email":    params.Email,
		"avatar":   params.Avatar,
		"bio":      params.Bio,
		"gender":   params.Gender,
		"phone":    params.Phone,
	}

	if err := s.Repo.Update(u, updates); err != nil {
		return nil, status.Errorf(codes.Internal, "update user failed")
	}

	// 更新后重新查询，返回最新数据
	return s.Repo.FindByID(id)
}

func (s *UserService) PswUpdate(id uint, old_pwd string, new_pwd string) (uint, error) {
	u, err := s.Repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(old_pwd)); err != nil {
		return 0, ErrUserInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(new_pwd), bcrypt.DefaultCost)

	err = s.Repo.UpdatePassword(u.ID, string(hash))
	if err != nil {
		return 0, ErrorUpdateFailed
	}
	return u.ID, nil
}
