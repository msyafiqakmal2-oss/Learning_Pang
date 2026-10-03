package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(t).Round(time.Microsecond))
	})
}

func main() {
	a := &App{
		store:    NewStore(),
		hasher:   NewHasher(os.Getenv("HASHER_URL")),
		secret:   []byte(getenv("JWT_SECRET", "dev-secret-ganti-di-produksi")),
		juryCode: getenv("JURY_CODE", "JURI2026"),
	}
	if getenv("SEED_DEMO", "true") == "true" {
		go a.seed()
	}

	siswa, guru, juri := a.only("siswa"), a.only("guru"), a.only("juri")
	guruJuri := a.only("guru", "juri")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", a.register)
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("GET /api/me", a.auth(a.me))
	mux.HandleFunc("GET /api/stats", a.auth(a.stats))
	mux.HandleFunc("GET /api/leaderboard", a.auth(a.leaderboard))

	mux.HandleFunc("GET /api/classes", a.auth(a.listClasses))
	mux.HandleFunc("POST /api/classes", a.auth(guru(a.createClass)))
	mux.HandleFunc("POST /api/classes/join", a.auth(siswa(a.joinClass)))
	mux.HandleFunc("DELETE /api/classes/{id}", a.auth(guru(a.deleteClass)))
	mux.HandleFunc("GET /api/classes/{id}/quizzes", a.auth(a.listQuizzes))
	mux.HandleFunc("POST /api/classes/{id}/quizzes", a.auth(guru(a.createQuiz)))
	mux.HandleFunc("DELETE /api/quizzes/{id}", a.auth(guru(a.deleteQuiz)))
	mux.HandleFunc("POST /api/quizzes/{id}/attempt", a.auth(siswa(a.attempt)))

	mux.HandleFunc("GET /api/classes/{id}/posts", a.auth(a.listPosts))
	mux.HandleFunc("POST /api/classes/{id}/posts", a.auth(a.createPost))
	mux.HandleFunc("POST /api/posts/{id}/like", a.auth(a.like))
	mux.HandleFunc("DELETE /api/posts/{id}", a.auth(a.deletePost))

	mux.HandleFunc("GET /api/projects", a.auth(a.listProjects))
	mux.HandleFunc("POST /api/projects", a.auth(siswa(a.createProject)))
	mux.HandleFunc("DELETE /api/projects/{id}", a.auth(a.only("siswa", "guru")(a.deleteProject)))
	mux.HandleFunc("POST /api/projects/{id}/score", a.auth(juri(a.scoreProject)))
	mux.HandleFunc("GET /api/audit", a.auth(guruJuri(a.audit)))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, map[string]string{"status": "ok"})
	})
	static := http.FileServer(http.Dir(getenv("WEB_DIR", "../web")))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		static.ServeHTTP(w, r)
	}))

	addr := ":" + getenv("PORT", "8080")
	log.Printf("EduNexus gateway (Go) berjalan di %s", addr)
	log.Fatal(http.ListenAndServe(addr, logging(mux)))
}
