package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Дефолти — GroqCloud (OpenAI-сумісний endpoint).
// Моделі за https://console.groq.com/docs/models
const (
	DefaultBaseURL = "https://api.groq.com/openai/v1"
	DefaultModel   = "llama-3.3-70b-versatile"
)

// Client — універсальний клієнт OpenAI-сумісного API (Groq, локальний Ollama тощо).
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// NewClient створює клієнта; порожні baseURL/model → дефолти Groq.
func NewClient(baseURL, apiKey, model string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: 75 * time.Second},
	}
}

// BaseURL/Model — гетери для діагностики.
func (c *Client) BaseURL() string { return c.baseURL }
func (c *Client) Model() string   { return c.model }

// Ping — швидка перевірка «ключ + endpoint живі» (GET /models).
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("недоступне: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return nil
}

// Message — повідомлення чату (role: system | user | assistant).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat виконує chat completion і повертає текст відповіді.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	body, _ := json.Marshal(chatRequest{Model: c.model, Messages: messages, Temperature: 0.2})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("API недоступне: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("невалідна відповідь API: %s", truncate(string(raw), 200))
	}
	if out.Error != nil {
		return "", fmt.Errorf("API: %s", out.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("API повернуло порожню відповідь")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

// GenerateSQL: запит українською + схема/БЗ → SQL для MS Access (лише SQL у відповіді).
func (c *Client) GenerateSQL(ctx context.Context, question, schema string) (string, error) {
	msgs := []Message{
		{Role: "system", Content: "Ти — генератор SQL для MS Access (ACE ODBC). Правила: лише один SELECT-запит; " +
			"імена таблиць і колонок у квадратних дужках; TOP замість LIMIT; дати у форматі #YYYY-MM-DD# або як параметри; " +
			"без markdown, без пояснень — у відповіді лише SQL."},
		{Role: "user", Content: "Схема та контекст:\n" + schema + "\n\nПитання: " + question + "\n\nSQL:"},
	}
	out, err := c.Chat(ctx, msgs)
	if err != nil {
		return "", err
	}
	return StripCodeFences(out), nil
}

// StripCodeFences прибирає markdown-огородження ```sql … ``` з відповіді моделі.
func StripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i >= 0 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	return strings.TrimSpace(s)
}

// IsSafeSelect пропускає лише одно-операторні SELECT/WITH-запити без небезпечних ключових слів.
func IsSafeSelect(q string) bool {
	t := strings.ToUpper(strings.TrimSpace(q))
	if !(strings.HasPrefix(t, "SELECT") || strings.HasPrefix(t, "WITH")) {
		return false
	}
	if strings.Contains(t, ";") {
		return false
	}
	for _, bad := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "TRUNCATE", "EXEC", "GRANT", "REVOKE", "ATTACH", "INTO"} {
		if strings.Contains(t, bad) {
			return false
		}
	}
	return true
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
