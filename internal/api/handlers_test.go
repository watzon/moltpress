package api

import (
	"net/http"
	"testing"
)

func TestHandleRegister_Agent(t *testing.T) {
	server := setupTestServer(t)

	resp := makeRequest(t, server, "POST", "/api/v1/register", map[string]interface{}{
		"username":     "test-agent",
		"display_name": "Test Agent",
		"is_agent":     true,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	result := parseResponse[map[string]interface{}](t, resp)

	if result["api_key"] == nil || result["api_key"] == "" {
		t.Error("expected api_key in response")
	}

	if result["verification_code"] == nil || result["verification_code"] == "" {
		t.Error("expected verification_code in response for agent")
	}

	if result["verification_url"] == nil || result["verification_url"] == "" {
		t.Error("expected verification_url in response for agent")
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user object in response")
	}

	if user["username"] != "test-agent" {
		t.Errorf("expected username 'test-agent', got %v", user["username"])
	}

	if user["is_agent"] != true {
		t.Errorf("expected is_agent true, got %v", user["is_agent"])
	}
}

func TestHandleRegister_Human(t *testing.T) {
	server := setupTestServer(t)

	resp := makeRequest(t, server, "POST", "/api/v1/register", map[string]interface{}{
		"username":     "test-human",
		"display_name": "Test Human",
		"is_agent":     false,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	result := parseResponse[map[string]interface{}](t, resp)

	if result["api_key"] == nil || result["api_key"] == "" {
		t.Error("expected api_key in response")
	}

	if result["verification_code"] != nil && result["verification_code"] != "" {
		t.Error("did not expect verification_code for human")
	}
}

func TestHandleRegister_MissingUsername(t *testing.T) {
	server := setupTestServer(t)

	resp := makeRequest(t, server, "POST", "/api/v1/register", map[string]interface{}{
		"display_name": "No Username",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
