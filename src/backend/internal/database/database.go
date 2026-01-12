package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"proxvn/backend/internal/models"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	database := &Database{db: db}
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

func (d *Database) migrate() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) DEFAULT 'user',
			api_key VARCHAR(255) UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS tunnels (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			protocol VARCHAR(10) NOT NULL CHECK (protocol IN ('tcp', 'udp')),
			local_host VARCHAR(255) NOT NULL,
			local_port INTEGER NOT NULL CHECK (local_port > 0 AND local_port < 65536),
			public_port INTEGER UNIQUE,
			status VARCHAR(20) DEFAULT 'inactive',
			client_id VARCHAR(255),
			auth_token VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS connections (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tunnel_id UUID NOT NULL REFERENCES tunnels(id) ON DELETE CASCADE,
			remote_addr INET NOT NULL,
			connected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			disconnected_at TIMESTAMP,
			bytes_up BIGINT DEFAULT 0,
			bytes_down BIGINT DEFAULT 0,
			duration BIGINT DEFAULT 0
		);
		`,
		`
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
		CREATE INDEX IF NOT EXISTS idx_tunnels_user_id ON tunnels(user_id);
		CREATE INDEX IF NOT EXISTS idx_tunnels_public_port ON tunnels(public_port);
		CREATE INDEX IF NOT EXISTS idx_tunnels_status ON tunnels(status);
		CREATE INDEX IF NOT EXISTS idx_connections_tunnel_id ON connections(tunnel_id);
		CREATE INDEX IF NOT EXISTS idx_connections_connected_at ON connections(connected_at);
		`,
		`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
		`,
		`
		CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		CREATE TRIGGER update_tunnels_updated_at BEFORE UPDATE ON tunnels
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	log.Println("Database migration completed successfully")
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password, role, api_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	err := d.db.QueryRow(query, user.Username, user.Email, user.Password, user.Role, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	return err
}

func (d *Database) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, role, api_key, created_at, updated_at
		FROM users WHERE username = $1
	`
	
	err := d.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.APIKey, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (d *Database) GetUserByAPIKey(apiKey string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, role, api_key, created_at, updated_at
		FROM users WHERE api_key = $1
	`
	
	err := d.db.QueryRow(query, apiKey).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.APIKey, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (d *Database) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, username, email, password, role, api_key, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Role, &user.APIKey, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, rows.Err()
}

func (d *Database) CreateTunnel(tunnel *models.Tunnel) error {
	query := `
		INSERT INTO tunnels (user_id, name, protocol, local_host, local_port, public_port, status, client_id, auth_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at, last_seen
	`
	
	err := d.db.QueryRow(query,
		tunnel.UserID, tunnel.Name, tunnel.Protocol,
		tunnel.LocalHost, tunnel.LocalPort, tunnel.PublicPort,
		tunnel.Status, tunnel.ClientID, tunnel.AuthToken,
	).Scan(&tunnel.ID, &tunnel.CreatedAt, &tunnel.UpdatedAt, &tunnel.LastSeen)
	
	return err
}

func (d *Database) GetTunnelsByUserID(userID string) ([]*models.Tunnel, error) {
	query := `
		SELECT id, user_id, name, protocol, local_host, local_port, public_port, 
			   status, client_id, auth_token, created_at, updated_at, last_seen
		FROM tunnels WHERE user_id = $1 ORDER BY created_at DESC
	`
	
	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tunnels []*models.Tunnel
	for rows.Next() {
		tunnel := &models.Tunnel{}
		err := rows.Scan(
			&tunnel.ID, &tunnel.UserID, &tunnel.Name, &tunnel.Protocol,
			&tunnel.LocalHost, &tunnel.LocalPort, &tunnel.PublicPort,
			&tunnel.Status, &tunnel.ClientID, &tunnel.AuthToken,
			&tunnel.CreatedAt, &tunnel.UpdatedAt, &tunnel.LastSeen,
		)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, tunnel)
	}
	
	return tunnels, rows.Err()
}

func (d *Database) GetTunnelByID(tunnelID string) (*models.Tunnel, error) {
	tunnel := &models.Tunnel{}
	query := `
		SELECT id, user_id, name, protocol, local_host, local_port, public_port,
			   status, client_id, auth_token, created_at, updated_at, last_seen
		FROM tunnels WHERE id = $1
	`
	
	err := d.db.QueryRow(query, tunnelID).Scan(
		&tunnel.ID, &tunnel.UserID, &tunnel.Name, &tunnel.Protocol,
		&tunnel.LocalHost, &tunnel.LocalPort, &tunnel.PublicPort,
		&tunnel.Status, &tunnel.ClientID, &tunnel.AuthToken,
		&tunnel.CreatedAt, &tunnel.UpdatedAt, &tunnel.LastSeen,
	)
	
	if err != nil {
		return nil, err
	}
	
	return tunnel, nil
}

func (d *Database) GetTunnelByPublicPort(port int) (*models.Tunnel, error) {
	tunnel := &models.Tunnel{}
	query := `
		SELECT id, user_id, name, protocol, local_host, local_port, public_port,
			   status, client_id, auth_token, created_at, updated_at, last_seen
		FROM tunnels WHERE public_port = $1
	`
	
	err := d.db.QueryRow(query, port).Scan(
		&tunnel.ID, &tunnel.UserID, &tunnel.Name, &tunnel.Protocol,
		&tunnel.LocalHost, &tunnel.LocalPort, &tunnel.PublicPort,
		&tunnel.Status, &tunnel.ClientID, &tunnel.AuthToken,
		&tunnel.CreatedAt, &tunnel.UpdatedAt, &tunnel.LastSeen,
	)
	
	if err != nil {
		return nil, err
	}
	
	return tunnel, nil
}

