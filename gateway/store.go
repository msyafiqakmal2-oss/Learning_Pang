package main

import (
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------- error ----------
type apiErr struct {
	code int
	msg  string
}

func (e apiErr) Error() string { return e.msg }
func bad(msg string) error     { return apiErr{400, msg} }

var (
	ErrExists    = apiErr{409, "Data sudah ada."}
	ErrNotFound  = apiErr{404, "Data tidak ditemukan."}
	ErrForbidden = apiErr{403, "Anda tidak punya izin untuk aksi ini."}
)

// ---------- model ----------
type dayStats struct {
	Date  string
	Count map[string]int
	Done  map[string]bool
}

type User struct {
	ID       int
	Username string
	Role     string // siswa | guru | juri
	Hash     string
	XP       int
	Streak   int
	LastDay  string
	Badges   []string
	Day      dayStats
}

type Class struct {
	ID          int
	Name, Code  string
	TeacherID   int
	TeacherName string
	Members     map[int]bool
}

type Question struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

type Quiz struct {
	ID, ClassID, XPReward int
	Title                 string
	Questions             []Question
	Takers, ScoreSum      int
	Correct               []int
}

type Attempt struct{ Score, Total int }

type Post struct {
	ID, ClassID, AuthorID int
	Author, Role, Text    string
	Likes                 map[int]bool
	At                    time.Time
}

type Score struct {
	Juror   string
	Vals    map[string]int
	Comment string
}

type Project struct {
	ID, OwnerID  int
	Owner, Title string
	Desc, Team   string
	Scores       map[int]Score
}

type Store struct {
	mu       sync.Mutex
	seq      int
	users    map[int]*User
	classes  map[int]*Class
	quizzes  map[int]*Quiz
	attempts map[[2]int]*Attempt
	posts    map[int]*Post
	projects map[int]*Project
	started  map[[2]int]time.Time
}

func NewStore() *Store {
	return &Store{users: map[int]*User{}, classes: map[int]*Class{}, quizzes: map[int]*Quiz{},
		attempts: map[[2]int]*Attempt{}, posts: map[int]*Post{}, projects: map[int]*Project{}, started: map[[2]int]time.Time{}}
}

func (s *Store) id() int { s.seq++; return s.seq }

// ---------- gamifikasi ----------
var quests = []struct {
	ID, Title string
	Goal, XP  int
}{
	{"quiz", "Kerjakan 1 kuis", 1, 25},
	{"post", "Tulis 1 diskusi di kelas", 1, 15},
	{"like", "Beri 2 apresiasi (like)", 2, 10},
}

var badgeCatalog = []map[string]string{
	{"id": "pemula", "name": "Langkah Pertama", "desc": "Menyelesaikan kuis pertama", "icon": "🚀"},
	{"id": "sempurna", "name": "Nilai Sempurna", "desc": "Menjawab semua soal dengan benar", "icon": "🎯"},
	{"id": "kolaborator", "name": "Kolaborator", "desc": "Menulis diskusi pertama", "icon": "🤝"},
	{"id": "streak3", "name": "Api Semangat", "desc": "Belajar 3 hari berturut-turut", "icon": "🔥"},
	{"id": "disukai", "name": "Bintang Kelas", "desc": "Satu postingan mendapat 5 like", "icon": "⭐"},
	{"id": "pencipta", "name": "Pencipta Karya", "desc": "Mengirim karya proyek", "icon": "🎨"},
	{"id": "kilat", "name": "Kilat", "desc": "Nilai sempurna dengan jawaban super cepat", "icon": "⚡"},
	{"id": "streak7", "name": "Konsisten", "desc": "Belajar 7 hari berturut-turut", "icon": "🌟"},
}

func levelTitle(l int) string {
	t := []string{"Pemula", "Penjelajah", "Petualang", "Pendekar", "Ahli", "Master"}
	if l >= 1 && l <= len(t) {
		return t[l-1]
	}
	return "Legenda"
}

func today() string { return time.Now().Format("2006-01-02") }

func (u *User) give(id string) {
	for _, b := range u.Badges {
		if b == id {
			return
		}
	}
	u.Badges = append(u.Badges, id)
}

func (u *User) rollDay() {
	if u.Day.Date != today() || u.Day.Count == nil {
		u.Day = dayStats{Date: today(), Count: map[string]int{}, Done: map[string]bool{}}
	}
}

// touch memperbarui streak harian.
func (u *User) touch() {
	t := today()
	if u.LastDay == t {
		return
	}
	if u.LastDay == time.Now().AddDate(0, 0, -1).Format("2006-01-02") {
		u.Streak++
	} else {
		u.Streak = 1
	}
	u.LastDay = t
	if u.Streak >= 3 {
		u.give("streak3")
	}
	if u.Streak >= 7 {
		u.give("streak7")
	}
}

// event mencatat progres misi harian dan mengembalikan bonus XP bila misi selesai.
func (u *User) event(kind string) int {
	u.rollDay()
	u.Day.Count[kind]++
	for _, q := range quests {
		if q.ID == kind && u.Day.Count[kind] >= q.Goal && !u.Day.Done[kind] {
			u.Day.Done[kind] = true
			u.XP += q.XP
			return q.XP
		}
	}
	return 0
}

func newBadges(u *User, before int) []string {
	out := []string{}
	for _, id := range u.Badges[before:] {
		for _, b := range badgeCatalog {
			if b["id"] == id {
				out = append(out, b["icon"]+" "+b["name"])
			}
		}
	}
	return out
}

func (s *Store) profile(u *User) map[string]any {
	u.rollDay()
	qs := []map[string]any{}
	for _, q := range quests {
		p := u.Day.Count[q.ID]
		if p > q.Goal {
			p = q.Goal
		}
		qs = append(qs, map[string]any{"id": q.ID, "title": q.Title, "goal": q.Goal, "progress": p, "xp": q.XP, "done": u.Day.Done[q.ID]})
	}
	return map[string]any{"id": u.ID, "username": u.Username, "role": u.Role, "xp": u.XP,
		"level": u.XP/100 + 1, "title": levelTitle(u.XP/100 + 1), "level_xp": u.XP % 100, "streak": u.Streak,
		"badges": u.Badges, "catalog": badgeCatalog, "quests": qs}
}

func (s *Store) Profile(id int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.users[id]
	if u == nil {
		return nil, ErrNotFound
	}
	return s.profile(u), nil
}

// ---------- user ----------
func (s *Store) CreateUser(name, role, hash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if strings.EqualFold(u.Username, name) {
			return nil, ErrExists
		}
	}
	u := &User{ID: s.id(), Username: name, Role: role, Hash: hash, Badges: []string{}}
	s.users[u.ID] = u
	return u, nil
}

