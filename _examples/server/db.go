package main

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	storesFile = "data/stores.json"
	petsFile   = "data/pets.json"
)

// StoreRecord is the flat file record for a store.
type StoreRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// PetRecord is the flat file record for a pet.
type PetRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Species  string `json:"species"`
	StoreID  string `json:"store_id"`
}

type DB struct {
	mu     sync.RWMutex
	stores []StoreRecord
	pets   []PetRecord
}

func NewDB() (*DB, error) {
	db := &DB{
		stores: []StoreRecord{},
		pets:   []PetRecord{},
	}
	if err := db.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return db, nil
}

func (db *DB) load() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if b, err := os.ReadFile(storesFile); err == nil {
		_ = json.Unmarshal(b, &db.stores)
	}
	if b, err := os.ReadFile(petsFile); err == nil {
		_ = json.Unmarshal(b, &db.pets)
	}
	return nil
}

// saveLocked persists the in-memory state to disk. Caller must hold db.mu (write lock).
func (db *DB) saveLocked() error {
	os.MkdirAll("data", 0755)
	stores := make([]StoreRecord, len(db.stores))
	copy(stores, db.stores)
	pets := make([]PetRecord, len(db.pets))
	copy(pets, db.pets)

	if b, err := json.MarshalIndent(stores, "", "  "); err != nil {
		return err
	} else if err := os.WriteFile(storesFile, b, 0644); err != nil {
		return err
	}
	if b, err := json.MarshalIndent(pets, "", "  "); err != nil {
		return err
	} else if err := os.WriteFile(petsFile, b, 0644); err != nil {
		return err
	}
	return nil
}

func (db *DB) ListStores() []StoreRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()
	out := make([]StoreRecord, len(db.stores))
	copy(out, db.stores)
	return out
}

func (db *DB) GetStore(id string) (StoreRecord, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, s := range db.stores {
		if s.ID == id {
			return s, true
		}
	}
	return StoreRecord{}, false
}

func (db *DB) CreateStore(s StoreRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.stores = append(db.stores, s)
	return db.saveLocked()
}

func (db *DB) UpdateStore(id string, fn func(*StoreRecord)) (StoreRecord, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i := range db.stores {
		if db.stores[i].ID == id {
			fn(&db.stores[i])
			_ = db.saveLocked()
			return db.stores[i], true
		}
	}
	return StoreRecord{}, false
}

func (db *DB) DeleteStore(id string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i, s := range db.stores {
		if s.ID == id {
			db.stores = append(db.stores[:i], db.stores[i+1:]...)
			_ = db.saveLocked()
			return true
		}
	}
	return false
}

func (db *DB) ListPets() []PetRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()
	out := make([]PetRecord, len(db.pets))
	copy(out, db.pets)
	return out
}

func (db *DB) GetPet(id string) (PetRecord, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, p := range db.pets {
		if p.ID == id {
			return p, true
		}
	}
	return PetRecord{}, false
}

func (db *DB) CreatePet(p PetRecord) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.pets = append(db.pets, p)
	return db.saveLocked()
}

func (db *DB) UpdatePet(id string, fn func(*PetRecord)) (PetRecord, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i := range db.pets {
		if db.pets[i].ID == id {
			fn(&db.pets[i])
			_ = db.saveLocked()
			return db.pets[i], true
		}
	}
	return PetRecord{}, false
}

func (db *DB) DeletePet(id string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i, p := range db.pets {
		if p.ID == id {
			db.pets = append(db.pets[:i], db.pets[i+1:]...)
			_ = db.saveLocked()
			return true
		}
	}
	return false
}

func (db *DB) PetsByStoreID(storeID string) []PetRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var out []PetRecord
	for _, p := range db.pets {
		if p.StoreID == storeID {
			out = append(out, p)
		}
	}
	return out
}
