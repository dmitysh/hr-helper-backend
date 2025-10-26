package entity

import "time"

type Question struct {
	ID        int64     `db:"id"`
	VacancyID int64     `db:"vacancy_id"`
	Content   string    `db:"content"`
	Reference string    `db:"reference"`
	TimeLimit int       `db:"time_limit"`
	Position  int       `db:"position"`
	CreatedAt time.Time `db:"created_at"`
}