func (s *Store) UserByName(name string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if strings.EqualFold(u.Username, name) {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

// ---------- kelas ----------
func (s *Store) canSee(c *Class, uid int, role string) bool {
	return c != nil && (role == "juri" || c.TeacherID == uid || c.Members[uid])
}

func (s *Store) classXP(c *Class) int {
	t := 0
	for m := range c.Members {
		if u := s.users[m]; u != nil {
			t += u.XP
		}
	}
	return t
}

func (s *Store) ListClasses(uid int, role string) []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []map[string]any{}
	ids := []int{}
	for id := range s.classes {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		c := s.classes[id]
		if !s.canSee(c, uid, role) {
			continue
		}
		v := map[string]any{"id": c.ID, "name": c.Name, "teacher": c.TeacherName, "members": len(c.Members), "total_xp": s.classXP(c)}
		if c.TeacherID == uid || role == "juri" {
			v["code"] = c.Code
		}
		out = append(out, v)
	}
	return out
}

func (s *Store) CreateClass(uid int, name string) (*Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = strings.TrimSpace(name)
	if len(name) < 3 || len(name) > 60 {
		return nil, bad("Nama kelas harus 3-60 karakter.")
	}
	u := s.users[uid]
	const abc = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := ""
	for {
		b := make([]byte, 5)
		for i := range b {
			b[i] = abc[rand.Intn(len(abc))]
		}
		code = string(b)
		dup := false
		for _, c := range s.classes {
			dup = dup || c.Code == code
		}
		if !dup {
			break
		}
	}
	c := &Class{ID: s.id(), Name: name, Code: code, TeacherID: uid, TeacherName: u.Username, Members: map[int]bool{}}
	s.classes[c.ID] = c
	return c, nil
}

func (s *Store) JoinClass(uid int, code string) (*Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, c := range s.classes {
		if c.Code == code {
			if c.Members[uid] {
				return nil, ErrExists
			}
			c.Members[uid] = true
			return c, nil
		}
	}
	return nil, apiErr{404, "Kode kelas tidak ditemukan."}
}

func (s *Store) DeleteClass(uid, id int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.classes[id]
	if c == nil {
		return "", ErrNotFound
	}
	if c.TeacherID != uid {
		return "", ErrForbidden
	}
	for qid, q := range s.quizzes {
		if q.ClassID == id {
			delete(s.quizzes, qid)
			for k := range s.attempts {
				if k[1] == qid {
					delete(s.attempts, k)
				}
			}
		}
	}
	for pid, p := range s.posts {
		if p.ClassID == id {
			delete(s.posts, pid)
		}
	}
	delete(s.classes, id)
	return c.Name, nil
}

// ---------- kuis ----------
func (s *Store) CreateQuiz(uid, classID int, title string, reward int, qs []Question) (*Quiz, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.classes[classID]
	if c == nil {
		return nil, ErrNotFound
	}
	if c.TeacherID != uid {
		return nil, ErrForbidden
	}
	title = strings.TrimSpace(title)
	if len(title) < 3 || len(title) > 80 {
		return nil, bad("Judul kuis harus 3-80 karakter.")
	}
	if len(qs) < 1 || len(qs) > 10 {
		return nil, bad("Kuis harus berisi 1-10 soal.")
	}
	for _, q := range qs {
		if strings.TrimSpace(q.Text) == "" || len(q.Options) < 2 || len(q.Options) > 5 || q.Answer < 0 || q.Answer >= len(q.Options) {
			return nil, bad("Setiap soal butuh pertanyaan, 2-5 pilihan, dan satu jawaban benar.")
		}
		for _, o := range q.Options {
			if strings.TrimSpace(o) == "" {
				return nil, bad("Pilihan jawaban tidak boleh kosong.")
			}
		}
	}
	if reward < 10 || reward > 200 {
		reward = 50
	}
	q := &Quiz{ID: s.id(), ClassID: classID, Title: title, XPReward: reward, Questions: qs, Correct: make([]int, len(qs))}
	s.quizzes[q.ID] = q
	return q, nil
}

func (s *Store) ListQuizzes(uid int, role string, classID int) ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canSee(s.classes[classID], uid, role) {
		return nil, ErrForbidden
	}
	ids := []int{}
	for id, q := range s.quizzes {
		if q.ClassID == classID {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	out := []map[string]any{}
	for _, id := range ids {
		q := s.quizzes[id]
		view := []map[string]any{}
		for _, qu := range q.Questions {
			view = append(view, map[string]any{"text": qu.Text, "options": qu.Options}) // tanpa kunci jawaban
		}
		v := map[string]any{"id": q.ID, "title": q.Title, "xp_reward": q.XPReward, "questions": view, "done": false}
		if a := s.attempts[[2]int{uid, id}]; a != nil {
			v["done"], v["best"], v["total"] = true, a.Score, a.Total
		}
		out = append(out, v)
	}
	return out, nil
}

func (s *Store) DeleteQuiz(uid, id int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.quizzes[id]
	if q == nil {
		return "", ErrNotFound
	}
	if s.classes[q.ClassID].TeacherID != uid {
		return "", ErrForbidden
	}
	delete(s.quizzes, id)
	for k := range s.attempts {
		if k[1] == id {
			delete(s.attempts, k)
		}
	}
	return q.Title, nil
}

// canPlay memeriksa siswa anggota kelas pemilik kuis.
func (s *Store) canPlay(uid int, q *Quiz) (*User, error) {
	u := s.users[uid]
	if u == nil || q == nil {
		return nil, ErrNotFound
	}
	c := s.classes[q.ClassID]
	if u.Role != "siswa" || c == nil || !c.Members[uid] {
		return nil, ErrForbidden
	}
	return u, nil
}

// StartQuiz mencatat waktu mulai di server (dasar bonus kecepatan).
func (s *Store) StartQuiz(uid, qid int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.quizzes[qid]
	if _, err := s.canPlay(uid, q); err != nil {
		return 0, err
	}
	s.started[[2]int{uid, qid}] = time.Now()
	return 20 * len(q.Questions), nil
}

func (s *Store) SubmitAttempt(uid, qid int, ans []int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.quizzes[qid]
	u, err := s.canPlay(uid, q)
	if err != nil {
		return nil, err
	}
	total := len(q.Questions)
	if len(ans) != total {
		return nil, bad("Jawab semua soal terlebih dahulu.")
	}
	score := 0
	correct := make([]int, total)
	for i, qu := range q.Questions {
		correct[i] = qu.Answer
		if ans[i] == qu.Answer {
			score++
		}
	}
	key := [2]int{uid, qid}
	secs := 0
	if t0, ok := s.started[key]; ok {
		secs = int(time.Since(t0).Seconds())
		delete(s.started, key)
	}
	before := len(u.Badges)
	xp, speed := 0, 0
	prev := s.attempts[key]
	first := prev == nil
	if first {
		xp = int(math.Round(float64(q.XPReward) * float64(score) / float64(total)))
		if score == total {
			xp += 10
			u.give("sempurna")
			if limit := 20 * total; secs > 0 && secs < limit { // bonus kecepatan maks. +10 XP
				speed = int(math.Round(10 * (1 - float64(secs)/float64(limit))))
			}
			if speed >= 6 {
				u.give("kilat")
			}
			xp += speed
		}
		u.give("pemula")
		u.XP += xp
		s.attempts[key] = &Attempt{score, total}
		if len(q.Correct) != total {
			q.Correct = make([]int, total)
		}
		q.Takers++
		q.ScoreSum += score
		for i := range correct {
			if ans[i] == correct[i] {
				q.Correct[i]++
			}
		}
	} else if score > prev.Score {
		prev.Score = score
	}
	u.touch()
	bonus := u.event("quiz")
	return map[string]any{"score": score, "total": total, "xp": xp, "bonus": bonus, "speed": speed, "seconds": secs,
		"correct": correct, "first": first, "badges": newBadges(u, before), "profile": s.profile(u)}, nil
}

// QuizStats: analitik untuk guru pemilik kelas.
func (s *Store) QuizStats(uid, qid int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := s.quizzes[qid]
	if q == nil {
		return nil, ErrNotFound
	}
	c := s.classes[q.ClassID]
	if c == nil || c.TeacherID != uid {
		return nil, ErrForbidden
	}
	pct := func(a, b int) int {
		if b == 0 {
			return 0
		}
		return int(math.Round(100 * float64(a) / float64(b)))
	}
	qs := []map[string]any{}
	for i, qu := range q.Questions {
		ok := 0
		if i < len(q.Correct) {
			ok = q.Correct[i]
		}
		qs = append(qs, map[string]any{"text": qu.Text, "correct_pct": pct(ok, q.Takers)})
	}
	res := []map[string]any{}
	for k, a := range s.attempts {
		if k[1] == qid && s.users[k[0]] != nil {
			res = append(res, map[string]any{"username": s.users[k[0]].Username, "score": a.Score, "total": a.Total})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i]["score"].(int) != res[j]["score"].(int) {
			return res[i]["score"].(int) > res[j]["score"].(int)
		}
		return res[i]["username"].(string) < res[j]["username"].(string)
	})
	return map[string]any{"title": q.Title, "takers": q.Takers, "members": len(c.Members),
		"avg_pct": pct(q.ScoreSum, q.Takers*len(q.Questions)), "questions": qs, "results": res}, nil
}

// ---------- diskusi ----------
func (s *Store) ListPosts(uid int, role string, classID int) ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.classes[classID]
	if !s.canSee(c, uid, role) {
		return nil, ErrForbidden
	}
	ps := []*Post{}
	for _, p := range s.posts {
		if p.ClassID == classID {
			ps = append(ps, p)
		}
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].ID > ps[j].ID })
	out := []map[string]any{}
	for _, p := range ps {
		out = append(out, map[string]any{"id": p.ID, "author": p.Author, "role": p.Role, "text": p.Text,
			"likes": len(p.Likes), "liked": p.Likes[uid], "mine": p.AuthorID == uid, "at": p.At,
			"can_delete": p.AuthorID == uid || c.TeacherID == uid})
	}
	return out, nil
}

