// Package oauth 提供了oauth 2.0 的server api支持
package oauth

import (
	"net"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/generates"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	log "logger"
)

const (
	oauthid   = "OAuthID"
	usertoken = "UCToken"
)

var (
	oauthSec     = []byte("00000000")
	userTokenSec = []byte("00000001")
	expDura      = 1 * time.Minute
)

func StartServer() {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// token store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// generate jwt access token
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte("00000000"), jwt.SigningMethodHS512))

	clientStore := store.NewClientStore()
	//TODO: how to init  & reload client info
	clientStore.Set("222222", &models.Client{
		ID:     "222222",
		Secret: "22222222",
		Domain: "http://localhost:9094",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	/*srv.SetPasswordAuthorizationHandler(func(username, password string) (userID string, err error) {
		if username == "test" && password == "test" {
			userID = "test"
		}
		return
	})*/
	srv.ClientInfoHandler = server.ClientFormHandler
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Errorf("Internal Error:%s", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Errorf("Response Error:%s", re.Error.Error())
	})

	hdl := Handler{srv: srv, man: manager}

	http.HandleFunc("/login", hdl.loginHandler)
	http.HandleFunc("/auth", hdl.authHandler)

	//负责记录redirect url
	http.HandleFunc("/authorize", hdl.authorizeHandler)

	http.HandleFunc("/grant", hdl.grantHandler)

	http.HandleFunc("/token", hdl.tokenHandler)

	http.HandleFunc("/test", hdl.testHandler)

	http.HandleFunc("/revoke", hdl.revokeHandler)
	lis, err := net.Listen("tcp", ":9096")
	if err != nil {
		panic(err)
	}
	if err := http.Serve(lis, nil); err != nil {
		panic(err)
	}
}
