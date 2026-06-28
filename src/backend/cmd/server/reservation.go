package main

import (
	"log"
	"sync"
	"time"

	"proxvn/backend/internal/database"
)

type ReservationStore interface {
	ReservePort(clientKey string, port int, expiry time.Duration) error
	ReserveSubdomain(clientKey string, subdomain string, expiry time.Duration) error
	GetReservedPort(clientKey string) (int, bool, error)
	GetReservedSubdomain(clientKey string) (string, bool, error)
	IsSubdomainReservedByOther(subdomain, clientKey string) (bool, error)
	IsPortReservedByOther(port int, clientKey string) (bool, error)
	DeleteReservation(clientKey string) error
	CleanupExpired() error
	GetExpired() ([]int, []string, error)
}

// SQLiteReservationStore uses SQLite database for persistent storage
type SQLiteReservationStore struct {
	db *database.Database
}

func NewSQLiteReservationStore(db *database.Database) *SQLiteReservationStore {
	return &SQLiteReservationStore{db: db}
}

func (s *SQLiteReservationStore) ReservePort(clientKey string, port int, expiry time.Duration) error {
	return s.db.SaveReservation(clientKey, port, "", expiry)
}

func (s *SQLiteReservationStore) ReserveSubdomain(clientKey string, subdomain string, expiry time.Duration) error {
	return s.db.SaveReservation(clientKey, 0, subdomain, expiry)
}

func (s *SQLiteReservationStore) GetReservedPort(clientKey string) (int, bool, error) {
	port, _, exists, err := s.db.GetReservation(clientKey)
	return port, exists && port > 0, err
}

func (s *SQLiteReservationStore) GetReservedSubdomain(clientKey string) (string, bool, error) {
	_, sub, exists, err := s.db.GetReservation(clientKey)
	return sub, exists && sub != "", err
}

func (s *SQLiteReservationStore) IsSubdomainReservedByOther(subdomain, clientKey string) (bool, error) {
	return s.db.IsSubdomainReservedByOther(subdomain, clientKey)
}

func (s *SQLiteReservationStore) IsPortReservedByOther(port int, clientKey string) (bool, error) {
	return s.db.IsPortReservedByOther(port, clientKey)
}

func (s *SQLiteReservationStore) DeleteReservation(clientKey string) error {
	return s.db.DeleteReservation(clientKey)
}

func (s *SQLiteReservationStore) CleanupExpired() error {
	return s.db.CleanupExpiredReservations()
}

func (s *SQLiteReservationStore) GetExpired() ([]int, []string, error) {
	return s.db.GetExpiredReservations()
}

// MemoryReservationStore fallback when database is not connected
type MemoryReservationStore struct {
	mu         sync.RWMutex
	ports      map[string]int
	subdomains map[string]string
	expiresAt  map[string]time.Time
}

func NewMemoryReservationStore() *MemoryReservationStore {
	return &MemoryReservationStore{
		ports:      make(map[string]int),
		subdomains: make(map[string]string),
		expiresAt:  make(map[string]time.Time),
	}
}

func (s *MemoryReservationStore) ReservePort(clientKey string, port int, expiry time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ports[clientKey] = port
	s.expiresAt[clientKey] = time.Now().Add(expiry)
	return nil
}

func (s *MemoryReservationStore) ReserveSubdomain(clientKey string, subdomain string, expiry time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subdomains[clientKey] = subdomain
	s.expiresAt[clientKey] = time.Now().Add(expiry)
	return nil
}

func (s *MemoryReservationStore) GetReservedPort(clientKey string) (int, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, exists := s.expiresAt[clientKey]
	if !exists || time.Now().After(exp) {
		return 0, false, nil
	}
	port, exists := s.ports[clientKey]
	return port, exists && port > 0, nil
}

func (s *MemoryReservationStore) GetReservedSubdomain(clientKey string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, exists := s.expiresAt[clientKey]
	if !exists || time.Now().After(exp) {
		return "", false, nil
	}
	sub, exists := s.subdomains[clientKey]
	return sub, exists && sub != "", nil
}

func (s *MemoryReservationStore) IsSubdomainReservedByOther(subdomain, clientKey string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	for key, sub := range s.subdomains {
		if key != clientKey && sub == subdomain {
			if exp, exists := s.expiresAt[key]; exists && now.Before(exp) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *MemoryReservationStore) IsPortReservedByOther(port int, clientKey string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	for key, p := range s.ports {
		if key != clientKey && p == port {
			if exp, exists := s.expiresAt[key]; exists && now.Before(exp) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *MemoryReservationStore) DeleteReservation(clientKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.ports, clientKey)
	delete(s.subdomains, clientKey)
	delete(s.expiresAt, clientKey)
	return nil
}

func (s *MemoryReservationStore) CleanupExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, exp := range s.expiresAt {
		if now.After(exp) {
			delete(s.ports, key)
			delete(s.subdomains, key)
			delete(s.expiresAt, key)
		}
	}
	return nil
}

func (s *MemoryReservationStore) GetExpired() ([]int, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var ports []int
	var subs []string
	for key, exp := range s.expiresAt {
		if now.After(exp) {
			if port, ok := s.ports[key]; ok && port > 0 {
				ports = append(ports, port)
			}
			if sub, ok := s.subdomains[key]; ok && sub != "" {
				subs = append(subs, sub)
			}
		}
	}
	return ports, subs, nil
}

// startReservationCleanupTicker runs cleanup every minute and releases ports back to the server
func startReservationCleanupTicker(store ReservationStore, srv *server) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			// Find expired ports
			ports, _, err := store.GetExpired()
			if err == nil {
				for _, port := range ports {
					srv.releasePort(port)
					log.Printf("[server] Released expired reserved port: %d", port)
				}
			}

			// Perform cleanup
			if err := store.CleanupExpired(); err != nil {
				log.Printf("[server] Failed to cleanup expired reservations: %v", err)
			}
		}
	}()
}
