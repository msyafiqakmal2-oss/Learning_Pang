package main

import (
	"math"
	"sort"
	"strings"
)

// Kriteria penilaian juri (bobot total 100).
var criteria = []struct {
	ID, Name, Hint string
	W              int
}{
	{"inovasi", "Inovasi", "Seberapa baru dan kreatif idenya?", 25},
	{"dampak", "Dampak Pendidikan", "Seberapa besar manfaat belajar bagi siswa?", 25},
	{"kolaborasi", "Kolaborasi", "Seberapa kuat kerja sama tim dan kelas?", 20},
	{"desain", "Desain & UX", "Seberapa menarik dan mudah dipakai?", 15},
	{"teknis", "Kualitas Teknis", "Seberapa kokoh arsitektur dan kodenya?", 15},
}

func weighted(vals map[string]int) float64 {
	t := 0.0
	for _, c := range criteria {
		t += float64(vals[c.ID]*c.W) / 10
	}
	return t
}

func round1(x float64) float64 { return math.Round(x*10) / 10 }

func (s *Store) CreateProject(uid int, title, desc, team string) (*Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.users[uid]
	if u == nil || u.Role != "siswa" {
		return nil, ErrForbidden
	}
	title, desc, team = strings.TrimSpace(title), strings.TrimSpace(desc), strings.TrimSpace(team)
	if len(title) < 3 || len(title) > 80 {
		return nil, bad("Judul karya harus 3-80 karakter.")
	}
	if len(desc) > 400 || len(team) > 60 {
		return nil, bad("Deskripsi maks. 400 karakter dan nama tim maks. 60 karakter.")
	}
	n := 0
	for _, p := range s.projects {
		if p.OwnerID == uid {
			n++
		}
	}
	if n >= 3 {
		return nil, bad("Setiap siswa maksimal mengirim 3 karya.")
	}
	if team == "" {
		team = "Tim " + u.Username
	}
	p := &Project{ID: s.id(), OwnerID: uid, Owner: u.Username, Title: title, Desc: desc, Team: team, Scores: map[int]Score{}}
	s.projects[p.ID] = p
	before := len(u.Badges)
	u.XP += 30
	u.give("pencipta")
	u.touch()
	_ = before
	return p, nil
}

func (s *Store) DeleteProject(uid int, role string, id int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.projects[id]
	if p == nil {
		return "", ErrNotFound
	}
	if p.OwnerID != uid && role != "guru" {
		return "", ErrForbidden
	}
	delete(s.projects, id)
	return p.Title, nil
}

func (s *Store) ScoreProject(jid, pid int, vals map[string]int, comment string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, p := s.users[jid], s.projects[pid]
	if j == nil || p == nil {
		return 0, ErrNotFound
	}
	if j.Role != "juri" {
		return 0, ErrForbidden
	}
	clean := map[string]int{}
	for _, c := range criteria {
		v, ok := vals[c.ID]
		if !ok || v < 1 || v > 10 {
			return 0, bad("Semua kriteria harus dinilai dengan angka 1-10.")
		}
		clean[c.ID] = v
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 300 {
		return 0, bad("Komentar maks. 300 karakter.")
	}
	p.Scores[jid] = Score{Juror: j.Username, Vals: clean, Comment: comment}
	return round1(weighted(clean)), nil
}

// ListProjects mengembalikan karya terurut peringkat (skor tertinggi dulu).
func (s *Store) ListProjects(uid int, role string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	type row struct {
		p     *Project
		avg   float64
		votes int
	}
	rows := []row{}
	for _, p := range s.projects {
		r := row{p: p, votes: len(p.Scores)}
		for _, sc := range p.Scores {
			r.avg += weighted(sc.Vals)
		}
		if r.votes > 0 {
			r.avg /= float64(r.votes)
		}
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if (a.votes > 0) != (b.votes > 0) {
			return a.votes > 0
		}
		if a.avg != b.avg {
			return a.avg > b.avg
		}
		return a.p.ID < b.p.ID
	})
	out := []map[string]any{}
	rank := 0
	for _, r := range rows {
		crit := map[string]float64{}
		comments := []map[string]string{}
		for _, c := range criteria {
			sum := 0
			for _, sc := range r.p.Scores {
				sum += sc.Vals[c.ID]
			}
			if r.votes > 0 {
				crit[c.ID] = round1(float64(sum) / float64(r.votes))
			}
		}
		for _, sc := range r.p.Scores {
			if sc.Comment != "" {
				comments = append(comments, map[string]string{"juror": sc.Juror, "text": sc.Comment})
			}
		}
		v := map[string]any{"id": r.p.ID, "title": r.p.Title, "desc": r.p.Desc, "team": r.p.Team, "owner": r.p.Owner,
			"votes": r.votes, "score": round1(r.avg), "criteria": crit, "comments": comments,
			"can_delete": r.p.OwnerID == uid || role == "guru", "rank": 0}
		if r.votes > 0 {
			rank++
			v["rank"] = rank
		}
		if sc, ok := r.p.Scores[uid]; ok {
			v["mine"] = map[string]any{"vals": sc.Vals, "comment": sc.Comment, "total": round1(weighted(sc.Vals))}
		}
		out = append(out, v)
	}
	cs := []map[string]any{}
	for _, c := range criteria {
		cs = append(cs, map[string]any{"id": c.ID, "name": c.Name, "hint": c.Hint, "weight": c.W})
	}
	return map[string]any{"criteria": cs, "projects": out}
}
