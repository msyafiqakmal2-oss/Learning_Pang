package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type App struct {
	store    *Store
	hasher   *Hasher
	secret   []byte
	juryCode string
}

func respond(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	respond(w, code, map[string]string{"error": msg})
}

// fx menerjemahkan error domain menjadi respons HTTP.
func fx(w http.ResponseWriter, err error) {
	var ae apiErr
	if errors.As(err, &ae) {
		fail(w, ae.code, ae.msg)
		return
	}
	fail(w, 500, "Terjadi kesalahan pada server.")
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		fail(w, 400, "Format data tidak valid.")
		return false
	}
	return true
}

func pid(r *http.Request) int { n, _ := strconv.Atoi(r.PathValue("id")); return n }

// ---------- auth ----------
type regReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Code     string `json:"code"`
}

func (a *App) session(w http.ResponseWriter, code int, u *User) {
	respond(w, code, map[string]any{"token": IssueToken(&User{ID: u.ID, Username: u.Username, Role: u.Role}, a.secret, 12*time.Hour)})
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var c regReq
	if !decode(w, r, &c) {
		return
	}
	c.Username = strings.TrimSpace(c.Username)
	if len(c.Username) < 3 || len(c.Username) > 32 {
		fail(w, 400, "Nama panggilan harus 3-32 karakter.")
		return
	}
	if len(c.Password) < 8 {
		fail(w, 400, "Kata sandi minimal 8 karakter.")
		return
	}
	if c.Role != "siswa" && c.Role != "guru" && c.Role != "juri" {
		fail(w, 400, "Pilih peran: siswa, guru, atau juri.")
		return
	}
	if c.Role == "juri" && c.Code != a.juryCode {
		fail(w, 403, "Kode juri salah. Lihat kode di panduan panitia.")
		return
	}
	hash, err := a.hasher.Hash(c.Password)
	if err != nil {
		fail(w, 502, "Layanan keamanan (Rust) tidak dapat dihubungi.")
		return
	}
	u, err := a.store.CreateUser(c.Username, c.Role, hash)
	if err != nil {
		fail(w, 409, "Nama panggilan sudah dipakai.")
		return
	}
	a.hasher.Audit(u.Username, "register", "peran="+u.Role)
	a.session(w, 201, u)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var c regReq
	if !decode(w, r, &c) {
		return
	}
	u, err := a.store.UserByName(strings.TrimSpace(c.Username))
	if err != nil || !a.hasher.Verify(c.Password, u.Hash) {
		a.hasher.Audit(c.Username, "login_gagal", r.RemoteAddr)
		fail(w, 401, "Nama pengguna atau kata sandi salah.")
		return
	}
	a.hasher.Audit(u.Username, "login", "peran="+u.Role)
	a.session(w, 200, u)
}

func (a *App) me(w http.ResponseWriter, r *http.Request, c *Claims) {
	p, err := a.store.Profile(c.Sub)
	if err != nil {
		fail(w, 401, "Akun tidak ditemukan. Silakan masuk lagi.")
		return
	}
	respond(w, 200, p)
}

// ---------- kelas & kuis ----------
func (a *App) listClasses(w http.ResponseWriter, r *http.Request, c *Claims) {
	respond(w, 200, a.store.ListClasses(c.Sub, c.Role))
}

func (a *App) createClass(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	cl, err := a.store.CreateClass(c.Sub, in.Name)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "kelas_dibuat", cl.Name)
	respond(w, 201, map[string]any{"id": cl.ID, "code": cl.Code})
}

func (a *App) joinClass(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Code string `json:"code"`
	}
	if !decode(w, r, &in) {
		return
	}
	cl, err := a.store.JoinClass(c.Sub, in.Code)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "gabung_kelas", cl.Name)
	respond(w, 200, map[string]any{"id": cl.ID, "name": cl.Name})
}

func (a *App) deleteClass(w http.ResponseWriter, r *http.Request, c *Claims) {
	name, err := a.store.DeleteClass(c.Sub, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "kelas_dihapus", name)
	w.WriteHeader(204)
}