func (d *Database) GetAllTunnels() ([]*models.Tunnel, error) {
	query := `
		SELECT id, user_id, name, protocol, local_host, local_port, public_port,
			   status, client_id, auth_token, created_at, updated_at, last_seen
		FROM tunnels ORDER BY created_at DESC
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tunnels []*models.Tunnel
	for rows.Next() {
		tunnel := &models.Tunnel{}
		err := rows.Scan(
			&tunnel.ID, &tunnel.UserID, &tunnel.Name, &tunnel.Protocol,
			&tunnel.LocalHost, &tunnel.LocalPort, &tunnel.PublicPort,
			&tunnel.Status, &tunnel.ClientID, &tunnel.AuthToken,
			&tunnel.CreatedAt, &tunnel.UpdatedAt, &tunnel.LastSeen,
		)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, tunnel)
	}
	
	return tunnels, rows.Err()
}

func (d *Database) UpdateTunnel(tunnelID, userID, name, localHost string, localPort int) error {
	query := `
		UPDATE tunnels 
		SET name = COALESCE(NULLIF($3, ''), name),
		    local_host = COALESCE(NULLIF($4, ''), local_host),
		    local_port = CASE WHEN $5 > 0 THEN $5 ELSE local_port END
		WHERE id = $1 AND user_id = $2
	`
	result, err := d.db.Exec(query, tunnelID, userID, name, localHost, localPort)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("tunnel not found or access denied")
	}
	return nil
}

func (d *Database) UpdateTunnelStatus(tunnelID string, status string) error {
	query := `
		UPDATE tunnels SET status = $1, last_seen = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := d.db.Exec(query, status, tunnelID)
	return err
}

func (d *Database) UpdateTunnelLastSeen(tunnelID string) error {
	query := `UPDATE tunnels SET last_seen = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := d.db.Exec(query, tunnelID)
	return err
}

func (d *Database) DeleteTunnel(tunnelID, userID string) error {
	query := `DELETE FROM tunnels WHERE id = $1 AND user_id = $2`
	result, err := d.db.Exec(query, tunnelID, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("tunnel not found or access denied")
	}
	return nil
}

func (d *Database) IsPortAvailable(port int) bool {
	query := `SELECT COUNT(*) FROM tunnels WHERE public_port = $1`
	var count int
	err := d.db.QueryRow(query, port).Scan(&count)
	if err != nil {
		return false
	}
	return count == 0
}

func (d *Database) CreateConnection(conn *models.Connection) error {
	query := `
		INSERT INTO connections (tunnel_id, remote_addr, connected_at, bytes_up, bytes_down, duration)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	return d.db.QueryRow(query,
		conn.TunnelID, conn.RemoteAddr, conn.ConnectedAt,
		conn.BytesUp, conn.BytesDown, conn.Duration,
	).Scan(&conn.ID)
}

func (d *Database) UpdateConnection(conn *models.Connection) error {
	query := `
		UPDATE connections 
		SET disconnected_at = $1, bytes_up = $2, bytes_down = $3, duration = $4
		WHERE id = $5
	`
	
	_, err := d.db.Exec(query,
		conn.DisconnectedAt, conn.BytesUp, conn.BytesDown, conn.Duration, conn.ID,
	)
	
	return err
}

func (d *Database) GetMetrics() (*models.Metrics, error) {
	metrics := &models.Metrics{}
	
	// Active tunnels
	err := d.db.QueryRow(`SELECT COUNT(*) FROM tunnels WHERE status = 'active'`).Scan(&metrics.ActiveTunnels)
	if err != nil {
		return nil, err
	}
	
	// Total connections
	err = d.db.QueryRow(`SELECT COUNT(*) FROM connections`).Scan(&metrics.TotalConnections)
	if err != nil {
		return nil, err
	}
	
	// Total bytes up
	err = d.db.QueryRow(`SELECT COALESCE(SUM(bytes_up), 0) FROM connections`).Scan(&metrics.TotalBytesUp)
	if err != nil {
		return nil, err
	}
	
	// Total bytes down
	err = d.db.QueryRow(`SELECT COALESCE(SUM(bytes_down), 0) FROM connections`).Scan(&metrics.TotalBytesDown)
	if err != nil {
		return nil, err
	}
	
	// Active users (users with active tunnels in last hour)
	err = d.db.QueryRow(`
		SELECT COUNT(DISTINCT t.user_id) 
		FROM tunnels t 
		WHERE t.status = 'active' AND t.last_seen > NOW() - INTERVAL '1 hour'
	`).Scan(&metrics.ActiveUsers)
	if err != nil {
		return nil, err
	}
	
	return metrics, nil
}

func (d *Database) GetTunnelStats(tunnelID string) (*models.TunnelStats, error) {
	stats := &models.TunnelStats{}
	
	query := `
		SELECT 
			COUNT(*) as connections,
			COALESCE(SUM(bytes_up), 0) as bytes_up,
			COALESCE(SUM(bytes_down), 0) as bytes_down,
			COALESCE(MAX(connected_at), created_at) as last_active
		FROM connections c
		LEFT JOIN tunnels t ON c.tunnel_id = t.id
		WHERE c.tunnel_id = $1
	`
	
	err := d.db.QueryRow(query, tunnelID).Scan(
		&stats.Connections, &stats.BytesUp, &stats.BytesDown, &stats.LastActive,
	)
	
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}
