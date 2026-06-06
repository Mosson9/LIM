package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mosson9/lim/backend/internal/models"
)

// PostgresStore is a Postgres-backed implementation of Store.
//
// Entities are stored as JSONB documents (reusing the models' json tags) with a
// few promoted columns for indexed lookups; password hashes live in their own
// column since User omits them from JSON. Friendly ids (U-/D-/T-/P-) come from a
// counters table so they match the file store's format. It satisfies the same
// Store interface, so the rest of the app is unchanged.
type PostgresStore struct {
	db *sql.DB
}

// OpenPostgres connects to Postgres (url = a libpq DSN or postgres:// URL) and
// ensures the schema exists.
func OpenPostgres(url string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &PostgresStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *PostgresStore) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS counters (name TEXT PRIMARY KEY, v BIGINT NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS kv (k TEXT PRIMARY KEY, doc JSONB NOT NULL);
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL DEFAULT '', last_active TIMESTAMPTZ NOT NULL, doc JSONB NOT NULL);
CREATE TABLE IF NOT EXISTS decisions (
  id TEXT PRIMARY KEY, user_id TEXT NOT NULL, status TEXT NOT NULL,
  verdict TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL, doc JSONB NOT NULL);
CREATE INDEX IF NOT EXISTS decisions_user_idx ON decisions(user_id);
CREATE TABLE IF NOT EXISTS wishlist (
  id TEXT PRIMARY KEY, user_id TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, doc JSONB NOT NULL);
CREATE INDEX IF NOT EXISTS wishlist_user_idx ON wishlist(user_id);
CREATE TABLE IF NOT EXISTS transactions (
  id TEXT PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL, doc JSONB NOT NULL);
CREATE TABLE IF NOT EXISTS pushes (
  id TEXT PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL, doc JSONB NOT NULL);
`)
	return err
}

// --- small helpers ---

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }

func (s *PostgresStore) nextSeq(name string) (int, error) {
	var v int
	err := s.db.QueryRow(
		`INSERT INTO counters(name, v) VALUES($1, 1)
		 ON CONFLICT (name) DO UPDATE SET v = counters.v + 1 RETURNING v`, name).Scan(&v)
	return v, err
}

func (s *PostgresStore) kvGet(key string, dst any) (bool, error) {
	var raw []byte
	err := s.db.QueryRow(`SELECT doc FROM kv WHERE k=$1`, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(raw, dst)
}

func (s *PostgresStore) kvSet(key string, val any) error {
	_, err := s.db.Exec(
		`INSERT INTO kv(k, doc) VALUES($1, $2)
		 ON CONFLICT (k) DO UPDATE SET doc = EXCLUDED.doc`, key, mustJSON(val))
	return err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// --- lifecycle ---

func (s *PostgresStore) IsEmpty() bool {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return false
	}
	return n == 0
}

// Save is a no-op: Postgres writes are durable on commit.
func (s *PostgresStore) Save() error { return nil }

// --- users ---

func (s *PostgresStore) CreateUser(u *models.User) error {
	if u.ID == "" {
		n, err := s.nextSeq("user")
		if err != nil {
			return err
		}
		u.ID = fmt.Sprintf("U-%05d", 20000+n)
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.LastActiveAt.IsZero() {
		u.LastActiveAt = u.CreatedAt
	}
	_, err := s.db.Exec(
		`INSERT INTO users(id, email, password_hash, last_active, doc) VALUES($1,$2,$3,$4,$5)`,
		u.ID, strings.ToLower(u.Email), u.PasswordHash, u.LastActiveAt, mustJSON(u))
	if isUniqueViolation(err) {
		return ErrDuplicate
	}
	return err
}

// scanUser reads a user row (doc + promoted password_hash).
func scanUser(row interface{ Scan(...any) error }) (*models.User, error) {
	var raw []byte
	var hash string
	if err := row.Scan(&raw, &hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var u models.User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, err
	}
	u.PasswordHash = hash
	return &u, nil
}

func (s *PostgresStore) GetUser(id string) (*models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT doc, password_hash FROM users WHERE id=$1`, id))
}