func (s *Store) CreatePost(uid, classID int, text string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, c := s.users[uid], s.classes[classID]
	if u == nil || u.Role == "juri" || !s.canSee(c, uid, u.Role) {
		return nil, ErrForbidden
	}
	text = strings.TrimSpace(text)
	if text == "" || len(text) > 500 {
		return nil, bad("Tulisan wajib diisi (maks. 500 karakter).")
	}
	p := &Post{ID: s.id(), ClassID: classID, AuthorID: uid, Author: u.Username, Role: u.Role, Text: text, Likes: map[int]bool{}, At: time.Now()}
	s.posts[p.ID] = p
	before := len(u.Badges)
	u.XP += 5
	u.give("kolaborator")
	u.touch()
	bonus := u.event("post")
	return map[string]any{"xp": 5, "bonus": bonus, "badges": newBadges(u, before), "profile": s.profile(u)}, nil
}

func (s *Store) Like(uid, pid int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, u := s.posts[pid], s.users[uid]
	if p == nil || u == nil {
		return nil, ErrNotFound
	}
	if !s.canSee(s.classes[p.ClassID], uid, u.Role) {
		return nil, ErrForbidden
	}
	if p.AuthorID == uid {
		return nil, bad("Kamu tidak bisa menyukai postinganmu sendiri.")
	}
	if p.Likes[uid] {
		return nil, ErrExists
	}
	p.Likes[uid] = true
	if a := s.users[p.AuthorID]; a != nil {
		a.XP += 2
		if len(p.Likes) >= 5 {
			a.give("disukai")
		}
	}
	u.touch()
	bonus := u.event("like")
	return map[string]any{"likes": len(p.Likes), "bonus": bonus, "profile": s.profile(u)}, nil
}

