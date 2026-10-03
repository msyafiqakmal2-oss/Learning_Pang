const $ = (id) => document.getElementById(id);
const S = { me: null, tab: "home", cls: null, clsTab: "quiz", lb: "siswa" };
const NAV = {
  siswa: [["home", "🏠 Beranda"], ["kelas", "📚 Kelas"], ["peringkat", "🏆 Peringkat"], ["karya", "🎨 Karya"], ["panduan", "🧭 Panduan"]],
  guru: [["home", "🏠 Beranda"], ["kelas", "📚 Kelas Saya"], ["peringkat", "🏆 Peringkat"], ["karya", "🎨 Karya"], ["panduan", "🧭 Panduan"]],
  juri: [["home", "🏠 Beranda"], ["karya", "⚖️ Penilaian"], ["peringkat", "🏆 Peringkat"], ["panduan", "🧭 Panduan Juri"]],
};

// ---------- util ----------
function h(tag, props, ...kids) {
  const n = document.createElement(tag);
  for (const [k, v] of Object.entries(props || {})) {
    if (v == null || v === false) continue;
    if (k === "class") n.className = v;
    else if (k.startsWith("on")) n.addEventListener(k.slice(2), v);
    else if (["value", "checked", "disabled", "selected"].includes(k)) n[k] = v;
    else n.setAttribute(k, v === true ? "" : v);
  }
  kids.flat(9).forEach((c) => c != null && c !== false && n.append(c.nodeType ? c : document.createTextNode(c)));
  return n;
}

async function api(path, opt = {}) {
  const token = localStorage.getItem("token");
  if (opt.body && typeof opt.body !== "string") opt.body = JSON.stringify(opt.body);
  const res = await fetch("/api" + path, {
    ...opt,
    headers: { "Content-Type": "application/json", ...(token && { Authorization: "Bearer " + token }) },
  });
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 401 && token && path !== "/auth/login") logout();
    throw new Error(data.error || `Server membalas ${res.status}. Pastikan halaman dibuka lewat http://localhost:8080.`);
  }
  return data;
}

let toastT;
function toast(t) {
  const el = $("toast");
  el.textContent = t;
  el.classList.add("show");
  clearTimeout(toastT);
  toastT = setTimeout(() => el.classList.remove("show"), 3200);
}

function reward(res) {
  const parts = [];
  if (res.xp) parts.push(`+${res.xp} XP`);
  if (res.bonus) parts.push(`🎯 Misi harian selesai +${res.bonus} XP`);
  (res.badges || []).forEach((b) => parts.push(`Lencana baru: ${b}`));
  if (res.profile) { S.me = res.profile; header(); }
  if (parts.length) toast(parts.join("  •  "));
}

function modal(node) {
  const m = $("modal");
  const close = () => { m.hidden = true; m.replaceChildren(); };
  m.replaceChildren(h("div", { class: "sheet", role: "dialog", "aria-modal": "true" }, h("button", { class: "x", "aria-label": "Tutup", onclick: close }, "✕"), node));
  m.hidden = false;
  m.onclick = (e) => { if (e.target === m) close(); };
  return close;
}

function confetti() {
  const colors = ["#6B4EFF", "#FFC931", "#1FBF8F", "#FF6B5B", "#4CC3FF"];
  for (let i = 0; i < 36; i++) {
    const c = h("i", { class: "cf", style: `left:${Math.random() * 100}%;background:${colors[i % 5]};animation-delay:${Math.random() * 0.6}s` });
    document.body.append(c);
    setTimeout(() => c.remove(), 3200);
  }
}

const put = (node, ...kids) => node.replaceChildren(...kids.flat(9).filter((k) => k != null && k !== false));
const bar = (pct, cls = "") => h("div", { class: "bar " + cls }, h("i", { style: `width:${Math.max(0, Math.min(100, pct))}%` }));
const empty = (t) => h("div", { class: "empty" }, t);
const confirmDel = (t) => confirm(t);
const run = (fn) => async (e) => { try { await fn(e); } catch (err) { toast("⚠️ " + err.message); } };

// ---------- auth ----------
let mode = "login";
const TEXT = {
  login: { title: "Selamat datang kembali", sub: "Masuk untuk melanjutkan petualangan belajarmu.", user: "Nama pengguna", ph: "Nama pengguna Anda", btn: "Masuk" },
  register: { title: "Buat akun baru", sub: "Gabung gratis, pilih peranmu, dan mulai kumpulkan XP.", user: "Nama panggilan", ph: "3-32 karakter", btn: "Buat akun" },
};
const roleVal = () => document.querySelector('input[name="role"]:checked').value;

