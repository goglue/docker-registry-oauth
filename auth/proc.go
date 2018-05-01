package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"strings"
	"github.com/goglue/docker-registry-oauth/model"
	"github.com/goglue/docker-registry-oauth/store"
	"github.com/goglue/docker-registry-oauth/store/memory"
	"time"
	"github.com/goglue/docker-registry-oauth/utils"
)

type (
	Processor interface {
		Authorize(c *gin.Context)
	}
	config struct {
		RegistryDomain string `env:"REGISTRY_SERVICE" envDefault:"http://localhost:5000"`
		IssuerDomain   string `env:"ISSUER_DOMAIN" envDefault:"http://localhost:4444"`
		PublicKey      string `env:"PUBLIC_KEY_PATH"`
		PrivateKey     string `env:"PRIVATE_KEY_PATH"`
		SignAlgorithm  string `env:"SIGN_ALGORITHM"`
		TokenDuration  int64  `env:"TOKEN_DURATION"`
	}
	processor struct {
		// The docker registry service name or domain name
		regDomain string
		// The Auth issuer name or domain name
		issDomain string
		// The public key
		publicKey libtrust.PublicKey
		// The private key
		privateKey libtrust.PrivateKey
		// The signing algorithm that will be used
		signAlgo string
		// When the token will be expired
		expiry int64
		// The DB where the accounts and the access are stored
		db store.Storage
	}
)

// authRequest method
func (ap *processor) authRequest(c *gin.Context) *model.AuthRequest {
	return &model.AuthRequest{
		Account:         ap.authAccount(c),
		ResourceActions: ap.resourceActions(c),
	}
}

// authUser method authorizes the user using Basic-Auth protocol, on failure it
// returns nil
func (ap *processor) authAccount(c *gin.Context) *model.Account {
	u, p, o := c.Request.BasicAuth()
	if !o {
		return nil
	}

	a := &model.Account{Username: u, Secret: p}
	if err := ap.db.Login(a); nil != err {
		return nil
	}
	a.Secret = ""

	return a
}

// resourceActions method receives the request and
func (ap *processor) resourceActions(c *gin.Context) *token.ResourceActions {
	service, scope := c.Query("service"), c.Query("scope")
	scopes := strings.Split(scope, ":")
	if len(scopes) < 3 || service != ap.regDomain {
		// FIXME: the user is doing login, fetch all his accesses
		return &token.ResourceActions{
			Type:"catalog",
			Actions: []string{"*"},
		}
	}

	return &token.ResourceActions{
		Type:    scopes[0],
		Name:    scopes[1],
		Actions: strings.Split(scopes[2], ","),
	}
}

func (ap *processor) issueToken(c *gin.Context, authRequest *model.AuthRequest) {
	now := time.Now().Unix()

	header := token.Header{
		Type:       "JWT",
		SigningAlg: ap.signAlgo,
		KeyID:      ap.publicKey.KeyID(),
	}
	headerJSON, _ := json.Marshal(header)

	claims := token.ClaimSet{
		Issuer:     ap.issDomain,
		Subject:    authRequest.Account.Username,
		Audience:   ap.regDomain,
		NotBefore:  now - 1,
		IssuedAt:   now,
		Expiration: now + ap.expiry,
		JWTID:      fmt.Sprintf("%d", rand.Int63()),
		Access: []*token.ResourceActions{
			{
				Type:    authRequest.ResourceActions.Type,
				Name:    authRequest.ResourceActions.Name,
				Actions: authRequest.ResourceActions.Actions,
			},
		},
	}

	claimsJSON, _ := json.Marshal(claims)
	payload := fmt.Sprintf("%s%s%s",
		joseBase64UrlEncode(headerJSON),
		token.TokenSeparator,
		joseBase64UrlEncode(claimsJSON),
	)
	sig, _, _ := ap.privateKey.Sign(strings.NewReader(payload), 0)
	signedToken := fmt.Sprintf("%s%s%s",
		payload,
		token.TokenSeparator,
		joseBase64UrlEncode(sig),
	)
	c.JSON(
		http.StatusOK,
		struct {
			Token string `json:"token"`
		}{
			Token: signedToken,
		},
	)
}
func (ap *processor) Authorize(c *gin.Context) {
	authRequest := ap.authRequest(c)

	if nil == authRequest ||
		nil == authRequest.Account ||
		nil == authRequest.ResourceActions {
		c.String(http.StatusUnauthorized, "%s", "Unauthorized request")
		return
	}

	ra := ap.db.HasAccess(authRequest)
	if nil == ra {
		c.String(http.StatusUnauthorized, "%s", "Unauthorized request")
		return
	}

	ap.issueToken(c, authRequest)
}

func joseBase64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func New() Processor {
	c := parseConf()

	priv, err := libtrust.FromCryptoPrivateKey(utils.PrivateKey(c.PrivateKey))
	if nil != err {
		panic(err)
	}

	pub, err := libtrust.FromCryptoPublicKey(
		utils.Certificate(c.PublicKey).PublicKey,
	)
	if nil != err {
		panic(err)
	}

	return &processor{
		regDomain:  c.RegistryDomain,
		issDomain:  c.IssuerDomain,
		publicKey:  pub,
		privateKey: priv,
		signAlgo:   c.SignAlgorithm,
		expiry:     c.TokenDuration,
		db:         memory.NewStore(),
	}
}

func parseConf() *config {
	c := new(config)
	env.Parse(c)

	return c
}