func (s *Store) DeletePost(uid, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.posts[id]
	if p == nil {
		return ErrNotFound
	}
	if p.AuthorID != uid && s.classes[p.ClassID].TeacherID != uid {
		return ErrForbidden
	}
	delete(s.posts, id)
	return nil
}

// ---------- peringkat ----------
func (s *Store) Leaderboard() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	us := []*User{}
	for _, u := range s.users {
		if u.Role == "siswa" {
			us = append(us, u)
		}
	}
	sort.Slice(us, func(i, j int) bool {
		if us[i].XP != us[j].XP {
			return us[i].XP > us[j].XP
		}
		return us[i].ID < us[j].ID
	})
	students := []map[string]any{}
	for i, u := range us {
		if i >= 20 {
			break
		}
		students = append(students, map[string]any{"rank": i + 1, "id": u.ID, "username": u.Username, "xp": u.XP,
			"level": u.XP/100 + 1, "title": levelTitle(u.XP/100 + 1), "streak": u.Streak, "badges": len(u.Badges)})
	}
	cs := []*Class{}
	for _, c := range s.classes {
		cs = append(cs, c)
	}
	sort.Slice(cs, func(i, j int) bool { return s.classXP(cs[i]) > s.classXP(cs[j]) })
	classes := []map[string]any{}
	for i, c := range cs {
		classes = append(classes, map[string]any{"rank": i + 1, "name": c.Name, "teacher": c.TeacherName, "members": len(c.Members), "xp": s.classXP(c)})
	}
	return map[string]any{"students": students, "classes": classes}
}

func (s *Store) Stats() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, gu, xp := 0, 0, 0
	for _, u := range s.users {
		if u.Role == "siswa" {
			st++
			xp += u.XP
		} else if u.Role == "guru" {
			gu++
		}
	}
	votes := 0
	for _, p := range s.projects {
		votes += len(p.Scores)
	}
	return map[string]any{"students": st, "teachers": gu, "classes": len(s.classes), "quizzes": len(s.quizzes),
		"attempts": len(s.attempts), "posts": len(s.posts), "projects": len(s.projects), "votes": votes, "xp": xp}
}