function setMode(m) {
  mode = m;
  const t = TEXT[m], reg = m === "register";
  $("tab-login").classList.toggle("on", !reg);
  $("tab-register").classList.toggle("on", reg);
  $("auth-title").textContent = t.title;
  $("auth-sub").textContent = t.sub;
  $("lbl-user").textContent = t.user;
  $("username").placeholder = t.ph;
  $("auth-submit").textContent = t.btn;
  $("role-wrap").hidden = !reg;
  $("confirm-wrap").hidden = !reg;
  $("password2").required = reg;
  syncCode();
  $("password").autocomplete = reg ? "new-password" : "current-password";
  $("auth-msg").textContent = "";
}
function syncCode() {
  const need = mode === "register" && roleVal() === "juri";
  $("code-wrap").hidden = !need;
  $("code").required = need;
}
$("tab-login").onclick = () => setMode("login");
$("tab-register").onclick = () => setMode("register");
document.querySelectorAll('input[name="role"]').forEach((r) => (r.onchange = syncCode));

async function doAuth(m, username, password) {
  const body = { username, password };
  if (m === "register") { body.role = roleVal(); body.code = $("code").value; }
  const d = await api("/auth/" + m, { method: "POST", body });
  localStorage.setItem("token", d.token);
  start(true);
}

$("auth-form").onsubmit = async (e) => {
  e.preventDefault();
  if (mode === "register" && $("password").value !== $("password2").value) {
    $("auth-msg").textContent = "Kata sandi dan ulangannya tidak sama.";
    return;
  }
  try { await doAuth(mode, $("username").value, $("password").value); }
  catch (err) { $("auth-msg").textContent = err.message; }
};
document.querySelectorAll("[data-demo]").forEach((b) => (b.onclick = async () => {
  try { await doAuth("login", b.dataset.demo, "demo12345"); }
  catch (err) { $("auth-msg").textContent = err.message + " (data demo butuh beberapa detik setelah server menyala)"; }
}));

function logout() {
  localStorage.removeItem("token");
  S.me = null; S.tab = "home"; S.cls = null;
  $("modal").hidden = true;
  $("app").hidden = true;
  $("auth").hidden = false;
}
$("logout").onclick = logout;

// ---------- kerangka ----------
function header() {
  const m = S.me;
  $("chip").textContent = m.role === "siswa" ? `${m.username} · Lv ${m.level} · ${m.xp} XP` : `${m.username} · ${m.role}`;
}
function nav() {
  $("nav").replaceChildren(...NAV[S.me.role].map(([id, label]) =>
    h("button", { class: S.tab === id ? "on" : "", onclick: () => go(id) }, label)));
}
function go(tab) { S.tab = tab; S.cls = null; S.clsTab = "quiz"; render(); window.scrollTo(0, 0); }

async function render() {
  header(); nav();
  const view = $("view");
  try { view.replaceChildren(await VIEWS[S.tab]()); }
  catch (err) { view.replaceChildren(h("div", { class: "card" }, "⚠️ " + err.message)); }
}

async function start(fresh) {
  try { S.me = await api("/me"); } catch { return logout(); }
  $("auth").hidden = true;
  $("app").hidden = false;
  S.tab = "home"; S.cls = null;
  await render();
  if (fresh) toast(`Halo, ${S.me.username}! ${S.me.role === "juri" ? "Buka menu Penilaian untuk menilai karya." : "Selamat belajar dan kumpulkan XP!"}`);
}

