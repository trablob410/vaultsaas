package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackAdapter sends messages via Slack Bot API (no SDK — minimal deps).
type SlackAdapter struct {
	botToken string
	client   *http.Client
}

// NewSlackAdapter creates a SlackAdapter. Returns nil if botToken is empty.
func NewSlackAdapter(botToken string) *SlackAdapter {
	if botToken == "" {
		return nil
	}
	return &SlackAdapter{botToken: botToken, client: &http.Client{}}
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackElement struct {
	Type     string     `json:"type"`
	Text     *slackText `json:"text"`
	ActionID string     `json:"action_id"`
	Style    string     `json:"style,omitempty"` // "primary" | "danger"
}

type slackBlock struct {
	Type     string         `json:"type"`
	Text     *slackText     `json:"text,omitempty"`
	Elements []slackElement `json:"elements,omitempty"`
}

// SendApprovalRequest sends a Block Kit DM with Approve/Reject buttons.
func (s *SlackAdapter) SendApprovalRequest(ctx context.Context, slackUserID, requestID, secretName, requester, reason string) error {
	payload := map[string]interface{}{
		"channel": slackUserID,
		"text":    fmt.Sprintf("Approval needed: %s requests access to %s", requester, secretName),
		"blocks": []slackBlock{
			{
				Type: "section",
				Text: &slackText{Type: "mrkdwn", Text: fmt.Sprintf(
					"*Access Request*\n*Secret:* %s\n*Requester:* %s\n*Reason:* %s",
					secretName, requester, reason,
				)},
			},
			{
				Type: "actions",
				Elements: []slackElement{
					{
						Type:     "button",
						ActionID: "approve:" + requestID,
						Text:     &slackText{Type: "plain_text", Text: "✓ Approve"},
						Style:    "primary",
					},
					{
						Type:     "button",
						ActionID: "reject:" + requestID,
						Text:     &slackText{Type: "plain_text", Text: "✗ Reject"},
						Style:    "danger",
					},
				},
			},
		},
	}
	return s.post(ctx, "https://slack.com/api/chat.postMessage", payload)
}

// UpdateMessage replaces the approval buttons with an outcome message.
func (s *SlackAdapter) UpdateMessage(ctx context.Context, channelID, ts, outcome string) error {
	payload := map[string]interface{}{
		"channel": channelID,
		"ts":      ts,
		"text":    "Request " + outcome + ".",
		"blocks": []slackBlock{
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: "*Request " + outcome + ".* (actioned via Slack)"}},
		},
	}
	return s.post(ctx, "https://slack.com/api/chat.update", payload)
}

func (s *SlackAdapter) post(ctx context.Context, url string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack api: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}
