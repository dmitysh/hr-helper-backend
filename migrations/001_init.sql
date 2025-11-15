-- +goose Up

CREATE TABLE candidate
(
    id                SERIAL PRIMARY KEY,
    telegram_id       BIGINT,
    telegram_username TEXT,
    full_name         TEXT,
    phone             TEXT,
    city              TEXT,
    created_at        TIMESTAMP WITH TIME ZONE default now()
);

CREATE INDEX candidate_telegram_id_idx ON candidate (telegram_id);

CREATE TABLE vacancy
(
    id               UUID PRIMARY KEY,
    title            TEXT,
    key_requirements TEXT[],
    created_at       TIMESTAMP WITH TIME ZONE default now()
);

CREATE TABLE candidate_vacancy_meta
(
    candidate_id    BIGINT REFERENCES candidate (id) ON DELETE CASCADE,
    vacancy_id      UUID REFERENCES vacancy (id) ON DELETE CASCADE,
    status          TEXT,
    interview_score SMALLINT,
    is_archived     BOOLEAN DEFAULT false,
    updated_at      TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX candidate_vacancy_meta_candidate_id_vacancy_id_unique_idx ON candidate_vacancy_meta (candidate_id, vacancy_id);

CREATE TABLE resume_screening
(
    id           SERIAL PRIMARY KEY,
    candidate_id BIGINT REFERENCES candidate (id) ON DELETE CASCADE,
    vacancy_id   UUID REFERENCES vacancy (id) ON DELETE CASCADE,
    score        SMALLINT,
    feedback     TEXT,
    created_at   TIMESTAMP WITH TIME ZONE default now(),
    updated_at   TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX resume_screening_candidate_id_vacancy_id_unique_idx ON resume_screening (candidate_id, vacancy_id);

CREATE TABLE question
(
    id         SERIAL PRIMARY KEY,
    vacancy_id UUID REFERENCES vacancy (id) ON DELETE CASCADE,
    content    TEXT,
    reference  TEXT,
    time_limit SMALLINT,
    "position" SMALLINT,
    created_at TIMESTAMP WITH TIME ZONE default now()
);

CREATE INDEX question_vacancy_id_idx ON question (vacancy_id);

CREATE TABLE answer
(
    id           SERIAL PRIMARY KEY,
    candidate_id BIGINT REFERENCES candidate (id) ON DELETE CASCADE,
    question_id  BIGINT REFERENCES question (id) ON DELETE CASCADE,
    content      TEXT,
    score        SMALLINT,
    time_taken   SMALLINT,
    created_at   TIMESTAMP WITH TIME ZONE default now()
);

-- +goose Down
DROP TABLE IF EXISTS resume_screening;
DROP TABLE IF EXISTS candidate_vacancy_meta;
DROP TABLE IF EXISTS answer;
DROP TABLE IF EXISTS question;
DROP TABLE IF EXISTS vacancy;
DROP TABLE IF EXISTS candidate;
