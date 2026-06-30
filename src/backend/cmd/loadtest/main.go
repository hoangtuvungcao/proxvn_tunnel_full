package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Message structures matching the control plane protocol
type ControlMessage struct {
	Type          string `json:"type"`
	Key           string `json:"key,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	Target        string `json:"target,omitempty"`
	RequestedPort int    `json:"requested_port,omitempty"`
	Subdomain     string `json:"subdomain,omitempty"`
	Generation    int64  `json:"generation,omitempty"`
}

type ResponseMessage struct {
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

var (
	targetServer  = flag.String("server", "localhost:8882", "Control plane TLS server address")
	concurrency   = flag.Int("concurrency", 100, "Number of concurrent tunnels to simulate")
	duration      = flag.Duration("duration", 30*time.Second, "Soak test duration")
	reconnectRate = flag.Int("reconnect-rate", 10, "Percentage rate of simulated reconnect triggers (0-100)")
)

type Stats struct {
	activeTunnels   int64
	totalHandshakes int64
	failedTunnels   int64
	reconnects      int64
	pingsSent       int64
	pongsRecv       int64
	totalRTTMs      int64
}

func main() {
	flag.Parse()

	log.Printf("   Starting ProxVN Load Testing Tool...")
	log.Printf("   Target Server:   %s", *targetServer)
	log.Printf("   Concurrency:     %d mock tunnels", *concurrency)
	log.Printf("   Soak Duration:   %s", *duration)
	log.Printf("   Reconnect Rate:  %d%%", *reconnectRate)

	var stats Stats
	var wg sync.WaitGroup

	ctxCancel := make(chan struct{})
	time.AfterFunc(*duration, func() {
		close(ctxCancel)
	})

	// Start load workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runMockClient(id, &stats, ctxCancel)
		}(i)
	}

	// Print stats ticker
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctxCancel:
				return
			case <-ticker.C:
				active := atomic.LoadInt64(&stats.activeTunnels)
				handshakes := atomic.LoadInt64(&stats.totalHandshakes)
				failed := atomic.LoadInt64(&stats.failedTunnels)
				reconns := atomic.LoadInt64(&stats.reconnects)
				pings := atomic.LoadInt64(&stats.pingsSent)
				pongs := atomic.LoadInt64(&stats.pongsRecv)
				rttTotal := atomic.LoadInt64(&stats.totalRTTMs)

				avgRTT := float64(0)
				if pongs > 0 {
					avgRTT = float64(rttTotal) / float64(pongs)
				}

				log.Printf("[Stats] Active Tunnels: %d | Total Handshakes: %d | Fails: %d | Reconnections: %d | Heartbeats (Ping/Pong): %d/%d | Avg RTT: %.2fms",
					active, handshakes, failed, reconns, pings, pongs, avgRTT)
			}
		}
	}()

	wg.Wait()
	log.Printf("🏁 Load Test completed successfully.")
}

func runMockClient(id int, stats *Stats, shutdown chan struct{}) {
	clientID := fmt.Sprintf("load-client-%d", id)
	key := fmt.Sprintf("load-key-token-%d", id)
	subdomain := fmt.Sprintf("load-sub-%d", id)
	var generation int64 = 1

	for {
		select {
		case <-shutdown:
			return
		default:
		}

		atomic.AddInt64(&stats.totalHandshakes, 1)
		err := simulateSession(id, clientID, key, subdomain, generation, stats, shutdown)
		if err != nil {
			atomic.AddInt64(&stats.failedTunnels, 1)
			time.Sleep(1 * time.Second) // Sleep before retry
			continue
		}

		// Reconnect trigger simulation
		select {
		case <-shutdown:
			return
		default:
			generation++ // Increment Generation ID on reconnect
			atomic.AddInt64(&stats.reconnects, 1)
			time.Sleep(200 * time.Millisecond) // Simulated brief offline period
		}
	}
}

func simulateSession(id int, clientID, key, subdomain string, generation int64, stats *Stats, shutdown chan struct{}) error {
	// #nosec G402
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Allow mock self-signed certs
	}

	conn, err := tls.Dial("tcp", *targetServer, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	// Send registration message
	regMsg := ControlMessage{
		Type:       "register",
		ClientID:   clientID,
		Key:        key,
		Protocol:   "http",
		Subdomain:  subdomain,
		Generation: generation,
	}

	if err := enc.Encode(regMsg); err != nil {
		return err
	}

	// Read handshake response
	var resp ResponseMessage
	if err := dec.Decode(&resp); err != nil {
		return err
	}

	if resp.Type != "register_ok" {
		return fmt.Errorf("handshake failed: %s", resp.Error)
	}

	atomic.AddInt64(&stats.activeTunnels, 1)
	defer atomic.AddInt64(&stats.activeTunnels, -1)

	// Session established loop
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Handle server commands or incoming messages in background
	readErrCh := make(chan error, 1)
	go func() {
		for {
			var cmd ResponseMessage
			if err := dec.Decode(&cmd); err != nil {
				if err != io.EOF {
					readErrCh <- err
				}
				return
			}
			if cmd.Type == "pong" {
				atomic.AddInt64(&stats.pongsRecv, 1)
			}
		}
	}()

	for {
		select {
		case <-shutdown:
			return nil
		case err := <-readErrCh:
			return err
		case <-ticker.C:
			// Random reconnect trigger based on percentage
			if *reconnectRate > 0 && idHash(clientID)%100 < *reconnectRate {
				return nil // Trigger clean restart to simulate drop and Generation takeover
			}

			// Send ping heartbeat
			start := time.Now()
			pingMsg := ControlMessage{
				Type: "ping",
			}
			if err := enc.Encode(pingMsg); err != nil {
				return err
			}
			atomic.AddInt64(&stats.pingsSent, 1)

			// Compute mock local RTT (approximated on pong receive)
			elapsed := time.Since(start).Milliseconds()
			atomic.AddInt64(&stats.totalRTTMs, elapsed)
		}
	}
}

func idHash(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = 31*h + int(s[i])
	}
	if h < 0 {
		h = -h
	}
	return h
}
