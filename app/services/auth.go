package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/assaidy/blink/app/env"
	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils/email"
	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/oklog/ulid/v2"
)

type AuthService struct {
	db      *sql.DB
	queries *repo.Queries
	mailer  email.Mailer
}

func NewAuthService(db *sql.DB, mailer email.Mailer) *AuthService {
	return &AuthService{
		db:      db,
		queries: repo.New(db),
		mailer:  mailer,
	}
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func validateRegisterParams(name, username, email, bio string) error {
	type Params struct {
		Name     string
		Username string
		Email    string
		Bio      string
	}
	params := Params{Name: name, Username: username, Email: email, Bio: bio}

	return validation.ValidateStruct(&params,
		validation.Field(&params.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&params.Username, validation.Required, validation.Length(2, 50),
			validation.Match(usernameRegex).Error("only letters, numbers, and _ are allowed")),
		// max len 255 because is.Email doesn't check the length
		validation.Field(&params.Email, validation.Required, is.Email, validation.Length(0, 255)),
		validation.Field(&params.Bio, validation.Length(0, 255)),
	)
}

func (me *AuthService) Register(name, username, email, bio string) error {
	name = strings.TrimSpace(name)
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	bio = strings.TrimSpace(bio)

	if err := validateRegisterParams(name, username, email, bio); err != nil {
		return fmt.Errorf("%w: %w", ErrValidation, err)
	}

	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	qtx := me.queries.WithTx(tx)

	if ok, err := qtx.CheckUsername(ctx, username); err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	} else if ok {
		return ErrUsernameTaken
	}

	if ok, err := qtx.CheckEmail(ctx, email); err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	} else if ok {
		return ErrEmailTaken
	}

	if err := qtx.InsertUser(ctx, repo.InsertUserParams{
		ID:       ulid.Make().String(),
		Name:     name,
		Username: username,
		Email:    email,
		Bio:      bio,
	}); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func validateSendOtpParams(channel, identifier, purpose string) error {
	type Params struct {
		Channel   string
		Identifer string
		Purpose   string
	}
	params := Params{Channel: channel, Identifer: identifier, Purpose: purpose}

	return validation.ValidateStruct(&params,
		validation.Field(&params.Channel, validation.Required, validation.In("email")),
		validation.Field(&params.Identifer, validation.Required, validation.By(func(value any) error {
			switch channel {
			case "email":
				return validation.Validate(value, is.Email)
			}
			return nil
		})),
		validation.Field(&params.Purpose, validation.Required, validation.In("login")),
	)
}

// SendOtp sends an OTP and returns its ID
func (me *AuthService) SendOtp(channel, identifier, purpose string) (string, error) {
	channel = strings.ToLower(strings.TrimSpace(channel))
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	purpose = strings.ToLower(strings.TrimSpace(purpose))

	if err := validateSendOtpParams(channel, identifier, purpose); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidation, err)
	}

	ctx := context.Background()

	var user repo.User
	switch channel {
	case "email":
		var err error
		user, err = me.queries.GetUserByEmail(ctx, identifier)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", ErrEmailNotFound
			}
			return "", fmt.Errorf("failed to get user by email: %w", err)
		}
	}

	otp, err := generateRandomOtp()
	if err != nil {
		return "", fmt.Errorf("failed to generate otp: %w", err)
	}
	otpHash := hashOtp(otp)

	otpID := ulid.Make().String()
	if err := me.queries.InsertOtp(ctx, repo.InsertOtpParams{
		ID:        otpID,
		UserID:    user.ID,
		OtpHash:   otpHash,
		Channel:   channel,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}); err != nil {
		return "", fmt.Errorf("faield to insert otp: %w", err)
	}

	switch channel {
	case "email":
		if err := me.mailer.SendEmail(
			user.Email,
			"Blink OTP",
			fmt.Sprintf("Your one-time password is: %s", otp),
		); err != nil {
			return "", fmt.Errorf("failed to send email: %w", err)
		}
	}

	return otpID, nil
}

// TODO: Use an advanced library e.g. https://github.com/pquerna/otp
func generateRandomOtp() (string, error) {
	// Generate a number between 0 and 899,999
	n, err := rand.Int(rand.Reader, big.NewInt(900_000))
	if err != nil {
		return "", err
	}
	// Add 100,000 to ensure it is always 6 digits
	return fmt.Sprintf("%06d", n.Int64()+100_000), nil
}

func hashOtp(otp string) string {
	mac := hmac.New(sha256.New, []byte(env.Secret))
	mac.Write([]byte(otp))
	sum := mac.Sum(nil)
	return hex.EncodeToString([]byte(sum))
}

func verifyOtp(otp string, hash string) bool {
	actual := hashOtp(otp)
	return hmac.Equal([]byte(actual), []byte(hash))
}

