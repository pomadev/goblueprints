package main

import (
	"flag"
	"github.com/go-ini/ini"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/stretchr/signature"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

// 現在アクティブなAvatarの実装
var avatars Avatar = TryAvatars{
	UseFileSystemAvatar,
	UseAuthAvatar,
	UseGravatar,
}

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTPはHTTPリクエストを処理します
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ =
			template.Must(template.ParseFiles(filepath.Join("templates",
				t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "アプリケーションのアドレス")
	flag.Parse()
	// Gomniauthのセットアップ
	cfg, err := ini.Load("chat.ini")
	if err != nil {
		log.Fatal("Fail to read file: ", err)
	}
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	gomniauth.WithProviders(
		facebook.New(cfg.Section("auth").Key("facebook_client_key").String(), cfg.Section("auth").Key("facebook_secret_key").String(), "http://localhost:8080/auth/callback/facebook"),
		github.New(cfg.Section("auth").Key("github_client_key").String(), cfg.Section("auth").Key("github_secret_key").String(), "http://localhost:8080/auth/callback/github"),
		google.New(cfg.Section("auth").Key("google_client_key").String(), cfg.Section("auth").Key("google_secret_key").String(), "http://localhost:8080/auth/callback/google"),
	)
	r := newRoom()
	//r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	http.HandleFunc("/uploader", uploaderHandler)
	http.Handle("/room", r)
	http.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("./avatars"))))
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	// チャットルームを開始します
	go r.run()
	// Webサーバーを起動します
	log.Println("Webサーバーを開始します。ポート: ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