func (a *App) listQuizzes(w http.ResponseWriter, r *http.Request, c *Claims) {
	qs, err := a.store.ListQuizzes(c.Sub, c.Role, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	respond(w, 200, qs)
}

func (a *App) createQuiz(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Title     string     `json:"title"`
		XPReward  int        `json:"xp_reward"`
		Questions []Question `json:"questions"`
	}
	if !decode(w, r, &in) {
		return
	}
	q, err := a.store.CreateQuiz(c.Sub, pid(r), in.Title, in.XPReward, in.Questions)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "kuis_dibuat", q.Title)
	respond(w, 201, map[string]any{"id": q.ID})
}

func (a *App) deleteQuiz(w http.ResponseWriter, r *http.Request, c *Claims) {
	t, err := a.store.DeleteQuiz(c.Sub, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "kuis_dihapus", t)
	w.WriteHeader(204)
}

func (a *App) attempt(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Answers []int `json:"answers"`
	}
	if !decode(w, r, &in) {
		return
	}
	res, err := a.store.SubmitAttempt(c.Sub, pid(r), in.Answers)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "kuis_dikerjakan", strconv.Itoa(res["score"].(int))+"/"+strconv.Itoa(res["total"].(int)))
	respond(w, 200, res)
}

// ---------- diskusi ----------
func (a *App) listPosts(w http.ResponseWriter, r *http.Request, c *Claims) {
	ps, err := a.store.ListPosts(c.Sub, c.Role, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	respond(w, 200, ps)
}

func (a *App) createPost(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &in) {
		return
	}
	res, err := a.store.CreatePost(c.Sub, pid(r), in.Text)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "diskusi_ditulis", "")
	respond(w, 201, res)
}

func (a *App) like(w http.ResponseWriter, r *http.Request, c *Claims) {
	res, err := a.store.Like(c.Sub, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	respond(w, 200, res)
}

func (a *App) deletePost(w http.ResponseWriter, r *http.Request, c *Claims) {
	if err := a.store.DeletePost(c.Sub, pid(r)); err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "diskusi_dihapus", strconv.Itoa(pid(r)))
	w.WriteHeader(204)
}

// ---------- peringkat, karya, juri ----------
func (a *App) leaderboard(w http.ResponseWriter, r *http.Request, c *Claims) {
	respond(w, 200, a.store.Leaderboard())
}

func (a *App) stats(w http.ResponseWriter, r *http.Request, c *Claims) {
	respond(w, 200, a.store.Stats())
}

func (a *App) listProjects(w http.ResponseWriter, r *http.Request, c *Claims) {
	respond(w, 200, a.store.ListProjects(c.Sub, c.Role))
}

func (a *App) createProject(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct{ Title, Desc, Team string }
	if !decode(w, r, &in) {
		return
	}
	p, err := a.store.CreateProject(c.Sub, in.Title, in.Desc, in.Team)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "karya_dikirim", p.Title)
	respond(w, 201, map[string]any{"id": p.ID})
}

func (a *App) deleteProject(w http.ResponseWriter, r *http.Request, c *Claims) {
	t, err := a.store.DeleteProject(c.Sub, c.Role, pid(r))
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "karya_dihapus", t)
	w.WriteHeader(204)
}

func (a *App) scoreProject(w http.ResponseWriter, r *http.Request, c *Claims) {
	var in struct {
		Scores  map[string]int `json:"scores"`
		Comment string         `json:"comment"`
	}
	if !decode(w, r, &in) {
		return
	}
	total, err := a.store.ScoreProject(c.Sub, pid(r), in.Scores, in.Comment)
	if err != nil {
		fx(w, err)
		return
	}
	a.hasher.Audit(c.Name, "juri_menilai", "karya#"+strconv.Itoa(pid(r))+" skor="+strconv.FormatFloat(total, 'f', 1, 64))
	respond(w, 200, map[string]any{"total": total})
}

func (a *App) audit(w http.ResponseWriter, r *http.Request, c *Claims) {
	raw, err := a.hasher.AuditList()
	if err != nil {
		fail(w, 502, "Log audit tidak tersedia (layanan Rust belum berjalan).")
		return
	}
	respond(w, 200, raw)
}
