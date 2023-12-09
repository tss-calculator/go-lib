package mysql

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type Migrator interface {
	MigrateUp() error
}

func NewMigrator(ctx context.Context, migrationsDir string, client ClientContext) Migrator {
	return &migrator{
		ctx:           ctx,
		migrationsDir: migrationsDir,
		client:        client,
	}
}

type migrator struct {
	ctx           context.Context
	migrationsDir string
	client        ClientContext
}

func (m *migrator) MigrateUp() error {
	err := m.createMigrationVersionsTable()
	if err != nil {
		return err
	}

	migrations, err := m.listMigrations()
	if err != nil {
		return err
	}

	executedMigrations, err := m.listExecutedMigrationVersions()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if executedMigrations[migration.Version] {
			continue
		}

		err = m.executeMigration(migration)
		if err != nil {
			return errors.Wrapf(err, "error executing migration %s", migration.Version)
		}

		err = m.saveMigrationVersion(migration.Version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *migrator) createMigrationVersionsTable() error {
	const sqlQuery = `CREATE TABLE IF NOT EXISTS migration_versions (version VARCHAR(50) NOT NULL)`
	_, err := m.client.ExecContext(m.ctx, sqlQuery)
	return errors.WithStack(err)
}

func (m *migrator) listMigrations() ([]migrationInfo, error) {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]migrationInfo, 0, len(files))
	for _, file := range files {
		if !m.isUpMigrationFile(file) {
			continue
		}

		version := m.getVersionFromFilename(file.Name())
		filePath := path.Join(m.migrationsDir, file.Name())

		result = append(result, migrationInfo{
			Version:  version,
			FilePath: filePath,
		})
	}

	return result, nil
}

func (m *migrator) isUpMigrationFile(file os.DirEntry) bool {
	return !file.IsDir() && strings.Contains(file.Name(), ".up.")
}

func (m *migrator) getVersionFromFilename(fileName string) string {
	return strings.Split(fileName, "_")[0]
}

func (m *migrator) listExecutedMigrationVersions() (map[string]bool, error) {
	const sqlQuery = `SELECT version FROM migration_versions`

	var versions []string
	err := m.client.SelectContext(m.ctx, &versions, sqlQuery)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[string]bool)
	for _, version := range versions {
		result[version] = true
	}

	return result, nil
}

func (m *migrator) executeMigration(migration migrationInfo) error {
	content, err := os.ReadFile(migration.FilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = m.client.ExecContext(m.ctx, string(content))
	return errors.WithStack(err)
}

func (m *migrator) saveMigrationVersion(version string) error {
	const sqlQuery = `INSERT INTO migration_versions SET version = ?`
	_, err := m.client.ExecContext(m.ctx, sqlQuery, version)
	return errors.WithStack(err)
}

type migrationInfo struct {
	Version  string
	FilePath string
}
