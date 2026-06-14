package actual

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"who-owes-me/internal/envutil"
)

type Client struct {
	BaseURL  string
	APIKey   string
	BudgetID string
	HTTP     *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL:  envutil.Getenv("ACTUAL_SERVER_URL"),
		APIKey:   envutil.Getenv("ACTUAL_API_KEY"),
		BudgetID: envutil.Getenv("ACTUAL_BUDGET_ID"),
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	if c.BudgetID == "" {
		return nil, fmt.Errorf("ACTUAL_BUDGET_ID is missing from environment variables")
	}
	
	req, err := http.NewRequest(method, fmt.Sprintf("%s/v1/budgets/%s%s", c.BaseURL, c.BudgetID, endpoint), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("actual API returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

type Payee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetPayees() ([]Payee, error) {
	data, err := c.doRequest("GET", "/payees", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Payee `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

type Transaction struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
	Amount   int    `json:"amount"` // in cents
	Payee    string `json:"payee"`
	Notes    string `json:"notes"`
	Account  string `json:"account"`
	Category string `json:"category"`
}

func (c *Client) RunQuery(aqlQuery interface{}) ([]Transaction, error) {
	payload := map[string]interface{}{
		"ActualQLquery": aqlQuery,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest("POST", "/run-query", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Transaction `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) GetTransactionsByPayee(payeeID string) ([]Transaction, error) {
	query := map[string]interface{}{
		"table": "transactions",
		"select": []string{"*"},
		"filter": map[string]interface{}{
			"payee": payeeID,
		},
	}
	return c.RunQuery(query)
}

func (c *Client) GetTransactionsByTag(tag string) ([]Transaction, error) {
	query := map[string]interface{}{
		"table": "transactions",
		"select": []string{"*"},
		"filter": map[string]interface{}{
			"notes": map[string]interface{}{
				"$like": "%" + tag + "%",
			},
		},
	}
	return c.RunQuery(query)
}