func (s *PostgresStore) GetUserByEmail(email string) (*models.User, error) {
	return scanUser(s.db.QueryRow(`SELECT doc, password_hash FROM users WHERE lower(email)=lower($1)`, email))
}

func (s *PostgresStore) UpdateUser(u *models.User) error {
	res, err := s.db.Exec(
		`UPDATE users SET email=$2, password_hash=$3, last_active=$4, doc=$5 WHERE id=$1`,
		u.ID, strings.ToLower(u.Email), u.PasswordHash, u.LastActiveAt, mustJSON(u))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListUsers() []*models.User {
	rows, err := s.db.Query(`SELECT doc, password_hash FROM users ORDER BY last_active DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*models.User
	for rows.Next() {
		if u, err := scanUser(rows); err == nil {
			out = append(out, u)
		}
	}
	return out
}

func (s *PostgresStore) TouchUser(id string) {
	now := time.Now()
	_, _ = s.db.Exec(
		`UPDATE users SET last_active=$2,
		   doc = jsonb_set(doc, '{last_active_at}', to_jsonb($2::timestamptz)) WHERE id=$1`,
		id, now)
}

// --- decisions ---

func (s *PostgresStore) CreateDecision(d *models.Decision) error {
	if d.ID == "" {
		n, err := s.nextSeq("decision")
		if err != nil {
			return err
		}
		d.ID = fmt.Sprintf("D-%05d", 88000+n)
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	_, err := s.db.Exec(
		`INSERT INTO decisions(id, user_id, status, verdict, created_at, doc) VALUES($1,$2,$3,$4,$5,$6)`,
		d.ID, d.UserID, string(d.Status), string(d.Verdict), d.CreatedAt, mustJSON(d))
	return err
}

func scanDecision(row interface{ Scan(...any) error }) (*models.Decision, error) {
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var d models.Decision
	return &d, json.Unmarshal(raw, &d)
}

func (s *PostgresStore) GetDecision(id string) (*models.Decision, error) {
	return scanDecision(s.db.QueryRow(`SELECT doc FROM decisions WHERE id=$1`, id))
}

func (s *PostgresStore) UpdateDecision(d *models.Decision) error {
	res, err := s.db.Exec(
		`UPDATE decisions SET user_id=$2, status=$3, verdict=$4, created_at=$5, doc=$6 WHERE id=$1`,
		d.ID, d.UserID, string(d.Status), string(d.Verdict), d.CreatedAt, mustJSON(d))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListDecisions(userID, filter string) []*models.Decision {
	q := `SELECT doc FROM decisions WHERE user_id=$1`
	args := []any{userID}
	switch filter {
	case "", "all":
		q += ` AND status <> 'pending'`
	case "resist":
		q += ` AND status = 'resisted'`
	case "buy":
		q += ` AND status = 'bought'`
	case "pending":
		q += ` AND status = 'pending'`
	default:
		q += ` AND verdict = $2`
		args = append(args, filter)
	}
	q += ` ORDER BY created_at DESC`
	return s.queryDecisions(q, args...)
}

func (s *PostgresStore) ListAllDecisions(filter string) []*models.Decision {
	q := `SELECT doc FROM decisions`
	args := []any{}
	if filter != "" && filter != "all" {
		q += ` WHERE verdict = $1`
		args = append(args, filter)
	}
	q += ` ORDER BY created_at DESC`
	return s.queryDecisions(q, args...)
}

func (s *PostgresStore) queryDecisions(q string, args ...any) []*models.Decision {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*models.Decision
	for rows.Next() {
		if d, err := scanDecision(rows); err == nil {
			out = append(out, d)
		}
	}
	return out
}

// --- wishlist ---

func (s *PostgresStore) AddWishlist(w *models.WishlistItem) error {
	if w.ID == "" {
		w.ID = uuid.NewString()
	}
	_, err := s.db.Exec(
		`INSERT INTO wishlist(id, user_id, expires_at, doc) VALUES($1,$2,$3,$4)`,
		w.ID, w.UserID, w.ExpiresAt, mustJSON(w))
	return err
}

func (s *PostgresStore) GetWishlistItem(id string) (*models.WishlistItem, error) {
	var raw []byte
	err := s.db.QueryRow(`SELECT doc FROM wishlist WHERE id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var w models.WishlistItem
	return &w, json.Unmarshal(raw, &w)
}

func (s *PostgresStore) ListWishlist(userID string) []*models.WishlistItem {
	rows, err := s.db.Query(`SELECT doc FROM wishlist WHERE user_id=$1 ORDER BY expires_at ASC`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*models.WishlistItem
	for rows.Next() {
		var raw []byte
		if rows.Scan(&raw) == nil {
			var w models.WishlistItem
			if json.Unmarshal(raw, &w) == nil {
				out = append(out, &w)
			}
		}
	}
	return out
}

func (s *PostgresStore) DeleteWishlist(id string) error {
	res, err := s.db.Exec(`DELETE FROM wishlist WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- transactions ---

func (s *PostgresStore) AddTransaction(t *models.Transaction) error {
	if t.ID == "" {
		n, err := s.nextSeq("txn")
		if err != nil {
			return err
		}
		t.ID = fmt.Sprintf("T-%04d", 5500+n)
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	_, err := s.db.Exec(`INSERT INTO transactions(id, created_at, doc) VALUES($1,$2,$3)`,
		t.ID, t.CreatedAt, mustJSON(t))
	return err
}

func (s *PostgresStore) ListTransactions() []*models.Transaction {
	rows, err := s.db.Query(`SELECT doc FROM transactions ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*models.Transaction
	for rows.Next() {
		var raw []byte
		if rows.Scan(&raw) == nil {
			var t models.Transaction
			if json.Unmarshal(raw, &t) == nil {
				out = append(out, &t)
			}
		}
	}
	return out
}

// --- config & content (kv-backed) ---

func (s *PostgresStore) AIConfig() models.AIConfig {
	var c models.AIConfig
	_, _ = s.kvGet("ai_config", &c)
	return c
}

func (s *PostgresStore) SetAIConfig(c models.AIConfig) error {
	c.UpdatedAt = time.Now()
	return s.kvSet("ai_config", c)
}

func (s *PostgresStore) Categories() []models.Category {
	var v []models.Category
	_, _ = s.kvGet("categories", &v)
	return v
}
func (s *PostgresStore) SetCategories(v []models.Category) error { return s.kvSet("categories", v) }

func (s *PostgresStore) Skins() []models.Skin {
	var v []models.Skin
	_, _ = s.kvGet("skins", &v)
	return v
}
func (s *PostgresStore) SetSkins(v []models.Skin) error { return s.kvSet("skins", v) }

func (s *PostgresStore) Plans() []models.PlanOption {
	var v []models.PlanOption
	_, _ = s.kvGet("plans", &v)
	return v
}
func (s *PostgresStore) SetPlans(v []models.PlanOption) error { return s.kvSet("plans", v) }

func (s *PostgresStore) Perks() []models.PlusPerk {
	var v []models.PlusPerk
	_, _ = s.kvGet("perks", &v)
	return v
}
func (s *PostgresStore) SetPerks(v []models.PlusPerk) error { return s.kvSet("perks", v) }

// --- pushes ---

func (s *PostgresStore) Pushes() []*models.Push {
	rows, err := s.db.Query(`SELECT doc FROM pushes ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*models.Push
	for rows.Next() {
		var raw []byte
		if rows.Scan(&raw) == nil {
			var p models.Push
			if json.Unmarshal(raw, &p) == nil {
				out = append(out, &p)
			}
		}
	}
	return out
}

func (s *PostgresStore) CreatePush(p *models.Push) error {
	if p.ID == "" {
		n, err := s.nextSeq("push")
		if err != nil {
			return err
		}
		p.ID = fmt.Sprintf("P-%03d", 200+n)
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	_, err := s.db.Exec(`INSERT INTO pushes(id, created_at, doc) VALUES($1,$2,$3)`,
		p.ID, p.CreatedAt, mustJSON(p))
	return err
}
