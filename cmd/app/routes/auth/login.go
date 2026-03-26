package auth

import (
	"context"
	"net/http"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
)

func (s *AuthService) Login(ctx context.Context, spec *v1.LoginRequest, request *http.Request) (*v1.LoginResponse, int, error) {
	return nil, http.StatusOK, nil
}