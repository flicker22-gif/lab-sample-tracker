package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migration 一个版本化迁移文件（文件名即版本，如 0001_init.sql）
type migration struct {
	version string
	name    string
	sql     string
}

// Migrate 执行版本化数据库迁移，可重复运行：
//   - schema_migrations 表记录已应用版本，已应用的跳过；
//   - 老库（由 GORM AutoMigrate 建表、无 schema_migrations）自动基线到 0001，
//     仅执行其后的迁移，避免重复建表；
//   - 每个迁移在独立事务中执行，失败即终止并返回错误。
func Migrate(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层连接失败: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return fmt.Errorf("未找到任何迁移文件")
	}

	if _, err := sqlDB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("创建 schema_migrations 失败: %w", err)
	}

	applied := map[string]bool{}
	rows, err := sqlDB.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("读取 schema_migrations 失败: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// 基线：老库已有业务表但无迁移记录时，将首个迁移标记为已应用
	if len(applied) == 0 {
		exists, err := tableExists(ctx, sqlDB, "users")
		if err != nil {
			return err
		}
		if exists {
			base := migrations[0]
			if _, err := sqlDB.ExecContext(ctx,
				`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, base.version, base.name); err != nil {
				return fmt.Errorf("写入基线迁移记录失败: %w", err)
			}
			applied[base.version] = true
			slog.Info("检测到既有库，已基线到初始迁移", "version", base.version)
		}
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		slog.Info("应用数据库迁移", "version", m.version, "name", m.name)
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("迁移 %s 执行失败: %w", m.version, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.version, m.name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("迁移 %s 记录失败: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("迁移 %s 提交失败: %w", m.version, err)
		}
	}
	return nil
}

// loadMigrations 读取内嵌迁移文件并按文件名排序
func loadMigrations() ([]migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	out := make([]migration, 0, len(files))
	for _, f := range files {
		b, err := migrationsFS.ReadFile("migrations/" + f)
		if err != nil {
			return nil, err
		}
		version := f
		if i := strings.Index(f, "_"); i > 0 {
			version = f[:i]
		}
		out = append(out, migration{
			version: version,
			name:    strings.TrimSuffix(f, ".sql"),
			sql:     string(b),
		})
	}
	return out, nil
}

func tableExists(ctx context.Context, sqlDB *sql.DB, table string) (bool, error) {
	var exists bool
	err := sqlDB.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`, table).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("检查表 %s 是否存在失败: %w", table, err)
	}
	return exists, nil
}
