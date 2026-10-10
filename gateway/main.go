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

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(c int) { w.code = c; w.ResponseWriter.WriteHeader(c) }

// wrap: log request + tandai data "berubah" setelah aksi tulis yang sukses.
func (a *App) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		sw := &statusWriter{w, 200}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.code, time.Since(t).Round(time.Microsecond))
		if r.Method != http.MethodGet && sw.code < 400 {
			a.markDirty()
		}
	})
}

func main() {
	a := &App{
		store:    NewStore(),
		hasher:   NewHasher(os.Getenv("HASHER_URL")),
		secret:   []byte(getenv("JWT_SECRET", "dev-secret-ganti-di-produksi")),
		juryCode: getenv("JURY_CODE", "JURI2026"),
	}
	if dataFile := getenv("DATA_FILE", "edunexus.json"); dataFile != "off" {
		if err := a.store.Load(dataFile); err != nil {
			log.Printf("gagal memuat data: %v", err)
		} else if n := len(a.store.users); n > 0 {
			log.Printf("data dimuat dari %s (%d akun)", dataFile, n)
		}
		a.dirty = make(chan struct{}, 1)
		go func() {
			for range a.dirty {
				time.Sleep(400 * time.Millisecond)
				if err := a.store.Save(dataFile); err != nil {
					log.Printf("gagal menyimpan data: %v", err)
				}
			}
		}()
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
	mux.HandleFunc("POST /api/quizzes/{id}/start", a.auth(siswa(a.startQuiz)))
	mux.HandleFunc("POST /api/quizzes/{id}/attempt", a.auth(siswa(a.attempt)))
	mux.HandleFunc("GET /api/quizzes/{id}/stats", a.auth(guru(a.quizStats)))

	mux.HandleFunc("GET /api/classes/{id}/posts", a.auth(a.listPosts))
	mux.HandleFunc("POST /api/classes/{id}/posts", a.auth(a.createPost))
	mux.HandleFunc("POST /api/posts/{id}/like", a.auth(a.like))
	mux.HandleFunc("DELETE /api/posts/{id}", a.auth(a.deletePost))

	mux.HandleFunc("GET /api/projects", a.auth(a.listProjects))
	mux.HandleFunc("GET /api/projects/export", a.auth(guruJuri(a.exportProjects)))
	mux.HandleFunc("POST /api/projects", a.auth(siswa(a.createProject)))
	mux.HandleFunc("DELETE /api/projects/{id}", a.auth(a.only("siswa", "guru")(a.deleteProject)))
	mux.HandleFunc("POST /api/projects/{id}/score", a.auth(juri(a.scoreProject)))
	mux.HandleFunc("GET /api/audit", a.auth(guruJuri(a.audit)))

	mux.HandleFunc("GET /healthz", a.health)
	static := http.FileServer(http.Dir(getenv("WEB_DIR", "../web")))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache") // selalu cek versi terbaru
		static.ServeHTTP(w, r)
	}))

	addr := ":" + getenv("PORT", "8080")
	log.Printf("EduNexus gateway (Go) berjalan di %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.wrap(mux)))
}
