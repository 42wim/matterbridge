package multidevice

import "database/sql"

type sqlitePersistence struct {
	db *sql.DB
}

func newSQLitePersistence(db *sql.DB) *sqlitePersistence {
	return &sqlitePersistence{db: db}
}

// GetActiveInstallations returns the active installations for a given identity
func (s *sqlitePersistence) GetActiveInstallations(maxInstallations int, identity []byte) ([]*Installation, error) {
	stmt, err := s.db.Prepare(`SELECT installation_id, version
				   FROM installations
				   WHERE enabled = 1 AND identity = ?
				   ORDER BY timestamp DESC
				   LIMIT ?`)
	if err != nil {
		return nil, err
	}

	var installations []*Installation
	rows, err := stmt.Query(identity, maxInstallations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			installationID string
			version        uint32
		)
		err = rows.Scan(
			&installationID,
			&version,
		)
		if err != nil {
			return nil, err
		}
		installations = append(installations, &Installation{
			ID:      installationID,
			Version: version,
			Enabled: true,
		})

	}

	return installations, nil

}

// GetInstallations returns all the installations for a given identity
// we both return the installations & the metadata
// metadata is currently stored in a separate table, as in some cases we
// might have metadata for a device, but no other information on the device
func (s *sqlitePersistence) GetInstallations(identity []byte) ([]*Installation, error) {
	installationMap := make(map[string]*Installation)
	var installations []*Installation

	// We query both tables as sqlite does not support full outer joins
	installationsStmt, err := s.db.Prepare(`SELECT installation_id, version, enabled, timestamp FROM installations WHERE identity = ?`)
	if err != nil {
		return nil, err
	}
	defer installationsStmt.Close()

	installationRows, err := installationsStmt.Query(identity)
	if err != nil {
		return nil, err
	}

	for installationRows.Next() {
		var installation Installation
		err = installationRows.Scan(
			&installation.ID,
			&installation.Version,
			&installation.Enabled,
			&installation.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		// We initialized to empty in this case as we want to
		// return metadata as well in this endpoint, but not in others
		installation.InstallationMetadata = &InstallationMetadata{}
		installationMap[installation.ID] = &installation
	}

	metadataStmt, err := s.db.Prepare(`SELECT installation_id, name, device_type, fcm_token FROM installation_metadata WHERE identity = ?`)
	if err != nil {
		return nil, err
	}
	defer metadataStmt.Close()

	metadataRows, err := metadataStmt.Query(identity)
	if err != nil {
		return nil, err
	}

	for metadataRows.Next() {
		var (
			installationID string
			name           sql.NullString
			deviceType     sql.NullString
			fcmToken       sql.NullString
			installation   *Installation
		)
		err = metadataRows.Scan(
			&installationID,
			&name,
			&deviceType,
			&fcmToken,
		)
		if err != nil {
			return nil, err
		}
		if _, ok := installationMap[installationID]; ok {
			installation = installationMap[installationID]
		} else {
			installation = &Installation{ID: installationID}
		}
		installation.InstallationMetadata = &InstallationMetadata{
			Name:       name.String,
			DeviceType: deviceType.String,
			FCMToken:   fcmToken.String,
		}
		installationMap[installationID] = installation
	}

	for _, installation := range installationMap {
		installations = append(installations, installation)
	}

	return installations, nil
}

// AddInstallations adds the installations for a given identity, maintaining the enabled flag
func (s *sqlitePersistence) AddInstallations(identity []byte, timestamp int64, installations []*Installation, defaultEnabled bool) ([]*Installation, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	var insertedInstallations []*Installation

	for _, installation := range installations {
		stmt, err := tx.Prepare(`SELECT enabled, version
					 FROM installations
					 WHERE identity = ? AND installation_id = ?
					 LIMIT 1`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		var oldEnabled bool
		// We don't override version once we saw one
		var oldVersion uint32
		latestVersion := installation.Version

		err = stmt.QueryRow(identity, installation.ID).Scan(&oldEnabled, &oldVersion)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if err == sql.ErrNoRows {
			stmt, err = tx.Prepare(`INSERT INTO installations(identity, installation_id, timestamp, enabled, version)
						VALUES (?, ?, ?, ?, ?)`)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()

			_, err = stmt.Exec(
				identity,
				installation.ID,
				timestamp,
				defaultEnabled,
				latestVersion,
			)
			if err != nil {
				return nil, err
			}
			insertedInstallations = append(insertedInstallations, installation)
		} else {
			// We update timestamp if present without changing enabled, only if this is a new bundle
			// and we set the version to the latest we ever saw
			if oldVersion > installation.Version {
				latestVersion = oldVersion
			}

			stmt, err = tx.Prepare(`UPDATE installations
					        SET timestamp = ?,  enabled = ?, version = ?
						WHERE identity = ?
						AND installation_id = ?
						AND timestamp < ?`)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()

			_, err = stmt.Exec(
				timestamp,
				oldEnabled,
				latestVersion,
				identity,
				installation.ID,
				timestamp,
			)
			if err != nil {
				return nil, err
			}
		}

	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return insertedInstallations, nil

}

// EnableInstallation enables the installation
func (s *sqlitePersistence) EnableInstallation(identity []byte, installationID string) error {
	stmt, err := s.db.Prepare(`UPDATE installations
				   SET enabled = 1
				   WHERE identity = ? AND installation_id = ?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(identity, installationID)
	return err

}

// DisableInstallation disable the installation
func (s *sqlitePersistence) DisableInstallation(identity []byte, installationID string) error {
	stmt, err := s.db.Prepare(`UPDATE installations
				   SET enabled = 0
				   WHERE identity = ? AND installation_id = ?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(identity, installationID)
	return err
}

// SetInstallationMetadata sets the metadata for a given installation
func (s *sqlitePersistence) SetInstallationMetadata(identity []byte, installationID string, metadata *InstallationMetadata) error {
	stmt, err := s.db.Prepare(`INSERT INTO installation_metadata(name, device_type, fcm_token, identity, installation_id) VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(metadata.Name, metadata.DeviceType, metadata.FCMToken, identity, installationID)
	return err
}

// SetInstallationName sets the only the name in metadata for a given installation
func (s *sqlitePersistence) SetInstallationName(identity []byte, installationID string, name string) error {
	stmt, err := s.db.Prepare(`UPDATE installation_metadata 
							   SET name = ? 
							   WHERE identity = ? AND installation_id = ?`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(name, identity, installationID)
	return err
}
