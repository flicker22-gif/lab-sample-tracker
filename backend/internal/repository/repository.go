// Package repository 是数据访问层：所有 GORM 查询都收敛在这里，
// service 层只通过 Repositories 访问数据库。多步写入通过 WithTx
// 在事务中执行，事务边界由 service 层显式控制。
package repository

import (
	"context"

	"gorm.io/gorm"
)

// Repositories 聚合全部领域仓库。通过 New 绑定一个 *gorm.DB
// （普通连接或事务），WithTx 会派生出绑定事务的副本。
type Repositories struct {
	Meta       *MetaRepository
	Samples    *SampleRepository
	Production *ProductionRepository
	Wafers     *WaferRepository

	db *gorm.DB
}

// New 以给定数据库句柄构建仓库集
func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Meta:       &MetaRepository{db: db},
		Samples:    &SampleRepository{db: db},
		Production: &ProductionRepository{db: db},
		Wafers:     &WaferRepository{db: db},
		db:         db,
	}
}

// WithTx 在单个数据库事务中执行 fn；fn 内拿到的仓库集绑定该事务。
// fn 返回 error 则整体回滚，否则提交。
func (r *Repositories) WithTx(ctx context.Context, fn func(tx *Repositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(New(tx))
	})
}
