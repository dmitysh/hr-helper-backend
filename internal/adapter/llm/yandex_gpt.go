package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"

	"hr-helper/internal/entity"
	"hr-helper/internal/service_models"
)

const completionURL = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"

type CompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int32   `json:"maxTokens,omitempty"`
}

type Message struct {
	Role string `json:"role"` // "system" | "user" | "assistant"
	Text string `json:"text"`
}

type CompletionRequest struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions CompletionOptions `json:"completionOptions"`
	Messages          []Message         `json:"messages"`
}

type Alternative struct {
	Message Message `json:"message"`
}

type Result struct {
	Alternatives []Alternative `json:"alternatives"`
}

type CompletionResponse struct {
	Result Result `json:"result"`
}

type YandexConfig struct {
	APIKey   string
	FolderID string
}

const (
	baseScoreResumePrompt = `Оцени резюме кандидата, проходящего на вакансию %s: опиши кандидата в общем, и дай ему оценку по 100-бальной шкале, 
где 100 - означает отличный кандидат подходящий идеально, 0 - кандидат не подходит под большинство критериев. Подойди к оценке комплексно.
Самое важное - это учесть в оценке требуемые для вакансии навыки и качества кандидата, вот их список: %s.
Твой ответ обязательно должен представлять собой JSON с двумя полями: {\"feedback\": \"<общее_описание, string>\", \"score\": \"<оценка, int>\"}.
Резюме кандидата: %s`

	baseScoreQuestionPrompt = `Оцени ответ кандидата: дай ему оценку по 100-бальной шкале, 
где 100 - означает отличный ответ, полностью соответствующий референсному ответу, 0 - крайне плохой ответ, не соответсвующий ни референсу, ни действительности. Подойди к оценке комплексно.
Твой ответ обязательно должен представлять собой JSON с одним полями: {\"score\": \"<оценка, int>\"}.
Ответ кандидата: %s, референсный ответ: %s`
)

type Yandex struct {
	cfg    YandexConfig
	client *http.Client
}

func NewYandex(cfg YandexConfig) *Yandex {
	return &Yandex{
		client: &http.Client{},
		cfg:    cfg,
	}
}

func (y *Yandex) ScoreResume(ctx context.Context, resumeText string, vacancy entity.Vacancy) (service_models.ResumeScreeningResult, error) {
	var res service_models.ResumeScreeningResult

	msgs := []Message{
		{Role: "system", Text: "Ты HR-специалист, проводящий скрининг резюме кандидатов"},
		{Role: "user", Text: fmt.Sprintf(baseScoreResumePrompt, vacancy.Title, strings.Join(vacancy.KeyRequirements, ","), resumeText)},
	}

	err := retry.Do(
		func() error {
			resp, err := y.doRequest(ctx, msgs)
			if err != nil {
				return fmt.Errorf("can't do llm request: %w", err)
			}

			resp = strings.Trim(resp, "`\n")
			err = json.Unmarshal([]byte(resp), &res)
			if err != nil {
				return fmt.Errorf("can't unmarshal result: %w", err)
			}

			return nil
		},
		retry.Attempts(5),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(time.Second*1),
	)
	if err != nil {
		return service_models.ResumeScreeningResult{}, fmt.Errorf("can't score resume: %w", err)
	}

	return res, nil
}

func (y *Yandex) ScoreAnswer(ctx context.Context, answer string, reference string) (service_models.AnswerScoringResult, error) {
	var res service_models.AnswerScoringResult

	err := retry.Do(
		func() error {
			resp, err := y.doRequest(ctx, []Message{
				{Role: "system", Text: "Ты специалист, проводящий скрининг ответов кандидатов"},
				{Role: "user", Text: fmt.Sprintf(baseScoreQuestionPrompt, answer, reference)},
			})
			if err != nil {
				return fmt.Errorf("can't do llm request: %w", err)
			}

			resp = strings.Trim(resp, "`\n")

			err = json.Unmarshal([]byte(resp), &res)
			if err != nil {
				return fmt.Errorf("can't unmarshal result: %w", err)
			}

			return nil
		},
		retry.Attempts(5),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(time.Second*1),
	)
	if err != nil {
		return service_models.AnswerScoringResult{}, fmt.Errorf("can't score resume: %w", err)
	}

	return res, nil
}

func (y *Yandex) doRequest(ctx context.Context, messages []Message) (string, error) {
	modelURI := fmt.Sprintf("gpt://%s/yandexgpt/latest", y.cfg.FolderID)

	reqBody := CompletionRequest{
		ModelURI: modelURI,
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.2,
			MaxTokens:   20_000,
		},
		Messages: messages,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("can't marshal json: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completionURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("can't create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+y.cfg.APIKey)
	req.Header.Set("x-folder-id", y.cfg.FolderID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("can't do req: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("non-2xx status: %s\nbody: %s\n", resp.Status, buf.String())
	}

	var comp CompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&comp)
	if err != nil {
		return "", fmt.Errorf("can't decode resp: %w", err)
	}

	if len(comp.Result.Alternatives) == 0 {
		return "", fmt.Errorf("no alternatives found")
	}

	return comp.Result.Alternatives[0].Message.Text, nil
}