// ---------- beranda ----------
async function home() {
  const m = S.me, st = await api("/stats");
  const wrap = h("div", { style: "display:grid;gap:22px" });
  const mk = (n, l) => h("div", { class: "stat" }, h("div", { class: "num" }, n), h("span", {}, l));
  if (m.role !== "juri") {
    wrap.append(h("div", { class: "card profile" },
      h("div", { class: "avatar" }, m.username[0].toUpperCase()),
      h("div", { style: "flex:1;min-width:220px" },
        h("div", { class: "row" }, h("h2", { style: "margin:0" }, `Halo, ${m.username}!`), h("span", { class: "tag" }, m.role)),
        m.role === "siswa" ? [
          h("div", { class: "row sb", style: "margin:10px 0 6px" }, h("span", { class: "big" }, `Level ${m.level}`), h("span", { class: "muted" }, `${m.level_xp}/100 XP menuju Level ${m.level + 1}`)),
          h("div", { class: "xpbar" }, h("i", { style: `width:${m.level_xp}%` })),
        ] : h("p", { class: "muted" }, "Buat kelas, susun kuis, dan pantau semangat belajar siswa."),
      ),
      m.role === "siswa" && h("div", { class: "stat" }, h("div", { class: "num" }, `🔥 ${m.streak}`), h("span", {}, "hari berturut-turut")),
    ));
  } else {
    wrap.append(h("div", { class: "card" }, h("h2", {}, "⚖️ Selamat datang, Juri!"),
      h("ol", {}, h("li", {}, "Buka menu ", h("b", {}, "Penilaian"), " lalu pilih karya."), h("li", {}, "Nilai 5 kriteria (1-10), tulis komentar singkat."), h("li", {}, "Peringkat dan Calon Juara 1 diperbarui otomatis.")),
      h("button", { class: "btn primary", onclick: () => go("karya") }, "Mulai menilai →")));
  }
  if (m.role === "siswa") {
    wrap.append(h("div", { class: "card" }, h("h2", {}, "🎯 Misi Harian"),
      m.quests.map((q) => h("div", { class: "quest" + (q.done ? " done" : "") },
        h("div", { class: "ck" }, q.done ? "✓" : ""),
        h("div", {}, h("b", {}, q.title), bar((q.progress / q.goal) * 100, "mint")),
        h("span", { class: "tag sun" }, `+${q.xp} XP`)))));
    wrap.append(h("div", { class: "card" }, h("h2", {}, "🏅 Koleksi Lencana"),
      h("div", { class: "badges" }, m.catalog.map((b) => h("div", { class: "badge" + (m.badges.includes(b.id) ? "" : " lock") },
        h("div", { class: "ic" }, b.icon), h("b", {}, b.name), h("small", {}, b.desc))))));
  }
  wrap.append(h("div", { class: "card" }, h("h2", {}, "📈 Dampak EduNexus saat ini"),
    h("div", { class: "stats" }, mk(st.students, "siswa aktif"), mk(st.classes, "kelas"), mk(st.attempts, "kuis dikerjakan"),
      mk(st.posts, "diskusi"), mk(st.projects, "karya proyek"), mk(st.votes, "penilaian juri"), mk(st.xp, "total XP terkumpul"))));
  return wrap;
}

// ---------- kelas ----------
async function kelas() {
  if (S.cls) return classDetail();
  const m = S.me, list = await api("/classes");
  const wrap = h("div", { style: "display:grid;gap:22px" });
  if (m.role === "siswa") {
    const code = h("input", { placeholder: "Contoh: MTK7A", maxlength: 8, "aria-label": "Kode kelas" });
    wrap.append(h("div", { class: "card" }, h("h2", {}, "🔑 Gabung kelas"),
      h("div", { class: "row" }, h("div", { style: "flex:1;min-width:180px" }, code),
        h("button", { class: "btn primary", style: "margin-top:5px", onclick: run(async () => {
          const r = await api("/classes/join", { method: "POST", body: { code: code.value } });
          toast(`Berhasil bergabung ke ${r.name} 🎉`); render();
        }) }, "Gabung")),
      h("p", { class: "muted", style: "margin:8px 0 0;font-size:.85rem" }, "Minta kode kelas dari gurumu. Kode demo: MTK7A atau ENG8B.")));
  }
  if (m.role === "guru") {
    const name = h("input", { placeholder: "Nama kelas, mis. IPA Kelas 8C", maxlength: 60, "aria-label": "Nama kelas" });
    wrap.append(h("div", { class: "card" }, h("h2", {}, "➕ Buat kelas baru"),
      h("div", { class: "row" }, h("div", { style: "flex:1;min-width:200px" }, name),
        h("button", { class: "btn primary", style: "margin-top:5px", onclick: run(async () => {
          const r = await api("/classes", { method: "POST", body: { name: name.value } });
          toast(`Kelas dibuat! Bagikan kode: ${r.code}`); render();
        }) }, "Buat kelas"))));
  }
  wrap.append(list.length ? h("div", { class: "grid" }, list.map((c) => h("div", { class: "card cls", onclick: () => { S.cls = c; S.clsTab = "quiz"; render(); } },
    h("h3", {}, c.name), h("p", { class: "muted", style: "margin:4px 0 10px" }, `Guru: ${c.teacher} · ${c.members} siswa`),
    h("div", { class: "row sb" }, h("span", { class: "tag mint" }, `⚡ ${c.total_xp} XP kelas`), c.code && h("span", { class: "tag sun" }, `Kode ${c.code}`)),
    m.role === "guru" && h("button", { class: "btn danger sm", style: "margin-top:12px", onclick: run(async (e) => {
      e.stopPropagation();
      if (!confirmDel(`Hapus kelas "${c.name}" beserta kuis dan diskusinya?`)) return;
      await api("/classes/" + c.id, { method: "DELETE" }); toast("Kelas dihapus"); render();
    }) }, "Hapus kelas")))) : empty(m.role === "siswa" ? "Belum bergabung ke kelas mana pun. Masukkan kode kelas di atas." : "Belum ada kelas."));
  return wrap;
}

