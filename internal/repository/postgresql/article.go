package postgresql

import (
	"context"
	"errors"
	"github.com/NRKA/gRPC-Server/internal/repository"
	"github.com/jackc/pgx/v5"
)

type ArticleRepo struct {
	db repository.DataBaseInterface
}

func NewArticleRepo(database repository.DataBaseInterface) *ArticleRepo {
	return &ArticleRepo{db: database}
}

func (r *ArticleRepo) Create(ctx context.Context, article repository.Article) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO articles(name,rating) VALUES($1,$2) RETURNING id;`,
		article.Name, article.Rating).Scan(&id)
	return id, err
}

func (r *ArticleRepo) GetByID(ctx context.Context, id int64) (repository.Article, error) {
	var article repository.Article
	err := r.db.Get(ctx, &article, "SELECT id,name,rating,created_at FROM articles WHERE id=$1", id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return article, repository.ErrArticalNotFound
		}
		return article, err
	}
	return article, nil
}
func (r *ArticleRepo) Delete(ctx context.Context, id int64) error {
	commandTag, err := r.db.Exec(ctx, "DELETE FROM articles WHERE id=$1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrArticalNotFound
	}
	return err
}

func (r *ArticleRepo) Update(ctx context.Context, article repository.Article) error {
	commandTag, err := r.db.Exec(ctx, "UPDATE articles SET name=$1, rating=$2 WHERE id=$3",
		article.Name, article.Rating, article.ID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return repository.ErrArticalNotFound
	}
	return err
}
