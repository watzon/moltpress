package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/watzon/moltpress/internal/users"
)

type mockUserRepo struct {
	users map[string]*users.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*users.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, req users.CreateUserRequest) (*users.CreateResult, error) {
	id := uuid.New()
	apiKey := "mp_test_" + id.String()[:8]
	verificationCode := "MP-testcode123456"

	user := &users.User{
		ID:               id,
		Username:         req.Username,
		DisplayName:      req.DisplayName,
		IsAgent:          req.IsAgent,
		APIKey:           &apiKey,
		VerificationCode: &verificationCode,
	}
	m.users[req.Username] = user

	return &users.CreateResult{
		User:             user,
		APIKey:           apiKey,
		VerificationCode: verificationCode,
	}, nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*users.User, error) {
	if user, ok := m.users[username]; ok {
		return user, nil
	}
	return nil, users.ErrUserNotFound
}

func (m *mockUserRepo) GetByAPIKey(ctx context.Context, apiKey string) (*users.User, error) {
	for _, user := range m.users {
		if user.APIKey != nil && *user.APIKey == apiKey {
			return user, nil
		}
	}
	return nil, users.ErrUserNotFound
}

type testServer struct {
	*httptest.Server
	userRepo *mockUserRepo
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	userRepo := newMockUserRepo()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		var req users.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Username == "" {
			writeError(w, http.StatusBadRequest, "username is required")
			return
		}

		result, err := userRepo.Create(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		resp := users.RegisterResponse{
			User:   result.User.ToPublic(),
			APIKey: result.APIKey,
		}

		if req.IsAgent && result.VerificationCode != "" {
			resp.VerificationCode = result.VerificationCode
			resp.VerificationURL = "https://x.com/intent/tweet?text=Verifying%20" + result.VerificationCode
		}

		writeJSON(w, http.StatusCreated, resp)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return &testServer{
		Server:   server,
		userRepo: userRepo,
	}
}

func makeRequest(t *testing.T, server *testServer, method, path string, body interface{}) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return resp
}

func parseResponse[T any](t *testing.T, resp *http.Response) T {
	t.Helper()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}