async function classDetail() {
  const c = S.cls, m = S.me;
  const wrap = h("div", { style: "display:grid;gap:18px" });
  wrap.append(h("div", { class: "row sb" },
    h("div", {}, h("button", { class: "btn sm", onclick: () => { S.cls = null; render(); } }, "← Semua kelas"), h("h2", { style: "margin-top:12px;font-size:1.8rem" }, c.name)),
    h("div", { class: "sub-tabs" }, [["quiz", "📝 Kuis"], ["post", "💬 Diskusi"]].map(([id, l]) => h("button", { class: S.clsTab === id ? "on" : "", onclick: () => { S.clsTab = id; render(); } }, l)))));
  wrap.append(S.clsTab === "quiz" ? await quizPanel(c, m) : await postPanel(c, m));
  return wrap;
}

async function quizPanel(c, m) {
  const qs = await api(`/classes/${c.id}/quizzes`);
  const card = h("div", { class: "card" });
  card.append(h("div", { class: "row sb" }, h("h2", { style: "margin:0" }, "Kuis kelas"),
    m.role === "guru" && h("button", { class: "btn primary sm", onclick: () => quizBuilder(c) }, "+ Buat kuis")));
  if (!qs.length) card.append(empty("Belum ada kuis."));
  qs.forEach((q) => card.append(h("div", { class: "quiz" },
    h("div", {}, h("b", {}, q.title), h("div", { class: "muted", style: "font-size:.88rem" }, `${q.questions.length} soal · hadiah hingga ${q.xp_reward + 10} XP`,
      q.done && h("span", { class: "tag mint", style: "margin-left:8px" }, `Terbaik ${q.best}/${q.total}`))),
    h("div", { class: "row" },
      m.role === "siswa" && h("button", { class: "btn " + (q.done ? "" : "mint"), onclick: () => playQuiz(q) }, q.done ? "Ulangi" : "Mulai ▶"),
      m.role === "guru" && h("button", { class: "btn danger sm", onclick: run(async () => {
        if (!confirmDel(`Hapus kuis "${q.title}"?`)) return;
        await api("/quizzes/" + q.id, { method: "DELETE" }); toast("Kuis dihapus"); render();
      }) }, "Hapus")))));
  return card;
}

