package handlers

import "golive/auth/internal/application/auth"

type AuthHandler struct {
	Register *auth.RegisterUseCase
	Login    *auth.LoginUseCase
	Refresh  *auth.RefreshUseCase
	Logout   *auth.LogoutUseCase
}

