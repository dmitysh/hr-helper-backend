-- +goose Up

ALTER TABLE candidate_vacancy_meta
    ADD COLUMN interview_score smallint;

-- +goose Down
ALTER TABLE candidate_vacancy_meta DROP COLUMN interview_score;