function playQuiz(q) {
  let i = 0; const ans = [];
  const box = h("div");
  const close = modal(box);
  const step = () => {
    const qu = q.questions[i];
    const next = h("button", { class: "btn primary", disabled: true, onclick: run(async () => {
      if (i < q.questions.length - 1) { i++; step(); return; }
      const res = await api(`/quizzes/${q.id}/attempt`, { method: "POST", body: { answers: ans } });
      reward(res); showResult(res);
    }) }, i < q.questions.length - 1 ? "Lanjut →" : "Selesai ✓");
    put(box, 
      h("span", { class: "tag" }, q.title), h("h3", { style: "margin-top:10px" }, `Soal ${i + 1} dari ${q.questions.length}`),
      h("div", { class: "progress" }, h("i", { style: `width:${(i / q.questions.length) * 100}%` })),
      h("p", { style: "font-size:1.1rem;font-weight:700" }, qu.text),
      qu.options.map((o, k) => h("button", { class: "opt" + (ans[i] === k ? " sel" : ""), onclick: (e) => {
        ans[i] = k; next.disabled = false;
        box.querySelectorAll(".opt").forEach((b) => b.classList.remove("sel")); e.currentTarget.classList.add("sel");
      } }, `${String.fromCharCode(65 + k)}. ${o}`)),
      h("div", { class: "row", style: "justify-content:flex-end;margin-top:10px" }, next));
    if (ans[i] != null) { box.querySelectorAll(".opt")[ans[i]].classList.add("sel"); next.disabled = false; }
  };
  const showResult = (res) => {
    if (res.score === res.total) confetti();
    put(box, h("div", { class: "result" },
      h("div", { class: "big" }, `${res.score}/${res.total}`),
      h("h3", {}, res.score === res.total ? "Sempurna! 🎉" : res.score >= res.total / 2 ? "Bagus sekali! 👏" : "Ayo coba lagi! 💪"),
      h("p", { class: "muted" }, res.first ? `Kamu mendapat +${res.xp} XP${res.bonus ? ` dan bonus misi +${res.bonus} XP` : ""}` : "XP hanya diberikan pada percobaan pertama."),
      res.badges.length ? h("p", {}, res.badges.map((b) => h("span", { class: "tag sun", style: "margin:2px" }, b))) : null),
      h("h3", { style: "margin:14px 0 6px" }, "Pembahasan"),
      q.questions.map((qu, k) => h("div", { style: "margin-bottom:12px" }, h("b", {}, `${k + 1}. ${qu.text}`),
        qu.options.map((o, j) => h("div", { class: "opt " + (j === res.correct[k] ? "ok" : j === ans[k] ? "no" : ""), style: "margin:4px 0;padding:6px 12px" },
          o, j === res.correct[k] ? " ✓" : j === ans[k] ? " ✗" : "")))),
      h("div", { class: "row", style: "justify-content:flex-end" }, h("button", { class: "btn primary", onclick: () => { close(); render(); } }, "Tutup")));
  };
  step();
}

function quizBuilder(c) {
  const title = h("input", { placeholder: "Judul kuis", maxlength: 80 });
  const xp = h("select", {}, [30, 50, 80, 100].map((v) => h("option", { value: v, selected: v === 50 }, `${v} XP`)));
  const blocks = h("div");
  let n = 0;
  const add = () => {
    const id = "q" + n++;
    blocks.append(h("div", { class: "qblock", "data-q": id },
      h("label", {}, "Pertanyaan", h("input", { class: "qt", placeholder: "Tulis pertanyaan" })),
      [0, 1, 2, 3].map((k) => h("div", { class: "optrow" },
        h("input", { type: "radio", name: id, value: k, checked: k === 0, "aria-label": `Jawaban benar ${k + 1}` }),
        h("input", { type: "text", class: "qo", placeholder: `Pilihan ${String.fromCharCode(65 + k)}${k > 1 ? " (opsional)" : ""}` }))),
      h("small", { class: "muted" }, "Pilih tombol bulat di samping jawaban yang benar.")));
  };
  add();
  modal(h("div", {}, h("h2", {}, "Buat kuis baru"),
    h("label", {}, "Judul", title), h("label", {}, "Hadiah XP", xp), blocks,
    h("div", { class: "row sb" }, h("button", { class: "btn sm", onclick: add }, "+ Tambah soal"),
      h("button", { class: "btn primary", onclick: run(async () => {
        const questions = [...blocks.children].map((b) => {
          const opts = [...b.querySelectorAll(".qo")].map((x) => x.value.trim());
          const correct = +b.querySelector("input[type=radio]:checked").value;
          if (!opts[correct]) throw new Error("Jawaban benar tidak boleh kosong.");
          const keep = opts.map((o, i) => [o, i]).filter(([o]) => o);
          return { text: b.querySelector(".qt").value, options: keep.map(([o]) => o), answer: keep.findIndex(([, i]) => i === correct) };
        });
        await api(`/classes/${c.id}/quizzes`, { method: "POST", body: { title: title.value, xp_reward: +xp.value, questions } });
        $("modal").hidden = true; toast("Kuis dibuat! 🎉"); render();
      }) }, "Simpan kuis"))));
}

