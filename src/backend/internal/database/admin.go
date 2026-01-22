package database

import "errors"

// DeleteUser xóa user (admin only)
func (d *Database) DeleteUser(userID string) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := d.db.Exec(query, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// DeleteTunnelByAdmin xóa tunnel bất kỳ (admin only)
func (d *Database) DeleteTunnelByAdmin(tunnelID string) error {
	query := `DELETE FROM tunnels WHERE id = ?`
	result, err := d.db.Exec(query, tunnelID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("tunnel not found")
	}

	return nil
}
