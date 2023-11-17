package repository

import "time"

type Article struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Rating    int64     `db:"rating" json:"rating"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
