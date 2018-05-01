package store

import (
	"github.com/docker/distribution/registry/auth/token"
	"github.com/goglue/docker-registry-oauth/model"
)

type (
	Storage interface {
		Login(*model.Account) error
		HasAccess(*model.AuthRequest) *token.ResourceActions
	}
)
