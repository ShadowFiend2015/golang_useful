package oauth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-session/session"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/utils/uuid"

	log "logger"
)

// Handler oauth2.0对应各个url的handler
type Handler struct {
	srv *server.Server
	man *manage.Manager
}

func (h *Handler) revokeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// verify id & secret
	clientid, clientsec := r.PostFormValue("client_id"), r.PostFormValue("client_secret")
	if clientid == "" || clientsec == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	cl, err := h.man.GetClient(clientid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if cl.GetSecret() != clientsec {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ret := make(map[string]interface{})
	ret["code"] = 0
	msg := ""
	token, refreshToken := r.PostFormValue("access_token"), r.PostFormValue("refresh_token")
	if refreshToken != "" {
		err = h.man.RemoveRefreshToken(refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		msg += "refreshToken revoked."
	}

	if token != "" {
		err = h.man.RemoveAccessToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		msg += "token revoked."
	}
	ret["msg"] = msg
	b, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json;charset=UTF-8")
	w.Write(b)
	return

}

func (h *Handler) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//1. 生成新的auth id
	var authid string
	if r.Form == nil {
		r.ParseForm()
	}

	uuidv, err := uuid.NewRandom()
	if err != nil {
		return
	}
	authid = uuidv.String()

	//add cookit
	cookie := http.Cookie{
		Name:    oauthid,
		Value:   authid,
		Expires: time.Now().Add(time.Hour),
	}

	http.SetCookie(w, &cookie)
	store.Set(authid, r.Form)

	store.Save()
	userID, verr := h.loginAuthorizeHandler(w, r)

	if verr != nil { //没有token
		//err = s.redirectError(w, req, verr)
		return
	} else if userID == "" { //无效token
		return
	}
}

func (h *Handler) grantHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var authid string
	var form url.Values
	ck, err := r.Cookie(oauthid)
	if err == nil { //取出auth id
		authid = ck.Value
		if v, ok := store.Get(authid); ok {
			form = v.(url.Values)
		} else {
			http.Error(w, "could not find redirect_url", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "could not find auth_ID", http.StatusBadRequest)
		return
	}
	r.Form = form
	store.Delete(authid)
	store.Save()

	err = h.srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Handler) testHandler(w http.ResponseWriter, r *http.Request) {
	token, err := h.srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"client_id":  token.GetClientID(),
		"user_id":    token.GetUserID(),
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(data)
}

func (h *Handler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	err := h.srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) loginAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	jwtck, err := r.Cookie(usertoken)
	//uid, ok := store.Get("LoggedInUserID")
	if err != nil { //未登录
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return userID, err
	}
	userID, err = isValidToken(jwtck.Value)
	if err == nil {
		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusFound)
	return userID, err

}

func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		jwtToken, err := h.generateJWT(r)
		if err != nil {
			log.Errorf("loginHandler:%v", err)
			http.Error(w, "login failed", http.StatusOK)
			return
		}
		jwtckie := http.Cookie{
			Name:    usertoken,
			Value:   jwtToken,
			Expires: time.Now().Add(time.Hour * 24),
		}
		http.SetCookie(w, &jwtckie)

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/login.html")
}

func (h *Handler) authHandler(w http.ResponseWriter, r *http.Request) {
	ck, err := r.Cookie(usertoken)
	if err != nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	_, err = isValidToken(ck.Value)
	if err != nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/auth.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	jwtck, err := r.Cookie(usertoken)
	//uid, ok := store.Get("LoggedInUserID")
	if err != nil { //未登录
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return userID, err
	}
	return isValidToken(jwtck.Value)
}
