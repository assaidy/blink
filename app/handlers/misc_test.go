package handlers

import (
	"encoding/json"
	"testing"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "encode chat cursor",
			input: getChatsCursor{LastMessageIDWithLastPartner: "msg_123"},
		},
		{
			name:  "encode message cursor",
			input: getChatMessagesCursor{LastMessageID: "msg_456"},
		},
		{
			name:  "encode profile cursor",
			input: searchProfilesCursor{LastUserID: "user_789"},
		},
		{
			name:  "encode empty struct",
			input: struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encodeCursor(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == "" {
				t.Error("expected non-empty result")
			}
			// Verify it's valid base64
			if len(result)%4 != 0 && len(result) < 4 {
				t.Errorf("result doesn't look like valid base64: %s", result)
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	chatCursor := getChatsCursor{LastMessageIDWithLastPartner: "msg_123"}
	encodedChatCursor, err := encodeCursor(chatCursor)
	if err != nil {
		t.Fatalf("failed to encode test data: %v", err)
	}

	messageCursor := getChatMessagesCursor{LastMessageID: "msg_456"}
	encodedMessageCursor, err := encodeCursor(messageCursor)
	if err != nil {
		t.Fatalf("failed to encode test data: %v", err)
	}

	profileCursor := searchProfilesCursor{LastUserID: "user_789"}
	encodedProfileCursor, err := encodeCursor(profileCursor)
	if err != nil {
		t.Fatalf("failed to encode test data: %v", err)
	}

	tests := []struct {
		name        string
		encoded     string
		target      any
		wantErr     bool
		checkResult func(t *testing.T, target any)
	}{
		{
			name:    "decode chat cursor",
			encoded: encodedChatCursor,
			target:  &getChatsCursor{},
			wantErr: false,
			checkResult: func(t *testing.T, target any) {
				c := target.(*getChatsCursor)
				if c.LastMessageIDWithLastPartner != "msg_123" {
					t.Errorf("expected msg_123, got %s", c.LastMessageIDWithLastPartner)
				}
			},
		},
		{
			name:    "decode message cursor",
			encoded: encodedMessageCursor,
			target:  &getChatMessagesCursor{},
			wantErr: false,
			checkResult: func(t *testing.T, target any) {
				c := target.(*getChatMessagesCursor)
				if c.LastMessageID != "msg_456" {
					t.Errorf("expected msg_456, got %s", c.LastMessageID)
				}
			},
		},
		{
			name:    "decode profile cursor",
			encoded: encodedProfileCursor,
			target:  &searchProfilesCursor{},
			wantErr: false,
			checkResult: func(t *testing.T, target any) {
				c := target.(*searchProfilesCursor)
				if c.LastUserID != "user_789" {
					t.Errorf("expected user_789, got %s", c.LastUserID)
				}
			},
		},
		{
			name:    "invalid base64",
			encoded: "not-valid-base64!!!",
			target:  &getChatsCursor{},
			wantErr: true,
		},
		{
			name:    "invalid json",
			encoded: "bm90LWpzb24=", // "not-json" in base64
			target:  &getChatsCursor{},
			wantErr: true,
		},
		{
			name:    "wrong cursor type",
			encoded: encodedChatCursor,
			target:  &getChatMessagesCursor{},
			wantErr: false,
		},
		{
			name:    "empty string",
			encoded: "",
			target:  &getChatsCursor{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decodeCursor(tt.encoded, tt.target)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tt.checkResult != nil {
					tt.checkResult(t, tt.target)
				}
			}
		})
	}
}

func TestCursoredResponse(t *testing.T) {
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	resp := cursoredResponse[Item]{
		Items: []Item{
			{ID: "1", Name: "Item 1"},
			{ID: "2", Name: "Item 2"},
		},
		Cursor: "abc123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded cursoredResponse[Item]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(decoded.Items))
	}

	if decoded.Cursor != "abc123" {
		t.Errorf("expected cursor abc123, got %s", decoded.Cursor)
	}
}

func TestExtractPlatformAndOSFromUserAgent(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		wantPlat string
		wantOS   string
	}{
		{
			name:     "empty user agent",
			ua:       "",
			wantPlat: "Unknown",
			wantOS:   "Unknown",
		},
		{
			name:     "chrome on windows 10",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Windows 10",
		},
		{
			name:     "firefox on windows 10",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			wantPlat: "Firefox",
			wantOS:   "Windows 10",
		},
		{
			name:     "safari on macos",
			ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			wantPlat: "Safari",
			wantOS:   "macOS",
		},
		{
			name:     "chrome on macos",
			ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "macOS",
		},
		{
			name:     "safari on ios",
			ua:       "Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
			wantPlat: "Safari",
			wantOS:   "macOS",
		},
		{
			name:     "chrome on android",
			ua:       "Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Linux",
		},
		{
			name:     "edge on windows",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
			wantPlat: "Chrome",
			wantOS:   "Windows 10",
		},
		{
			name:     "opera on windows",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
			wantPlat: "Chrome",
			wantOS:   "Windows 10",
		},
		{
			name:     "brave on windows",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Brave/1.65.84",
			wantPlat: "Chrome",
			wantOS:   "Windows 10",
		},
		{
			name:     "chrome on linux",
			ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Linux",
		},
		{
			name:     "ipad",
			ua:       "Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/604.1",
			wantPlat: "Safari",
			wantOS:   "macOS",
		},
		{
			name:     "windows 8.1",
			ua:       "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Windows 8.1",
		},
		{
			name:     "windows 8",
			ua:       "Mozilla/5.0 (Windows NT 6.2; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Windows 8",
		},
		{
			name:     "windows 7",
			ua:       "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantPlat: "Chrome",
			wantOS:   "Windows 7",
		},
		{
			name:     "unknown browser",
			ua:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			wantPlat: "Unknown",
			wantOS:   "Windows 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plat, os := extractPlatformAndOSFromUserAgent(tt.ua)
			if plat != tt.wantPlat {
				t.Errorf("platform: expected %q, got %q", tt.wantPlat, plat)
			}
			if os != tt.wantOS {
				t.Errorf("os: expected %q, got %q", tt.wantOS, os)
			}
		})
	}
}