async function postPanel(c, m) {
  const posts = await api(`/classes/${c.id}/posts`);
  const wrap = h("div", { style: "display:grid;gap:18px" });
  if (m.role !== "juri") {
    const t = h("textarea", { placeholder: "Bagikan tips, pertanyaan, atau ajakan belajar bareng…", maxlength: 500, "aria-label": "Tulisan baru" });
    wrap.append(h("div", { class: "card" }, t, h("div", { class: "row sb", style: "margin-top:8px" },
      h("span", { class: "muted", style: "font-size:.85rem" }, "+5 XP untuk setiap diskusi, +2 XP setiap kali tulisanmu disukai"),
      h("button", { class: "btn primary", onclick: run(async () => {
        const r = await api(`/classes/${c.id}/posts`, { method: "POST", body: { text: t.value } });
        reward(r); render();
      }) }, "Kirim"))));
  }
  const card = h("div", { class: "card" }, h("h2", {}, "Diskusi kelas"));
  if (!posts.length) card.append(empty("Belum ada diskusi. Jadilah yang pertama!"));
  posts.forEach((p) => card.append(h("div", { class: "post" },
    h("div", { class: "row" }, h("b", {}, p.author), h("span", { class: "tag" }, p.role), h("span", { class: "muted", style: "font-size:.8rem" }, new Date(p.at).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" }))),
    h("p", {}, p.text),
    h("div", { class: "row" },
      h("button", { class: "like" + (p.liked ? " on" : ""), disabled: p.liked || p.mine, onclick: run(async () => {
        const r = await api(`/posts/${p.id}/like`, { method: "POST" }); reward(r); render();
      }) }, `♥ ${p.likes}`),
      p.can_delete && h("button", { class: "btn danger sm", onclick: run(async () => {
        if (!confirmDel("Hapus tulisan ini?")) return;
        await api("/posts/" + p.id, { method: "DELETE" }); toast("Tulisan dihapus"); render();
      }) }, "Hapus")))));
  wrap.append(card);
  return wrap;
}

// ---------- peringkat ----------
async function peringkat() {
  const lb = await api("/leaderboard"), m = S.me;
  const wrap = h("div", { class: "card" });
  wrap.append(h("div", { class: "row sb" }, h("h2", { style: "margin:0" }, "🏆 Papan Peringkat"),
    h("div", { class: "sub-tabs" }, [["siswa", "Siswa"], ["kelas", "Tantangan Kelas"]].map(([id, l]) => h("button", { class: S.lb === id ? "on" : "", onclick: () => { S.lb = id; render(); } }, l)))));
  const medal = (r) => (r === 1 ? "🥇" : r === 2 ? "🥈" : r === 3 ? "🥉" : r);
  const list = h("div", { style: "margin-top:14px;display:grid;gap:8px" });
  if (S.lb === "siswa") {
    const top = lb.students[0]?.xp || 1;
    lb.students.forEach((s) => list.append(h("div", { class: "rankrow" + (s.id === m.id ? " me" : "") },
      h("div", { class: "pos" }, medal(s.rank)),
      h("div", {}, h("b", {}, s.username, s.id === m.id ? " (kamu)" : ""), h("div", { class: "muted", style: "font-size:.8rem" }, `Level ${s.level} · 🔥 ${s.streak} hari · 🏅 ${s.badges}`), bar((s.xp / top) * 100)),
      h("div", { class: "num" }, `${s.xp} XP`))));
  } else {
    const top = lb.classes[0]?.xp || 1;
    lb.classes.forEach((c) => list.append(h("div", { class: "rankrow" },
      h("div", { class: "pos" }, medal(c.rank)),
      h("div", {}, h("b", {}, c.name), h("div", { class: "muted", style: "font-size:.8rem" }, `${c.teacher} · ${c.members} siswa`), bar((c.xp / top) * 100, "mint")),
      h("div", { class: "num" }, `${c.xp} XP`))));
    list.append(h("p", { class: "muted", style: "font-size:.85rem" }, "Total XP seluruh anggota kelas. Belajar bersama membuat kelasmu naik peringkat!"));
  }
  if (!list.children.length) list.append(empty("Belum ada data."));
  wrap.append(list);
  return wrap;
}

// ---------- karya & juri ----------
async function karya() {
  const m = S.me, d = await api("/projects"), ps = d.projects, crit = d.criteria;
  const wrap = h("div", { style: "display:grid;gap:22px" });
  const scored = ps.filter((p) => p.rank > 0).slice(0, 3);
  if (scored.length) {
    wrap.append(h("div", { class: "card" }, h("h2", {}, "👑 Podium Juara Karya"),
      h("div", { class: "podium" }, scored.map((p) => h("div", { class: "pod p" + p.rank },
        h("div", { class: "crown" }, p.rank === 1 ? "👑" : p.rank === 2 ? "🥈" : "🥉"),
        p.rank === 1 && h("span", { class: "tag" }, "Calon Juara 1"),
        h("h3", {}, p.title), h("div", { class: "muted", style: "font-size:.8rem" }, p.team),
        h("div", { class: "num" }, p.score), h("small", {}, `dari ${p.votes} juri`))))));
  }
  if (m.role === "siswa") {
    const t = h("input", { placeholder: "Judul karya / proyek", maxlength: 80 });
    const team = h("input", { placeholder: "Nama tim (opsional)", maxlength: 60 });
    const desc = h("textarea", { placeholder: "Ceritakan idemu dan manfaatnya bagi pembelajaran", maxlength: 400 });
    wrap.append(h("div", { class: "card" }, h("h2", {}, "🎨 Kirim karya proyekmu"),
      h("div", { class: "grid", style: "grid-template-columns:repeat(auto-fit,minmax(220px,1fr))" }, h("label", {}, "Judul", t), h("label", {}, "Tim", team)),
      h("label", {}, "Deskripsi", desc),
      h("button", { class: "btn primary", onclick: run(async () => {
        await api("/projects", { method: "POST", body: { title: t.value, team: team.value, desc: desc.value } });
        toast("Karya terkirim! +30 XP & lencana Pencipta Karya 🎨"); S.me = await api("/me"); render();
      }) }, "Kirim karya")));
  }
  if (m.role === "juri") {
    wrap.append(h("div", { class: "card" }, h("h2", {}, "⚖️ Panduan penilaian"),
      h("p", { class: "muted", style: "margin-top:0" }, "Nilai tiap kriteria 1-10. Skor akhir (0-100) dihitung dengan bobot berikut:"),
      h("div", { class: "row" }, crit.map((c) => h("span", { class: "tag sun" }, `${c.name} ${c.weight}%`)))));
  }
  if (!ps.length) wrap.append(empty("Belum ada karya."));
  ps.forEach((p) => wrap.append(projectCard(p, crit, m)));
  return wrap;
}

function projectCard(p, crit, m) {
  const panel = h("div", { hidden: true });
  const card = h("div", { class: "card" });
  const head = h("div", { class: "proj" },
    h("div", { class: "rk r" + p.rank }, p.rank || "–"),
    h("div", {}, h("h3", {}, p.title, p.rank === 1 && " 👑"), h("div", { class: "muted", style: "font-size:.85rem" }, `${p.team} · oleh ${p.owner}`),
      p.desc && h("p", { style: "margin:8px 0" }, p.desc)),
    h("div", { class: "score" }, p.votes ? [h("div", { class: "num" }, p.score), h("small", { class: "muted" }, `/100 · ${p.votes} juri`)] : h("small", { class: "muted" }, "Belum dinilai")));
  card.append(head);
  if (p.votes) card.append(h("div", { style: "margin-top:8px" }, crit.map((c) => h("div", { class: "crit" }, h("span", {}, c.name), bar((p.criteria[c.id] || 0) * 10), h("span", {}, p.criteria[c.id])))));
  p.comments.forEach((c) => card.append(h("div", { class: "quote" }, `💬 ${c.juror}: ${c.text}`)));
  const actions = h("div", { class: "row", style: "margin-top:12px" });
  if (m.role === "juri") {
    actions.append(h("button", { class: "btn " + (p.mine ? "" : "sun"), onclick: () => { panel.hidden = !panel.hidden; } }, p.mine ? `✏️ Ubah nilaiku (${p.mine.total})` : "⚖️ Beri nilai"));
    const vals = {}; let totalEl;
    const sum = () => crit.reduce((t, c) => t + (vals[c.id] * c.weight) / 10, 0);
    panel.append(...crit.map((c) => {
      vals[c.id] = p.mine ? p.mine.vals[c.id] : 5;
      const out = h("b", {}, vals[c.id]);
      return h("div", { class: "slider" }, h("div", { class: "row" }, h("span", {}, h("b", {}, c.name), ` (${c.weight}%) · `, h("span", { class: "muted" }, c.hint)), out),
        h("input", { type: "range", min: 1, max: 10, value: vals[c.id], "aria-label": c.name, oninput: (e) => { vals[c.id] = +e.target.value; out.textContent = vals[c.id]; totalEl.textContent = sum().toFixed(1); } }));
    }));
    const note = h("textarea", { placeholder: "Komentar singkat untuk tim (opsional)", maxlength: 300, value: p.mine?.comment || "" });
    totalEl = h("span", { class: "num", style: "font-size:1.6rem;color:var(--violet)" }, sum().toFixed(1));
    panel.append(note, h("div", { class: "row sb", style: "margin-top:10px" }, h("span", {}, "Skor akhirmu: ", totalEl, " / 100"),
      h("button", { class: "btn primary", onclick: run(async () => {
        const r = await api(`/projects/${p.id}/score`, { method: "POST", body: { scores: vals, comment: note.value } });
        toast(`Nilai tersimpan: ${r.total}/100 ✅`); render();
      }) }, "Simpan penilaian")));
  }
  if (p.can_delete) actions.append(h("button", { class: "btn danger sm", onclick: run(async () => {
    if (!confirmDel(`Hapus karya "${p.title}"?`)) return;
    await api("/projects/" + p.id, { method: "DELETE" }); toast("Karya dihapus"); render();
  }) }, "Hapus"));
  card.append(actions, panel);
  return card;
}

// ---------- panduan ----------
async function panduan() {
  const m = S.me, wrap = h("div", { style: "display:grid;gap:22px" });
  wrap.append(h("div", { class: "card" }, h("h2", {}, "🧭 Panduan demo 3 menit"),
    h("ol", {},
      h("li", {}, h("b", {}, "Siswa"), ": gabung kelas, kerjakan kuis, lihat XP/level/streak naik, selesaikan misi harian."),
      h("li", {}, h("b", {}, "Kolaborasi"), ": tulis diskusi dan beri like, lalu lihat Tantangan Kelas di Peringkat."),
      h("li", {}, h("b", {}, "Guru"), ": buat kelas dan kuis, hapus konten, pantau kelas."),
      h("li", {}, h("b", {}, "Juri"), ": menu Penilaian → beri nilai 5 kriteria → lihat Podium Juara otomatis."))));
  wrap.append(h("div", { class: "card" }, h("h2", {}, "🔑 Akun demo"),
    h("table", {}, h("tr", {}, h("th", {}, "Peran"), h("th", {}, "Nama pengguna"), h("th", {}, "Kata sandi")),
      [["Guru", "bu_sari"], ["Siswa", "rina"], ["Siswa", "putri"], ["Siswa", "dimas"], ["Juri", "juri1"], ["Juri", "juri2"]].map(([r, u]) => h("tr", {}, h("td", {}, r), h("td", {}, h("code", {}, u)), h("td", {}, h("code", {}, "demo12345"))))),
    h("p", { class: "muted", style: "margin-bottom:0" }, "Mendaftar sebagai juri baru? Gunakan kode ", h("code", {}, "JURI2026"), ".")));
  const map = [
    ["Inovasi (25%)", "Gamifikasi lengkap: XP, level, streak, misi harian, lencana; Tantangan Kelas; arsitektur multi-bahasa Go + Rust."],
    ["Dampak Pendidikan (25%)", "Kuis dengan pembahasan, XP hanya di percobaan pertama (anti-curang), guru dapat membuat kuis sendiri."],
    ["Kolaborasi (20%)", "Diskusi kelas dengan apresiasi, XP gabungan kelas, karya tim yang dinilai juri."],
    ["Desain & UX (15%)", "Antarmuka ramah, responsif, mendukung keyboard, kontras tinggi, animasi dapat dimatikan (reduced motion)."],
    ["Kualitas Teknis (15%)", "Go (API, JWT, RBAC 3 peran), Rust (Argon2id + audit log), Docker Compose, validasi server, XSS-safe."],
  ];
  wrap.append(h("div", { class: "card" }, h("h2", {}, "🗺️ Peta fitur → kriteria juri"), h("div", { class: "grid" }, map.map(([t, d]) => h("div", { class: "stat", style: "text-align:left" }, h("b", {}, t), h("p", { class: "muted", style: "margin:6px 0 0;font-size:.88rem" }, d))))));
  if (m.role === "guru" || m.role === "juri") {
    const box = h("div", { class: "log" }, "Memuat…");
    wrap.append(h("div", { class: "card" }, h("h2", {}, "🔍 Log audit transparan ", h("small", { class: "muted", style: "font-weight:400;font-size:.8rem" }, "(dari layanan Rust)")), box));
    api("/audit").then((log) => box.replaceChildren(...(log.length ? log.map((a) => h("div", {}, `${new Date(a.ts * 1000).toLocaleTimeString("id-ID")}  ${a.actor}  ${a.action}  ${a.detail}`)) : [h("div", {}, "Belum ada aktivitas (atau layanan Rust belum berjalan).")])))
      .catch((e) => box.replaceChildren(h("div", {}, e.message)));
  }
  return wrap;
}

const VIEWS = { home, kelas, peringkat, karya, panduan };
setMode("login");
if (localStorage.getItem("token")) start(false);
