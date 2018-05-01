package model

import "github.com/docker/distribution/registry/auth/token"

type (
	AuthRequest struct {
		Account         *Account
		ResourceActions *token.ResourceActions
	}
	Account struct {
		Username string `json:"username"`
		Secret   string `json:"secret"`
	}
)
