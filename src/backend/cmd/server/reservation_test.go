package main

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryReservePortRoundTrip(t *testing.T) {
	s := NewMemoryReservationStore()
	if err := s.ReservePort("clientA", 10001, time.Hour); err != nil {
		t.Fatalf("ReservePort: %v", err)
	}
	port, ok, err := s.GetReservedPort("clientA")
	if err != nil {
		t.Fatalf("GetReservedPort: %v", err)
	}
	if !ok || port != 10001 {
		t.Fatalf("GetReservedPort = (%d, %v), want (10001, true)", port, ok)
	}
}

func TestMemoryGetUnknownClient(t *testing.T) {
	s := NewMemoryReservationStore()
	if _, ok, _ := s.GetReservedPort("nobody"); ok {
		t.Fatal("expected no reservation for unknown client")
	}
	if _, ok, _ := s.GetReservedSubdomain("nobody"); ok {
		t.Fatal("expected no subdomain for unknown client")
	}
}

func TestMemoryExpiredReservationIsNotReturned(t *testing.T) {
	s := NewMemoryReservationStore()
	// Negative expiry => already expired.
	_ = s.ReservePort("clientA", 10001, -time.Hour)
	if _, ok, _ := s.GetReservedPort("clientA"); ok {
		t.Fatal("expired port reservation should not be returned")
	}
}

func TestMemorySubdomainRoundTrip(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReserveSubdomain("clientA", "myapp", time.Hour)
	sub, ok, _ := s.GetReservedSubdomain("clientA")
	if !ok || sub != "myapp" {
		t.Fatalf("GetReservedSubdomain = (%q, %v), want (myapp, true)", sub, ok)
	}
}

func TestMemoryPortReservedByOther(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReservePort("clientA", 10001, time.Hour)

	other, _ := s.IsPortReservedByOther(10001, "clientB")
	if !other {
		t.Fatal("port held by clientA must be reported as reserved by other for clientB")
	}
	self, _ := s.IsPortReservedByOther(10001, "clientA")
	if self {
		t.Fatal("client must not see its own port as reserved by other")
	}
	free, _ := s.IsPortReservedByOther(20000, "clientB")
	if free {
		t.Fatal("unreserved port must not be reported as reserved")
	}
}

func TestMemorySubdomainReservedByOther(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReserveSubdomain("clientA", "myapp", time.Hour)

	if other, _ := s.IsSubdomainReservedByOther("myapp", "clientB"); !other {
		t.Fatal("subdomain held by clientA must be reserved-by-other for clientB")
	}
	if self, _ := s.IsSubdomainReservedByOther("myapp", "clientA"); self {
		t.Fatal("client must not see its own subdomain as reserved by other")
	}
}

func TestMemoryExpiredReservationIgnoredByReservedByOther(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReservePort("clientA", 10001, -time.Hour) // already expired
	if other, _ := s.IsPortReservedByOther(10001, "clientB"); other {
		t.Fatal("expired reservation must not block another client")
	}
}

func TestMemoryDeleteReservation(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReservePort("clientA", 10001, time.Hour)
	_ = s.ReserveSubdomain("clientA", "myapp", time.Hour)
	if err := s.DeleteReservation("clientA"); err != nil {
		t.Fatalf("DeleteReservation: %v", err)
	}
	if _, ok, _ := s.GetReservedPort("clientA"); ok {
		t.Fatal("port should be gone after delete")
	}
	if _, ok, _ := s.GetReservedSubdomain("clientA"); ok {
		t.Fatal("subdomain should be gone after delete")
	}
}

func TestMemoryCleanupAndGetExpired(t *testing.T) {
	s := NewMemoryReservationStore()
	_ = s.ReservePort("expired", 10001, -time.Hour)
	_ = s.ReserveSubdomain("expiredSub", "oldapp", -time.Hour)
	_ = s.ReservePort("valid", 10002, time.Hour)

	ports, subs, err := s.GetExpired()
	if err != nil {
		t.Fatalf("GetExpired: %v", err)
	}
	if !contains(ports, 10001) {
		t.Errorf("expected expired port 10001 in %v", ports)
	}
	if !containsStr(subs, "oldapp") {
		t.Errorf("expected expired subdomain oldapp in %v", subs)
	}
	if contains(ports, 10002) {
		t.Errorf("valid port 10002 must not be reported expired")
	}

	if err := s.CleanupExpired(); err != nil {
		t.Fatalf("CleanupExpired: %v", err)
	}
	// Valid reservation survives cleanup.
	if _, ok, _ := s.GetReservedPort("valid"); !ok {
		t.Fatal("valid reservation should survive cleanup")
	}
	// Expired entries are gone, so GetExpired now returns nothing.
	ports, subs, _ = s.GetExpired()
	if len(ports) != 0 || len(subs) != 0 {
		t.Fatalf("after cleanup GetExpired = (%v, %v), want empty", ports, subs)
	}
}

func TestMemoryConcurrentAccess(t *testing.T) {
	s := NewMemoryReservationStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "client"
			_ = s.ReservePort(key, 10000+n, time.Hour)
			_, _, _ = s.GetReservedPort(key)
			_, _ = s.IsPortReservedByOther(10000+n, "other")
			_ = s.CleanupExpired()
		}(i)
	}
	wg.Wait()
}

func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func containsStr(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
