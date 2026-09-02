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

// Client — клієнт OpenAI-сумісного API (aimlapi.com).
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// NewClient створює клієнта з типовим endpoint aimlapi.
func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "openai/gpt-5-5"
	}
	return &Client{
		BaseURL: "https://api.aimlapi.com/v1",
		APIKey:  apiKey,
		Model:   model,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// GenerateSQL перетворює запит українською в SELECT-запит діалектом Access SQL.
func (c *Client) GenerateSQL(ctx context.Context, question, schemaDesc string) (string, error) {
	system := "Ти — генератор SQL для MS Access (ACE/Jet SQL). " +
		"Відповідай ЛИШЕ одним SELECT-запитом, без пояснень і без markdown-огороджень. " +
		"Правила діалекту: дати у форматі #YYYY-MM-DD#, обмеження кількості — TOP n (не LIMIT), " +
		"ідентифікатори у [квадратних дужках], без крапки з комою в кінці. " +
		"Заборонено INSERT/UPDATE/DELETE/DROP/CREATE/ALTER та SELECT INTO. " +
		"Схема БД:\n" + schemaDesc

	out, err := c.ask(ctx, system, question)
	if err != nil {
		return "", err
	}
	return stripMarkdown(out), nil
}

// IsSafeSelect — політика безпеки: AI-запити тільки на читання.
func IsSafeSelect(q string) bool {
	t := strings.ToUpper(strings.TrimSpace(q))
	if !strings.HasPrefix(t, "SELECT") {
		return false
	}
	for _, bad := range []string{";", "INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "EXEC", "INTO", "GRANT"} {
		if strings.Contains(t, bad) {
			return false
		}
	}
	return true
}

func (c *Client) ask(ctx context.Context, system, user string) (string, error) {
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai api status %d: %s", resp.StatusCode, truncate(string(data), 200))
	}

	var parsed chatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("ai parse: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("ai api: порожня відповідь")
	}
	return parsed.Choices[0].Message.Content, nil
}

// stripMarkdown прибирає ```-огородження, які ШІ любить додавати навколо SQL.
func stripMarkdown(s string) string {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "```sql")
	t = strings.TrimPrefix(t, "```")
	t = strings.TrimSuffix(t, "```")
	return strings.TrimSpace(t)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
