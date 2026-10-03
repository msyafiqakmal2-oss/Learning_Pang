package main

import (
	"log"
	"time"
)

const demoPass = "demo12345"

// seed mengisi data contoh agar juri langsung melihat aplikasi yang "hidup".
func (a *App) seed() {
	var hash string
	var err error
	for i := 0; i < 30; i++ { // tunggu layanan Rust siap
		if hash, err = a.hasher.Hash(demoPass); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Printf("seed dibatalkan: %v", err)
		return
	}
	s := a.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.users) > 0 {
		return
	}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	add := func(name, role string, xp, streak int, badges ...string) *User {
		u := &User{ID: s.id(), Username: name, Role: role, Hash: hash, XP: xp, Streak: streak, LastDay: yesterday, Badges: append([]string{}, badges...)}
		s.users[u.ID] = u
		return u
	}
	guru := add("bu_sari", "guru", 0, 0)
	add("juri1", "juri", 0, 0)
	juri2 := add("juri2", "juri", 0, 0)
	rina := add("rina", "siswa", 340, 4, "pemula", "sempurna", "kolaborator", "streak3")
	dimas := add("dimas", "siswa", 280, 2, "pemula", "kolaborator")
	putri := add("putri", "siswa", 410, 6, "pemula", "sempurna", "kolaborator", "streak3", "disukai")
	bayu := add("bayu", "siswa", 150, 1, "pemula")
	citra := add("citra", "siswa", 220, 3, "pemula", "kolaborator", "streak3")

	mkClass := func(name, code string, members ...*User) *Class {
		c := &Class{ID: s.id(), Name: name, Code: code, TeacherID: guru.ID, TeacherName: guru.Username, Members: map[int]bool{}}
		for _, m := range members {
			c.Members[m.ID] = true
		}
		s.classes[c.ID] = c
		return c
	}
	mtk := mkClass("Matematika Seru 7A", "MTK7A", rina, dimas, putri, bayu, citra)
	eng := mkClass("English Club 8B", "ENG8B", putri, rina, citra)

	mkQuiz := func(c *Class, title string, xp int, qs ...Question) *Quiz {
		q := &Quiz{ID: s.id(), ClassID: c.ID, Title: title, XPReward: xp, Questions: qs}
		s.quizzes[q.ID] = q
		return q
	}
	q1 := mkQuiz(mtk, "Pecahan Dasar", 60,
		Question{"Hasil dari 1/2 + 1/4 adalah...", []string{"1/6", "3/4", "2/6", "1/8"}, 1},
		Question{"Pecahan yang senilai dengan 2/3 adalah...", []string{"4/6", "3/4", "2/6", "4/9"}, 0},
		Question{"Manakah pecahan terbesar?", []string{"1/3", "2/5", "3/4", "1/2"}, 2})
	mkQuiz(mtk, "Bangun Datar", 50,
		Question{"Luas persegi dengan sisi 6 cm adalah...", []string{"12 cm²", "24 cm²", "36 cm²", "18 cm²"}, 2},
		Question{"Banyak sisi pada segitiga adalah...", []string{"2", "3", "4", "5"}, 1},
		Question{"Keliling persegi panjang (p=8, l=3) adalah...", []string{"22", "24", "11", "16"}, 0})
	mkQuiz(eng, "Simple Present", 50,
		Question{"She ___ to school every day.", []string{"go", "goes", "going", "gone"}, 1},
		Question{"They ___ football on Sunday.", []string{"plays", "play", "playing", "played"}, 1},
		Question{"The plural of 'child' is...", []string{"childs", "childes", "children", "childrens"}, 2})
	s.attempts[[2]int{rina.ID, q1.ID}] = &Attempt{3, 3}
	s.attempts[[2]int{dimas.ID, q1.ID}] = &Attempt{2, 3}
	s.attempts[[2]int{putri.ID, q1.ID}] = &Attempt{3, 3}

	post := func(c *Class, u *User, text string, likers ...*User) {
		p := &Post{ID: s.id(), ClassID: c.ID, AuthorID: u.ID, Author: u.Username, Role: u.Role, Text: text, Likes: map[int]bool{}, At: time.Now().Add(-time.Duration(s.seq) * time.Minute)}
		for _, l := range likers {
			p.Likes[l.ID] = true
		}
		s.posts[p.ID] = p
	}
	post(mtk, guru, "Selamat datang di kelas Matematika Seru! Kerjakan kuis Pecahan Dasar untuk mengumpulkan XP pertama kalian. 🚀", rina, dimas, putri)
	post(mtk, rina, "Tips dari aku: ubah dulu penyebutnya jadi sama, baru dijumlahkan. Kalau mau aku bahas bareng, balas ya!", dimas, putri, bayu, citra)
	post(mtk, dimas, "Ada yang mau belajar kelompok untuk bangun datar? Kita bisa bikin kartu rumus bareng.", rina)

	mkProj := func(owner *User, title, desc, team string, v ...int) {
		p := &Project{ID: s.id(), OwnerID: owner.ID, Owner: owner.Username, Title: title, Desc: desc, Team: team, Scores: map[int]Score{}}
		if len(v) == 5 {
			p.Scores[juri2.ID] = Score{Juror: juri2.Username, Comment: "Ide kuat dan relevan dengan kebutuhan sekolah.",
				Vals: map[string]int{"inovasi": v[0], "dampak": v[1], "kolaborasi": v[2], "desain": v[3], "teknis": v[4]}}
		}
		s.projects[p.ID] = p
	}
	mkProj(putri, "EcoLearn: Bank Sampah Digital", "Siswa mengumpulkan poin dari memilah sampah sekolah, lalu menukarnya jadi XP belajar.", "Tim Hijau", 9, 9, 8, 8, 7)
	mkProj(dimas, "Peta Konsep AR", "Kartu pelajaran yang menampilkan peta konsep 3D lewat kamera ponsel.", "Tim Realita", 8, 7, 8, 9, 9)
	mkProj(rina, "Smart Study Planner", "Perencana belajar yang menyusun jadwal berdasarkan hasil kuis dan streak.", "Tim Fokus", 7, 8, 9, 7, 8)
	mkProj(bayu, "Kamus Bahasa Daerah", "Kamus kolaboratif tempat siswa menambah kosakata bahasa daerah dengan audio.", "Tim Nusantara")
	log.Printf("data demo siap (akun: bu_sari, rina, dimas, putri, bayu, citra, juri1, juri2 | sandi: %s)", demoPass)
}
