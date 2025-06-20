package transaction

import (
	"context"

	"github.com/zuhrulumam/go-hris/pkg"
	"gorm.io/gorm"
)

//go:generate mockgen -source=business/domain/transaction/transaction.go -destination=mocks/domain/transaction/mock_transaction.go -package=mocks
type DomainItf interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Option struct {
	DB *gorm.DB
}

type transaction struct {
	db *gorm.DB
}

func Init(opt Option) DomainItf {
	return &transaction{
		db: opt.DB,
	}
}

func (t *transaction) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := t.db.Begin()

	// Create new context with tx
	ctxWithTx := context.WithValue(ctx, pkg.TxCtxValue, tx)

	err := fn(ctxWithTx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
