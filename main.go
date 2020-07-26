package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var once sync.Once

var provider *oidc.Provider
var oauth2Config *oauth2.Config

func main() {

	// 認証で保護したいページ。ログインしていなければKeycloakのOpenID Connect認証ページに飛ばす
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// クッキーがない時はリダイレクト
		if _, err := r.Cookie("Authorization"); err != nil {
			config, _ := getConfig()
			url := config.AuthCodeURL("")
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		io.WriteString(w, "login success")
	})

	// OpenID Connectの認証が終わった時に呼ばれるハンドラ
	// もろもろトークンを取り出したりした後に、クッキーを設定して元のページに飛ばす
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		config, provider := getConfig()
		if err := r.ParseForm(); err != nil {
			http.Error(w, "parse form error", http.StatusInternalServerError)
			return
		}

		accessToken, err := config.Exchange(context.Background(), r.Form.Get("code"))
		if err != nil {
			http.Error(w, "Can't get access token", http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := accessToken.Extra("id_token").(string)
		if !ok {
			http.Error(w, "missing token", http.StatusInternalServerError)
			return
		}
		oidcConfig := &oidc.Config{
			ClientID: "test-app",
		}
		verifier := provider.Verifier(oidcConfig)
		idToken, err := verifier.Verify(context.Background(), rawIDToken)
		if err != nil {
			http.Error(w, "id token verify error: " + err.Error(), http.StatusInternalServerError)
			return
		}
		// IDトークンのクレームをとりあえずダンプ
		// アプリで必要なものはセッションストレージに入れておくと良いでしょう
		idTokenClaims := map[string]interface{}{}
		if err := idToken.Claims(&idTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("%#v", idTokenClaims)
		http.SetCookie(w, &http.Cookie{
			Name:  "Authorization",
			Value: "Bearer " + rawIDToken, // 行儀が悪いので真似しないねで
			Path:  "/",
		})
		http.Redirect(w, r, "/", http.StatusFound)
	})
	log.Println(http.ListenAndServe(":8080", nil))
}


func getConfig() (*oauth2.Config, *oidc.Provider) {
	once.Do(func() {
		var err error
		// ここにissuer情報を設定
		provider, err = oidc.NewProvider(context.Background(), "http://localhost:18080/auth/realms/master")
		if err != nil {
			panic(err)
		}
		oauth2Config = &oauth2.Config{
			// ここにクライアントIDとクライアントシークレットを設定
			ClientID:     "test-app",
			ClientSecret: "b4fd4da3-4a87-48bc-8327-ae50bdf2614c",
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID},
			RedirectURL:  "http://localhost:8080/callback",
		}
	})
	return oauth2Config, provider
}
