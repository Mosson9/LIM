// Package store is the persistence layer for LIM.
//
// The default implementation keeps all state in memory and snapshots it to a
// JSON file on every mutation, which makes the server runnable with zero
// external services while still surviving restarts. The Store type is the only
// thing the HTTP layer depends on, so a Postgres-backed implementation can be
// dropped in later without touching the handlers (see docs/DEPLOYMENT.md).
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mosson9/lim/backend/internal/models"
)

// ErrNotFound is returned when a lookup fails.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a unique constraint (e.g. email) is violated.
var ErrDuplicate = errors.New("duplicate")

// state is the serialisable portion of the store.
type state struct {
	Users        map[string]*models.User         `json:"users"`
	Decisions    map[string]*models.Decision     `json:"decisions"`
	Wishlist     map[string]*models.WishlistItem `json:"wishlist"`
	Transactions map[string]*models.Transaction  `json:"transactions"`
	Pushes       map[string]*models.Push         `json:"pushes"`
	AIConfig     models.AIConfig                 `json:"ai_config"`
	Categories   []models.Category               `json:"categories"`
	Skins        []models.Skin                   `json:"skins"`
	Plans        []models.PlanOption             `json:"plans"`
	Perks        []models.PlusPerk               `json:"perks"`
	DecisionSeq  int                             `json:"decision_seq"`
	TxnSeq       int                             `json:"txn_seq"`
	UserSeq      int                             `json:"user_seq"`
	PushSeq      int                             `json:"push_seq"`
}

// Store is a concurrency-safe, file-backed data store.
type Store struct {
	mu    sync.RWMutex
	path  string
	st    state
	email map[string]string // lower(email) -> userID, rebuilt on load
}

// Open loads the store from path, or returns an empty (unseeded) store if the
// file does not exist. Callers should seed an empty store.
func Open(path string) (*Store, error) {
	s := &Store{path: path, email: map[string]string{}}
	s.st = state{
		Users:        map[string]*models.User{},
		Decisions:    map[string]*models.Decision{},
		Wishlist:     map[string]*models.WishlistItem{},
		Transactions: map[string]*models.Transaction{},
		Pushes:       map[string]*models.Push{},
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.st); err != nil {
		return nil, fmt.Errorf("decode store: %w", err)
	}
	s.reindex()
	return s, nil
}

// IsEmpty reports whether the store has no users (i.e. needs seeding).
func (s *Store) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.st.Users) == 0
}

func (s *Store) reindex() {
	s.email = make(map[string]string, len(s.st.Users))
	for id, u := range s.st.Users {
		s.email[strings.ToLower(u.Email)] = id
	}
}

// persist writes the current state to disk atomically. Caller must hold s.mu.
func (s *Store) persist() error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s.st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Save flushes the store to disk (used by the seeder).
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persist()
}

// --- Users ---

// CreateUser inserts a new user, assigning IDs and a friendly display id.
func (s *Store) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.ToLower(u.Email)
	if _, ok := s.email[key]; ok {
		return ErrDuplicate
	}
	if u.ID == "" {
		s.st.UserSeq++
		u.ID = fmt.Sprintf("U-%05d", 20000+s.st.UserSeq)
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.LastActiveAt.IsZero() {
		u.LastActiveAt = u.CreatedAt
	}
	cp := *u
	s.st.Users[u.ID] = &cp
	s.email[key] = u.ID
	return s.persist()
}

// GetUser returns a copy of the user with the given id.
func (s *Store) GetUser(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.st.Users[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

// GetUserByEmail looks a user up by (case-insensitive) email.
func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.email[strings.ToLower(email)]
	if !ok {
		return nil, ErrNotFound
	}
	u := s.st.Users[id]
	cp := *u
	return &cp, nil
}

// UpdateUser replaces the stored user (matched by ID).
func (s *Store) UpdateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.st.Users[u.ID]; !ok {
		return ErrNotFound
	}
	cp := *u
	s.st.Users[u.ID] = &cp
	s.email[strings.ToLower(u.Email)] = u.ID
	return s.persist()
}

// ListUsers returns all users sorted by most-recently-active first.
func (s *Store) ListUsers() []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.User, 0, len(s.st.Users))
	for _, u := range s.st.Users {
		cp := *u
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActiveAt.After(out[j].LastActiveAt) })
	return out
}

// TouchUser updates a user's last-active timestamp (best-effort).
func (s *Store) TouchUser(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.st.Users[id]; ok {
		u.LastActiveAt = time.Now()
		_ = s.persist()
	}
}

// --- Decisions ---

