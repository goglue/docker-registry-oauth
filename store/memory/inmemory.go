package memory

import (
	"errors"
	"github.com/caarlos0/env"
	"github.com/docker/distribution/registry/auth/token"
	"sync"
	"github.com/goglue/docker-registry-oauth/model"
	"github.com/goglue/docker-registry-oauth/store"
	"strings"
)

var (
	errNotFound = errors.New("account.not-found")
)

type (
	config struct {
		Backend      string   `env:"STORE_BACKEND"`
		InitAccounts []string `env:"STORE_INIT_ACCOUNTS" envSeparator:" "`
	}
	inmemStore struct {
		db map[string]string
		rw *sync.RWMutex
	}
)

func (im *inmemStore) HasAccess(a *model.AuthRequest) *token.ResourceActions {
	if nil != a.ResourceActions {
		return a.ResourceActions
	}

	return nil
}

func (im *inmemStore) Login(a *model.Account) error {
	im.rw.RLock()
	defer im.rw.RUnlock()

	p, ok := im.db[a.Username]
	if !ok || p != a.Secret {
		return errNotFound
	}

	return nil
}

func NewStore() store.Storage {
	mem := &inmemStore{
		db: make(map[string]string, 0),
		rw: new(sync.RWMutex),
	}

	c := parseConf()
	if len(c.InitAccounts) > 0 {
		mem.rw.Lock()
		defer mem.rw.Unlock()

		for k := range c.InitAccounts {
			acc := strings.Split(c.InitAccounts[k], ":")
			mem.db[acc[0]] = acc[1]
		}
	}

	return mem
}

func parseConf() *config {
	c := new(config)
	env.Parse(c)

	return c
}
