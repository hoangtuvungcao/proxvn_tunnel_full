package database

import (
	"database/sql"
	"time"
)

type Reservation struct {
	ClientKey string
	Port      int
	Subdomain string
	ExpiresAt time.Time
}

func (d *Database) SaveReservation(clientKey string, port int, subdomain string, duration time.Duration) error {
	query := `
		INSERT INTO reservations (client_key, port, subdomain, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(client_key) DO UPDATE SET
			port = CASE WHEN $2 > 0 THEN $2 ELSE port END,
			subdomain = CASE WHEN $3 != '' THEN $3 ELSE subdomain END,
			expires_at = $4
	`
	expiresAt := time.Now().Add(duration)
	_, err := d.db.Exec(query, clientKey, port, subdomain, expiresAt)
	return err
}

func (d *Database) GetReservation(clientKey string) (int, string, bool, error) {
	query := `
		SELECT port, COALESCE(subdomain, ''), expires_at FROM reservations
		WHERE client_key = $1 AND expires_at > CURRENT_TIMESTAMP
	`
	var port int
	var subdomain string
	var expiresAt time.Time

	err := d.db.QueryRow(query, clientKey).Scan(&port, &subdomain, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	return port, subdomain, true, nil
}

func (d *Database) DeleteReservation(clientKey string) error {
	query := `DELETE FROM reservations WHERE client_key = $1`
	_, err := d.db.Exec(query, clientKey)
	return err
}

func (d *Database) CleanupExpiredReservations() error {
	query := `DELETE FROM reservations WHERE expires_at <= CURRENT_TIMESTAMP`
	_, err := d.db.Exec(query)
	return err
}

func (d *Database) GetExpiredReservations() ([]int, []string, error) {
	query := `SELECT port, COALESCE(subdomain, '') FROM reservations WHERE expires_at <= CURRENT_TIMESTAMP`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ports []int
	var subs []string
	for rows.Next() {
		var port int
		var sub string
		if err := rows.Scan(&port, &sub); err == nil {
			if port > 0 {
				ports = append(ports, port)
			}
			if sub != "" {
				subs = append(subs, sub)
			}
		}
	}
	return ports, subs, nil
}

func (d *Database) IsSubdomainReservedByOther(subdomain, clientKey string) (bool, error) {
	if subdomain == "" {
		return false, nil
	}
	query := `
		SELECT COUNT(*) FROM reservations
		WHERE subdomain = $1 AND client_key != $2 AND expires_at > CURRENT_TIMESTAMP
	`
	var count int
	err := d.db.QueryRow(query, subdomain, clientKey).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) IsPortReservedByOther(port int, clientKey string) (bool, error) {
	if port <= 0 {
		return false, nil
	}
	query := `
		SELECT COUNT(*) FROM reservations
		WHERE port = $1 AND client_key != $2 AND expires_at > CURRENT_TIMESTAMP
	`
	var count int
	err := d.db.QueryRow(query, port, clientKey).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
