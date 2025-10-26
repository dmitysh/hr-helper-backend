package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/multierr"

	"hr-helper/internal/adapter/llm"
	"hr-helper/internal/adapter/objstorage"
	"hr-helper/internal/adapter/repository"
	"hr-helper/internal/handler/httpapi"
	"hr-helper/internal/pkg/houston/closer"
	"hr-helper/internal/pkg/houston/config"
	"hr-helper/internal/pkg/houston/dobby"
	"hr-helper/internal/pkg/houston/loggy"
	"hr-helper/internal/pkg/houston/secret"
	"hr-helper/internal/service/candidate"
	"hr-helper/internal/service/interview"
	"hr-helper/internal/service/vacancy"
)

type App struct {
	cfg Config
}

func NewApp(config Config) *App {
	return &App{
		cfg: config,
	}
}

func (a *App) Run(ctx context.Context) {
	pgPool, err := dobby.NewPGXPool(ctx,
		dobby.PGXConfig{
			Username:      os.Getenv("POSTGRES_USER"),
			Password:      os.Getenv("POSTGRES_PASSWORD"),
			Host:          config.String("postgres.host"),
			Port:          config.String("postgres.port"),
			DBName:        config.String("postgres.db"),
			SSLMode:       config.String("postgres.ssl"),
			TLSServerName: config.String("postgres.tls_server_name"),
		})
	if err != nil {
		loggy.Fatalf("can't init pg pool: %v", err)
	}

	minioClient, err := minio.New(
		config.String("s3.endpoint"),
		&minio.Options{
			Creds: credentials.NewStaticV4(
				secret.GetString("S3_ACCESS_KEY"),
				secret.GetString("S3_SECRET_KEY"),
				"",
			),
			Secure: config.Bool("s3.secure"),
			Region: "ru-central1",
		},
	)
	if err != nil {
		loggy.Fatalf("can't init minio client: %v", err)
	}

	_, err = minioClient.ListBuckets(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	resumeStorage := objstorage.NewResumeStorage(config.String("s3.resume_bucket"), minioClient)
	candidateStorage := repository.NewCandidateRepository(pgPool)
	interviewStorage := repository.NewInterviewRepository(pgPool)
	vacancyStorage := repository.NewVacancyRepository(pgPool)

	yandexLLM := llm.NewYandex(llm.YandexConfig{
		APIKey:   secret.GetString("YANDEX_LLM_API_KEY"),
		FolderID: secret.GetString("YANDEX_FOLDER_ID"),
	})

	candidateService := candidate.NewService(candidateStorage, resumeStorage, vacancyStorage, yandexLLM)
	interviewService := interview.NewService(interviewStorage, yandexLLM)
	vacancyService := vacancy.NewService(vacancyStorage)

	srv := httpapi.NewServer(
		config.String("http.addr"),
		candidateService,
		interviewService,
		vacancyService,
	)
	a.runHTTPServer(srv)

	closer.Add(func() error {
		var err error
		err = multierr.Append(err, srv.Shutdown(ctx))

		if err != nil {
			return err
		}
		return err
	})

	closer.Wait()
}

func (a *App) runHTTPServer(srv *httpapi.Server) {
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggy.Fatalf("http server error: %v", err)
		}
	}()
}
