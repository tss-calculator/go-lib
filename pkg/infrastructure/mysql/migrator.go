package mysql

import (
	"context"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const migratorTimeout = 5 * time.Second

type Migrator interface {
	MigrateUp() error
}

func NewMigrator(ctx context.Context, migrationsDir string, factory LockableUnitOfWorkFactory) Migrator {
	return &migrator{
		ctx:           ctx,
		migrationsDir: migrationsDir,
		factory:       factory,
	}
}

type migrator struct {
	ctx           context.Context
	migrationsDir string
	factory       LockableUnitOfWorkFactory
}

func (m *migrator) MigrateUp() error {
	return m.executeUnitOfWork(m.ctx, func(client ClientContext) error {
		err := m.createMigrationVersionsTable(client)
		if err != nil {
			return err
		}

		migrations, err := m.listMigrations()
		if err != nil {
			return err
		}

		executedMigrations, err := m.listExecutedMigrationVersions(client)
		if err != nil {
			return err
		}

		for _, migration := range migrations {
			if executedMigrations[migration.Version] {
				continue
			}

			err = m.executeMigration(client, migration)
			if err != nil {
				return err
			}

			err = m.saveMigrationVersion(client, migration.Version)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *migrator) createMigrationVersionsTable(client ClientContext) error {
	const sqlQuery = `CREATE TABLE IF NOT EXISTS migration_versions (version VARCHAR(50) NOT NULL)`
	_, err := client.ExecContext(m.ctx, sqlQuery)
	return errors.WithStack(err)
}

func (m *migrator) listMigrations() ([]migrationInfo, error) {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]migrationInfo, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
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

func (m *migrator) getVersionFromFilename(fileName string) string {
	return strings.Split(fileName, "_")[0]
}

func (m *migrator) listExecutedMigrationVersions(client ClientContext) (map[string]bool, error) {
	const sqlQuery = `SELECT version FROM migration_versions`

	var versions []string
	err := client.SelectContext(m.ctx, &versions, sqlQuery)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make(map[string]bool)
	for _, version := range versions {
		result[version] = true
	}

	return result, nil
}

func (m *migrator) executeMigration(client ClientContext, migration migrationInfo) error {
	content, err := os.ReadFile(migration.FilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = client.ExecContext(m.ctx, string(content))
	return errors.WithStack(err)
}

func (m *migrator) saveMigrationVersion(client ClientContext, version string) error {
	const sqlQuery = `INSERT INTO migration_versions SET version = ?`
	_, err := client.ExecContext(m.ctx, sqlQuery, version)
	return errors.WithStack(err)
}

func (m *migrator) executeUnitOfWork(ctx context.Context, f func(ClientContext) error) error {
	unitOfWork, err := m.factory.NewLockableUnitOfWork(ctx, "", migratorTimeout)
	if err != nil {
		return err
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()

	err = f(unitOfWork.ClientContext())
	return err
}

type migrationInfo struct {
	Version  string
	FilePath string
}
