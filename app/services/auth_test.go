package services

import (
	"strings"
	"testing"
)

func TestValidateRegisterParams(t *testing.T) {
	tests := []struct {
		name      string
		nameVal   string
		username  string
		email     string
		bio       string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid params",
			nameVal:  "John Doe",
			username: "johndoe",
			email:    "john@example.com",
			bio:      "Hello world",
			wantErr:  false,
		},
		{
			name:     "valid params with no optional fields",
			nameVal:  "John",
			username: "john",
			email:    "john@example.com",
			bio:      "",
			wantErr:  false,
		},
		{
			name:      "name too short",
			nameVal:   "J",
			username:  "johndoe",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Name",
		},
		{
			name:      "name too long",
			nameVal:   "John Doe John Doe John Doe John Doe John Doe John Doe John",
			username:  "johndoe",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Name",
		},
		{
			name:      "name required",
			nameVal:   "",
			username:  "johndoe",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Name",
		},
		{
			name:      "username too short",
			nameVal:   "John Doe",
			username:  "j",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Username",
		},
		{
			name:     "username too long",
			nameVal:  "John Doe",
			username: "johndoe_johndoe_johndoe_johndoe_johndoe_johndoe",
			email:    "john@example.com",
			bio:      "",
			wantErr:  false,
		},
		{
			name:      "username required",
			nameVal:   "John Doe",
			username:  "",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Username",
		},
		{
			name:      "username with invalid characters",
			nameVal:   "John Doe",
			username:  "john-doe",
			email:     "john@example.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "letters, numbers",
		},
		{
			name:      "invalid email",
			nameVal:   "John Doe",
			username:  "johndoe",
			email:     "johnatexample.com",
			bio:       "",
			wantErr:   true,
			errSubstr: "Email",
		},
		{
			name:      "email required",
			nameVal:   "John Doe",
			username:  "johndoe",
			email:     "",
			bio:       "",
			wantErr:   true,
			errSubstr: "Email",
		},
		{
			name:      "bio too long",
			nameVal:   "John Doe",
			username:  "johndoe",
			email:     "john@example.com",
			bio:       string(make([]byte, 256)),
			wantErr:   true,
			errSubstr: "Bio",
		},
		{
			name:     "bio at max length",
			nameVal:  "John Doe",
			username: "johndoe",
			email:    "john@example.com",
			bio:      string(make([]byte, 255)),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterParams(tt.nameVal, tt.username, tt.email, tt.bio)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateSendOtpParams(t *testing.T) {
	tests := []struct {
		name       string
		channel    string
		identifier string
		purpose    string
		wantErr    bool
		errSubstr  string
	}{
		{
			name:       "valid email params",
			channel:    "email",
			identifier: "john@example.com",
			purpose:    "login",
			wantErr:    false,
		},
		{
			name:       "channel required",
			channel:    "",
			identifier: "john@example.com",
			purpose:    "login",
			wantErr:    true,
			errSubstr:  "Channel",
		},
		{
			name:       "invalid channel",
			channel:    "sms",
			identifier: "john@example.com",
			purpose:    "login",
			wantErr:    true,
			errSubstr:  "Channel",
		},
		{
			name:       "identifier required",
			channel:    "email",
			identifier: "",
			purpose:    "login",
			wantErr:    true,
			errSubstr:  "Identifier",
		},
		{
			name:       "invalid email format",
			channel:    "email",
			identifier: "johnatexample.com",
			purpose:    "login",
			wantErr:    true,
			errSubstr:  "Identifier",
		},
		{
			name:       "purpose required",
			channel:    "email",
			identifier: "john@example.com",
			purpose:    "",
			wantErr:    true,
			errSubstr:  "Purpose",
		},
		{
			name:       "invalid purpose",
			channel:    "email",
			identifier: "john@example.com",
			purpose:    "register",
			wantErr:    true,
			errSubstr:  "Purpose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSendOtpParams(tt.channel, tt.identifier, tt.purpose)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateVerifyOtpParams(t *testing.T) {
	tests := []struct {
		name      string
		otpID     string
		otp       string
		platform  string
		os        string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid params",
			otpID:    "valid_otp_id",
			otp:      "123456",
			platform: "Chrome",
			os:       "Windows 10",
			wantErr:  false,
		},
		{
			name:      "otpID required",
			otpID:     "",
			otp:       "123456",
			platform:  "Chrome",
			os:        "Windows 10",
			wantErr:   true,
			errSubstr: "OtpID",
		},
		{
			name:      "otp required",
			otpID:     "valid_otp_id",
			otp:       "",
			platform:  "Chrome",
			os:        "Windows 10",
			wantErr:   true,
			errSubstr: "Otp",
		},
		{
			name:      "platform required",
			otpID:     "valid_otp_id",
			otp:       "123456",
			platform:  "",
			os:        "Windows 10",
			wantErr:   true,
			errSubstr: "Platform",
		},
		{
			name:      "os required",
			otpID:     "valid_otp_id",
			otp:       "123456",
			platform:  "Chrome",
			os:        "",
			wantErr:   true,
			errSubstr: "OS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVerifyOtpParams(tt.otpID, tt.otp, tt.platform, tt.os)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGenerateRandomOtp(t *testing.T) {
	for i := 0; i < 1000; i++ {
		otp, err := generateRandomOtp()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(otp) != 6 {
			t.Errorf("expected 6 digits, got %d", len(otp))
		}
		if otp[0] == '0' {
			t.Errorf("expected first digit to not be 0, got %c", otp[0])
		}
	}
}

func TestHashOtp(t *testing.T) {
	secret := "test_secret_key"

	hash1 := hashOtp("123456", secret)
	hash2 := hashOtp("123456", secret)
	hash3 := hashOtp("654321", secret)

	if hash1 != hash2 {
		t.Errorf("same OTP should produce same hash, got %s and %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("different OTPs should produce different hashes")
	}

	if hash1 == "" {
		t.Error("hash should not be empty")
	}
}

func TestCompareOtpAndHash(t *testing.T) {
	secret := "test_secret_key"

	otp := "123456"
	hash := hashOtp(otp, secret)

	if !compareOtpAndHash(otp, hash, secret) {
		t.Error("correct OTP should match hash")
	}

	if compareOtpAndHash("000000", hash, secret) {
		t.Error("incorrect OTP should not match hash")
	}

	if compareOtpAndHash("12345", hash, secret) {
		t.Error("wrong length OTP should not match hash")
	}
}
