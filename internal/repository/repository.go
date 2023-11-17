//go:generate mockgen -source ./repository.go -destination=./mocks/repository.go -package=mock_repository

package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArticleInterface interface {
	Create(ctx context.Context, article Article) (int64, error)
	GetByID(ctx context.Context, id int64) (Article, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, article Article) error
}
type DataBaseInterface interface {
	GetPool() *pgxpool.Pool
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}