// CreateDecision stores a decision, assigning a friendly D-##### id.
func (s *Store) CreateDecision(d *models.Decision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d.ID == "" {
		s.st.DecisionSeq++
		d.ID = fmt.Sprintf("D-%05d", 88000+s.st.DecisionSeq)
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	cp := *d
	s.st.Decisions[d.ID] = &cp
	return s.persist()
}

// GetDecision returns a copy of one decision.
func (s *Store) GetDecision(id string) (*models.Decision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.st.Decisions[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

// UpdateDecision replaces a decision (matched by ID).
func (s *Store) UpdateDecision(d *models.Decision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.st.Decisions[d.ID]; !ok {
		return ErrNotFound
	}
	cp := *d
	s.st.Decisions[d.ID] = &cp
	return s.persist()
}

// ListDecisions returns a user's decisions (newest first), optionally filtered
// by status ("all"/""/"resist"/"buy"/"pending").
func (s *Store) ListDecisions(userID, filter string) []*models.Decision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*models.Decision{}
	for _, d := range s.st.Decisions {
		if d.UserID != userID {
			continue
		}
		if !decisionMatches(d, filter) {
			continue
		}
		cp := *d
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

// ListAllDecisions returns every decision (admin feed), newest first.
func (s *Store) ListAllDecisions(filter string) []*models.Decision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*models.Decision{}
	for _, d := range s.st.Decisions {
		if filter != "" && filter != "all" && string(d.Verdict) != filter {
			continue
		}
		cp := *d
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func decisionMatches(d *models.Decision, filter string) bool {
	switch filter {
	case "", "all":
		return d.Status != models.StatusPending
	case "resist":
		return d.Status == models.StatusResisted
	case "buy":
		return d.Status == models.StatusBought
	case "pending":
		return d.Status == models.StatusPending
	default:
		return string(d.Verdict) == filter
	}
}

// --- Wishlist ---

// AddWishlist parks an item in the cooling-off list.
func (s *Store) AddWishlist(w *models.WishlistItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w.ID == "" {
		w.ID = uuid.NewString()
	}
	cp := *w
	s.st.Wishlist[w.ID] = &cp
	return s.persist()
}

// GetWishlistItem returns one wishlist entry.
func (s *Store) GetWishlistItem(id string) (*models.WishlistItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.st.Wishlist[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *w
	return &cp, nil
}

// ListWishlist returns a user's wishlist (soonest-expiring first).
func (s *Store) ListWishlist(userID string) []*models.WishlistItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*models.WishlistItem{}
	for _, w := range s.st.Wishlist {
		if w.UserID != userID {
			continue
		}
		cp := *w
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ExpiresAt.Before(out[j].ExpiresAt) })
	return out
}

// DeleteWishlist removes a wishlist entry.
func (s *Store) DeleteWishlist(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.st.Wishlist[id]; !ok {
		return ErrNotFound
	}
	delete(s.st.Wishlist, id)
	return s.persist()
}

// --- Transactions ---

// AddTransaction records a payment.
func (s *Store) AddTransaction(t *models.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		s.st.TxnSeq++
		t.ID = fmt.Sprintf("T-%04d", 5500+s.st.TxnSeq)
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	cp := *t
	s.st.Transactions[t.ID] = &cp
	return s.persist()
}

// ListTransactions returns all transactions, newest first.
func (s *Store) ListTransactions() []*models.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Transaction, 0, len(s.st.Transactions))
	for _, t := range s.st.Transactions {
		cp := *t
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

// --- Config & content (singletons / lists) ---

// AIConfig returns the live engine configuration.
func (s *Store) AIConfig() models.AIConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.st.AIConfig
}

// SetAIConfig replaces the engine configuration.
func (s *Store) SetAIConfig(c models.AIConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.UpdatedAt = time.Now()
	s.st.AIConfig = c
	return s.persist()
}

// Categories returns the spending categories.
func (s *Store) Categories() []models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Category(nil), s.st.Categories...)
}

// SetCategories replaces the category list.
func (s *Store) SetCategories(c []models.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Categories = c
	return s.persist()
}

// Skins returns the app-icon skins.
func (s *Store) Skins() []models.Skin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Skin(nil), s.st.Skins...)
}

// Plans returns the purchasable subscription tiers.
func (s *Store) Plans() []models.PlanOption {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.PlanOption(nil), s.st.Plans...)
}

// Perks returns the Plus marketing bullets.
func (s *Store) Perks() []models.PlusPerk {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.PlusPerk(nil), s.st.Perks...)
}

// Pushes returns operations campaigns, newest first.
func (s *Store) Pushes() []*models.Push {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Push, 0, len(s.st.Pushes))
	for _, p := range s.st.Pushes {
		cp := *p
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

// CreatePush stores a campaign.
func (s *Store) CreatePush(p *models.Push) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		s.st.PushSeq++
		p.ID = fmt.Sprintf("P-%03d", 200+s.st.PushSeq)
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	cp := *p
	s.st.Pushes[p.ID] = &cp
	return s.persist()
}

// SetSkins replaces the app-icon skin catalogue.
func (s *Store) SetSkins(v []models.Skin) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Skins = v
	return s.persist()
}

// SetPlans replaces the subscription tiers.
func (s *Store) SetPlans(v []models.PlanOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Plans = v
	return s.persist()
}

// SetPerks replaces the Plus marketing bullets.
func (s *Store) SetPerks(v []models.PlusPerk) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.st.Perks = v
	return s.persist()
}
