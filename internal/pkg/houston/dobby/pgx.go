package dobby

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PGXConfig struct {
	Username      string
	Password      string
	Host          string
	Port          string
	DBName        string
	SSLMode       string
	TLSServerName string
}

func NewPGXPool(ctx context.Context, cfg PGXConfig) (*pgxpool.Pool, error) {
	rootCertPool := x509.NewCertPool()
	ca := ".postgresql/root.crt"
	pem, err := os.ReadFile(ca)
	if err != nil {
		return nil, fmt.Errorf("can't read ca file: %w", err)
	}

	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return nil, fmt.Errorf("can't append certs: %w", err)
	}

	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s target_session_attrs=read-write",
		cfg.Host, cfg.Port, cfg.DBName, cfg.Username, cfg.Password, cfg.SSLMode)

	connConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("can't parse config: %w", err)
	}

	connConfig.ConnConfig.TLSConfig = &tls.Config{
		RootCAs:    rootCertPool,
		ServerName: "c-c9q5761n664cbgcoc506.rw.mdb.yandexcloud.net",
	}

	db, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("can't init pgpool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't ping pg: %w", err)
	}

	return db, nil
}

func AutoMigratePostgres(dbPool *pgxpool.Pool, migrationsFolder string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("can't set postgres dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(dbPool)

	err := goose.Up(db, "migrations")
	if err != nil {
		return fmt.Errorf("can't open db from pgxpool: %w", err)
	}

	err = goose.Up(db, migrationsFolder)
	if err != nil {
		return fmt.Errorf("can't up migrations: %w", err)
	}

	return nil
}