type LoginSession struct {
	SessionToken string
	CsrfToken    string
	ExpiresAt    time.Time
}

func validateVerifyOtpParams(otpID, otp, platform, os string) error {
	type Params struct {
		OtpID    string
		Otp      string
		Platform string
		OS       string
	}
	params := Params{OtpID: otpID, Otp: otp, Platform: platform, OS: os}

	return validation.ValidateStruct(&params,
		validation.Field(&params.OtpID, validation.Required),
		validation.Field(&params.Otp, validation.Required),
		validation.Field(&params.Platform, validation.Required),
		validation.Field(&params.OS, validation.Required),
	)
}

// VerifyOtp verifies OTP and creates a session if OTP was requested for login, otherwise returns a nil session
func (me *AuthService) VerifyOtp(otpID, otp, platform, os string) (*LoginSession, error) {
	platform = strings.TrimSpace(platform)
	os = strings.TrimSpace(os)

	if err := validateVerifyOtpParams(otpID, otp, platform, os); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	ctx := context.Background()

	storedOtp, err := me.queries.GetOtpByID(ctx, otpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidOTP
		}
		return nil, fmt.Errorf("failed to get otp by id: %w", err)
	}

	if time.Since(storedOtp.ExpiresAt) >= 0 || !verifyOtp(otp, storedOtp.OtpHash) {
		return nil, ErrInvalidOTP
	}

	switch storedOtp.Purpose {
	case "login":
		sessionID := ulid.Make()
		sessionToken := fmt.Sprintf("%s_%s", sessionID, generateCryptoRandomHex(32))
		csrfToken := fmt.Sprintf("%s_%s", sessionID, generateCryptoRandomHex(32))
		// Set cookie max-age to 400 days: https://developer.chrome.com/blog/cookie-max-age-expires
		sessionExpiration := time.Now().Add(400 * 24 * time.Hour)

		if err := me.queries.InsertSession(ctx, repo.InsertSessionParams{
			ID:        sessionID.String(),
			Token:     sessionToken,
			CsrfToken: csrfToken,
			UserID:    storedOtp.UserID,
			Platform:  platform,
			Os:        os,
			ExpiresAt: sessionExpiration,
		}); err != nil {
			return nil, fmt.Errorf("failed to insert session: %w", err)
		}

		if err := me.queries.MarkEmailAsVerified(ctx, storedOtp.UserID); err != nil {
			return nil, fmt.Errorf("failed to mark email as verified: %w", err)
		}

		return &LoginSession{SessionToken: sessionToken, CsrfToken: csrfToken, ExpiresAt: sessionExpiration}, nil
	}

	return nil, nil
}

func generateCryptoRandomHex(nBytes uint) string {
	buf := make([]byte, nBytes)
	rand.Reader.Read(buf)
	return hex.EncodeToString(buf)
}

func (me *AuthService) ValidateSessionToken(sessionToken string) (sessionID, userID string, error error) {
	if sessionToken == "" {
		return "", "", ErrUnauthorized
	}

	ctx := context.Background()
	session, err := me.queries.GetSessionById(ctx, strings.Split(sessionToken, "_")[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrUnauthorized
		}
		return "", "", fmt.Errorf("failed to get session by id: %w", err)
	}

	if session.Token != sessionToken || !time.Now().Before(session.ExpiresAt) {
		return "", "", ErrUnauthorized
	}

	return session.ID, session.UserID, nil
}

func (me *AuthService) ValidateSessionAndCsrfTokens(sessionToken, csrfToken string) (sessionID, userID string, error error) {
	if sessionToken == "" || csrfToken == "" {
		return "", "", ErrUnauthorized
	}

	ctx := context.Background()
	session, err := me.queries.GetSessionById(ctx, strings.Split(sessionToken, "_")[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrUnauthorized
		}
		return "", "", fmt.Errorf("failed to get session by id: %w", err)
	}

	if session.Token != sessionToken || session.CsrfToken != csrfToken || !time.Now().Before(session.ExpiresAt) {
		return "", "", ErrUnauthorized
	}

	return session.ID, session.UserID, nil
}

func (me *AuthService) DeleteSession(userID, sessionID string) error {
	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	qtx := me.queries.WithTx(tx)
	if ok, err := qtx.CheckSessionForUser(ctx, repo.CheckSessionForUserParams{
		ID:     sessionID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("failed to check session for user: %w", err)
	} else if !ok {
		return ErrNotFound
	}

	if err := qtx.RemoveSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to remove session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

type ActiveSession = repo.GetActiveSessionsForUserRow

func (me *AuthService) GetActiveSessionsForUser(userID string) ([]ActiveSession, error) {
	ctx := context.Background()
	sessions, err := me.queries.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("faield to get user sessions: %w", err)
	}
	return sessions, nil
}
