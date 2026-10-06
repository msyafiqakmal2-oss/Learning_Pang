package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type snapAttempt struct{ U, Q, Score, Total int }

type snapshot struct {
	Seq      int
	Users    []*User
	Classes  []*Class
	Quizzes  []*Quiz
	Attempts []snapAttempt
	Posts    []*Post
	Projects []*Project
}

// Save menulis seluruh data ke satu berkas JSON (atomik: tulis sementara lalu rename).
func (s *Store) Save(path string) error {
	s.mu.Lock()
	snap := snapshot{Seq: s.seq}
	for _, u := range s.users {
		snap.Users = append(snap.Users, u)
	}
	for _, c := range s.classes {
		snap.Classes = append(snap.Classes, c)
	}
	for _, q := range s.quizzes {
		snap.Quizzes = append(snap.Quizzes, q)
	}
	for k, a := range s.attempts {
		snap.Attempts = append(snap.Attempts, snapAttempt{k[0], k[1], a.Score, a.Total})
	}
	for _, p := range s.posts {
		snap.Posts = append(snap.Posts, p)
	}
	for _, p := range s.projects {
		snap.Projects = append(snap.Projects, p)
	}
	data, err := json.Marshal(snap)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load membaca data tersimpan (tidak ada berkas = mulai kosong).
func (s *Store) Load(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq = snap.Seq
	for _, u := range snap.Users {
		if u.Badges == nil {
			u.Badges = []string{}
		}
		s.users[u.ID] = u
	}
	for _, c := range snap.Classes {
		if c.Members == nil {
			c.Members = map[int]bool{}
		}
		s.classes[c.ID] = c
	}
	for _, q := range snap.Quizzes {
		s.quizzes[q.ID] = q
	}
	for _, a := range snap.Attempts {
		s.attempts[[2]int{a.U, a.Q}] = &Attempt{a.Score, a.Total}
	}
	for _, p := range snap.Posts {
		if p.Likes == nil {
			p.Likes = map[int]bool{}
		}
		s.posts[p.ID] = p
	}
	for _, p := range snap.Projects {
		if p.Scores == nil {
			p.Scores = map[int]Score{}
		}
		s.projects[p.ID] = p
	}
	return nil
}
