package store

import (
	"fmt"

	"github.com/mosson9/lim/backend/internal/models"
)

// Store is the persistence contract the rest of the application depends on.
//
// Two implementations satisfy it: FileStore (in-memory + JSON snapshot, the
// zero-config default) and PostgresStore (for real deployments). The HTTP
// handlers, seeder and server only ever see this interface, so swapping the
// backing store needs no changes anywhere else.
type Store interface {
	// lifecycle
	IsEmpty() bool
	Save() error

	// users
	CreateUser(u *models.User) error
	GetUser(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByOriginalTransactionID(otx string) (*models.User, error)
	UpdateUser(u *models.User) error
	ListUsers() []*models.User
	TouchUser(id string)

	// decisions
	CreateDecision(d *models.Decision) error
	GetDecision(id string) (*models.Decision, error)
	UpdateDecision(d *models.Decision) error
	ListDecisions(userID, filter string) []*models.Decision
	ListAllDecisions(filter string) []*models.Decision

	// wishlist
	AddWishlist(w *models.WishlistItem) error
	GetWishlistItem(id string) (*models.WishlistItem, error)
	ListWishlist(userID string) []*models.WishlistItem
	DeleteWishlist(id string) error

	// transactions
	AddTransaction(t *models.Transaction) error
	ListTransactions() []*models.Transaction

	// config & content
	AIConfig() models.AIConfig
	SetAIConfig(c models.AIConfig) error
	Categories() []models.Category
	SetCategories(c []models.Category) error
	Skins() []models.Skin
	SetSkins(v []models.Skin) error
	Plans() []models.PlanOption
	SetPlans(v []models.PlanOption) error
	Perks() []models.PlusPerk
	SetPerks(v []models.PlusPerk) error
	Pushes() []*models.Push
	CreatePush(p *models.Push) error
}

// Compile-time guarantees that both implementations satisfy the interface.
var (
	_ Store = (*FileStore)(nil)
	_ Store = (*PostgresStore)(nil)
)

// Open selects a store implementation from configuration. When databaseURL is
// non-empty a PostgresStore is used; otherwise the JSON-file store at dataFile.
func Open(databaseURL, dataFile string) (Store, error) {
	if databaseURL != "" {
		s, err := OpenPostgres(databaseURL)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		return s, nil
	}
	s, err := OpenFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("open file store: %w", err)
	}
	return s, nil
}
