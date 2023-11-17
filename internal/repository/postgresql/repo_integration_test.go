//go:build integration

package postgresql

import (
	"context"
	"github.com/NRKA/gRPC-Server/internal/db/postgres"
	"github.com/NRKA/gRPC-Server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateArticle(t *testing.T) {
	dbConnection := postgres.NewFromEnv()
	defer dbConnection.DB.GetPool().Close()
	testCases := []struct {
		name        string
		article     repository.Article
		expectError error
	}{{
		name: "Success",
		article: repository.Article{
			Name:   "someName",
			Rating: 10,
		},
		expectError: nil,
	},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dbConnection.SetUp(t)
			defer dbConnection.TearDown()

			// Arrange
			repo := NewArticleRepo(dbConnection.DB)

			// Act
			respCreate, err := repo.Create(context.Background(), tc.article)

			// Assert
			require.NoError(t, err)
			respGetByID, err := repo.GetByID(context.Background(), respCreate)
			require.NoError(t, err)
			assert.Equal(t, tc.article.Name, respGetByID.Name)
			assert.Equal(t, tc.article.Rating, respGetByID.Rating)
			assert.NotZero(t, respGetByID)
		})
	}
}

func TestGetArticle(t *testing.T) {
	dbConnection := postgres.NewFromEnv()
	defer dbConnection.DB.GetPool().Close()

	ctx := context.Background()

	testCases := []struct {
		name        string
		article     repository.Article
		expectError error
	}{
		{
			name: "Success",
			article: repository.Article{
				Name:   "someName",
				Rating: int64(22),
			},
			expectError: nil,
		},
		{
			name: "Fail, article not found",
			article: repository.Article{
				ID: 123123123,
			},
			expectError: repository.ErrArticalNotFound,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dbConnection.SetUp(t)
			defer dbConnection.TearDown()

			//arrange
			repo := NewArticleRepo(dbConnection.DB)
			if tc.expectError != nil {
				//act
				respGetByID, err := repo.GetByID(ctx, tc.article.ID)

				//assert
				assert.ErrorIs(t, err, tc.expectError)
				assert.Equal(t, tc.article.Name, respGetByID.Name)
				assert.Equal(t, tc.article.Rating, respGetByID.Rating)
				assert.Zero(t, respGetByID.ID)
				return
			}
			respCreate, err := repo.Create(ctx, tc.article)

			require.NoError(t, err)

			//act
			respGetByID, err := repo.GetByID(ctx, respCreate)

			//assert
			require.NoError(t, err)
			assert.Equal(t, tc.article.Name, respGetByID.Name)
			assert.Equal(t, tc.article.Rating, respGetByID.Rating)
			assert.NotZero(t, respGetByID.ID)
		})
	}
}

func TestUpdateArticle(t *testing.T) {
	dbConnection := postgres.NewFromEnv()
	defer dbConnection.DB.GetPool().Close()

	ctx := context.Background()
	testCases := []struct {
		name        string
		article     repository.Article
		expectedErr error
	}{{
		name: "Success",
		article: repository.Article{
			Name:   "NewName",
			Rating: 120,
		},
		expectedErr: nil,
	}, {
		name: "article not found",
		article: repository.Article{
			ID:     123123,
			Name:   "",
			Rating: 0,
		},
		expectedErr: repository.ErrArticalNotFound,
	},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dbConnection.SetUp(t)
			defer dbConnection.TearDown()

			//arrange
			repo := NewArticleRepo(dbConnection.DB)

			if tc.expectedErr != nil {
				//act
				err := repo.Update(ctx, tc.article)

				//assert
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.article.Name, repository.Article{}.Name)
				assert.Equal(t, tc.article.Rating, repository.Article{}.Rating)
				return
			}
			respCreate, err := repo.Create(ctx, tc.article)
			require.NoError(t, err)

			//act
			tc.article.ID = respCreate
			err = repo.Update(ctx, tc.article)
			require.NoError(t, err)

			//assert
			respGetByID, err := repo.GetByID(ctx, tc.article.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.article.Name, respGetByID.Name)
			assert.Equal(t, tc.article.Rating, respGetByID.Rating)
			assert.Equal(t, tc.article.ID, respGetByID.ID)
		})
	}
}

func TestDeleteArticle(t *testing.T) {
	dbConnection := postgres.NewFromEnv()
	defer dbConnection.DB.GetPool().Close()

	ctx := context.Background()
	id := int64(12313)

	testCases := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "success",
			expectedErr: nil,
		},
		{
			name:        "article not found",
			expectedErr: repository.ErrArticalNotFound,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dbConnection.SetUp(t)
			defer dbConnection.TearDown()
			repo := NewArticleRepo(dbConnection.DB)

			if tc.expectedErr != nil {
				err := repo.Delete(ctx, id)

				//assert
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}
			respCreate, err := repo.Create(ctx, repository.Article{
				Name:   "Name",
				Rating: 22,
			})

			require.NoError(t, err)
			assert.NotZero(t, respCreate)

			//act
			err = repo.Delete(ctx, respCreate)
			assert.ErrorIs(t, err, tc.expectedErr)

			//assert
			_, err = repo.GetByID(ctx, respCreate)

			assert.ErrorIs(t, err, repository.ErrArticalNotFound)
		})
	}
}
