package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const apiBaseURL = "https://api.elevenlabs.io/v1/convai"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return resp, nil
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return resp, fmt.Errorf("API error: %s, status code: %d, body: %s", resp.Status, resp.StatusCode, string(bodyBytes))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

type Agent struct {
	AgentID            string             `json:"agent_id,omitempty"`
	Name               string             `json:"name,omitempty"`
	ConversationConfig ConversationConfig `json:"conversation_config"`
}

type ConversationConfig struct {
	Agent AgentConfig `json:"agent"`
}

type AgentConfig struct {
	FirstMessage string       `json:"first_message,omitempty"`
	Prompt       PromptConfig `json:"prompt"`
}

type PromptConfig struct {
	Prompt string `json:"prompt"`
	LLM    string `json:"llm,omitempty"`
}

func (c *Client) CreateAgent(ctx context.Context, agent *Agent) (*Agent, error) {
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("%s/agents/create", apiBaseURL), agent)
	if err != nil {
		return nil, err
	}

	var createdAgent Agent
	_, err = c.do(req, &createdAgent)
	if err != nil {
		return nil, err
	}
	return &createdAgent, nil
}

func (c *Client) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("%s/agents/%s", apiBaseURL, agentID), nil)
	if err != nil {
		return nil, err
	}

	var agent Agent
	resp, err := c.do(req, &agent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	agent.AgentID = agentID
	return &agent, nil
}

func (c *Client) UpdateAgent(ctx context.Context, agentID string, agent *Agent) error {
	req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("%s/agents/%s/update", apiBaseURL, agentID), agent)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

type Tool struct {
	ToolID      string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	APISchema   APISchema `json:"api_schema"`
}

type APISchema struct {
	URL    string `json:"url"`
	Method string `json:"method,omitempty"`
}

func (c *Client) CreateTool(ctx context.Context, tool *Tool) (*Tool, error) {
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("%s/tools", apiBaseURL), tool)
	if err != nil {
		return nil, err
	}

	var createdTool Tool
	_, err = c.do(req, &createdTool)
	if err != nil {
		return nil, err
	}
	return &createdTool, nil
}

func (c *Client) GetTool(ctx context.Context, toolID string) (*Tool, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("%s/tools/%s", apiBaseURL, toolID), nil)
	if err != nil {
		return nil, err
	}

	var tool Tool
	resp, err := c.do(req, &tool)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	tool.ToolID = toolID
	return &tool, nil
}

func (c *Client) UpdateTool(ctx context.Context, toolID string, tool *Tool) error {
	req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("%s/tools/%s", apiBaseURL, toolID), tool)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

func (c *Client) DeleteTool(ctx context.Context, toolID string) error {
	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("%s/tools/%s", apiBaseURL, toolID), nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

type KnowledgeBaseDocumentCreateResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type KnowledgeBaseDocumentFromURLRequest struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type,omitempty"`
}

type KnowledgeBaseDocumentFromTextRequest struct {
	Text     string `json:"text"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type,omitempty"`
}

type KnowledgeBaseDocumentDetails struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	SourceType        string `json:"source_type"`
	SourceURL         string `json:"source_url,omitempty"`
	SourceTextPreview string `json:"source_text_preview,omitempty"`
	SourceFileName    string `json:"source_file_name,omitempty"`
	MimeType          string `json:"mime_type"`
	SizeBytes         int    `json:"size_bytes"`
	CharacterCount    int    `json:"character_count"`
	Status            string `json:"status"`
	CreatedAtUnix     int64  `json:"created_at_unix"`
}

func (c *Client) CreateKnowledgeBaseDocumentFromURL(ctx context.Context, doc *KnowledgeBaseDocumentFromURLRequest) (*KnowledgeBaseDocumentCreateResponse, error) {
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("%s/knowledge-base/documents/create-from-url", apiBaseURL), doc)
	if err != nil {
		return nil, err
	}

	var createResp KnowledgeBaseDocumentCreateResponse
	_, err = c.do(req, &createResp)
	if err != nil {
		return nil, err
	}
	return &createResp, nil
}

func (c *Client) CreateKnowledgeBaseDocumentFromText(ctx context.Context, doc *KnowledgeBaseDocumentFromTextRequest) (*KnowledgeBaseDocumentCreateResponse, error) {
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("%s/knowledge-base/documents/create-from-text", apiBaseURL), doc)
	if err != nil {
		return nil, err
	}

	var createResp KnowledgeBaseDocumentCreateResponse
	_, err = c.do(req, &createResp)
	if err != nil {
		return nil, err
	}
	return &createResp, nil
}

func (c *Client) CreateKnowledgeBaseDocumentFromFile(ctx context.Context, name, mimeType, filePath string) (*KnowledgeBaseDocumentCreateResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	_ = writer.WriteField("name", name)
	if mimeType != "" {
		_ = writer.WriteField("mime_type", mimeType)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/knowledge-base/documents/create-from-file", apiBaseURL), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("xi-api-key", c.apiKey)

	var createResp KnowledgeBaseDocumentCreateResponse
	_, err = c.do(req, &createResp)
	if err != nil {
		return nil, err
	}
	return &createResp, nil
}

func (c *Client) GetKnowledgeBaseDocument(ctx context.Context, documentID string) (*KnowledgeBaseDocumentDetails, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("%s/knowledge-base/documents/%s", apiBaseURL, documentID), nil)
	if err != nil {
		return nil, err
	}

	var doc KnowledgeBaseDocumentDetails
	resp, err := c.do(req, &doc)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return &doc, nil
}

func (c *Client) DeleteKnowledgeBaseDocument(ctx context.Context, documentID string) error {
	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("%s/knowledge-base/documents/%s", apiBaseURL, documentID), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

func (c *Client) DeleteAgent(ctx context.Context, agentID string) error {
	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("%s/agents/%s/delete", apiBaseURL, agentID), nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}