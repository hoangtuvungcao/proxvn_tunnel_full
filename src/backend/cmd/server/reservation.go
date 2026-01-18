package main

import (
	"log"
	"time"
)

// cleanupPortReservations removes expired port reservations every minute
func (s *server) cleanupPortReservations() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.reservationMu.Lock()
		now := time.Now()
		for key, reservation := range s.portReservations {
			if now.After(reservation.expiresAt) {
				delete(s.portReservations, key)
				log.Printf("[server] Expired port reservation for client key %s (port %d)", key, reservation.port)
			}
		}
		s.reservationMu.Unlock()
	}
}
