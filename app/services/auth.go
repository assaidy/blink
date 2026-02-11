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
	"github.com/assaidy/blink/app/utils"
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

type RegisterParams struct {
	Name     string
	Username string
	Email    string
	Bio      string
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (me *RegisterParams) cleanAndValidate() error {
	me.Name = strings.TrimSpace(me.Name)
	me.Username = strings.TrimSpace(me.Username)
	me.Email = strings.ToLower(strings.TrimSpace(me.Email))
	me.Bio = strings.TrimSpace(me.Bio)

	return validation.ValidateStruct(me,
		validation.Field(&me.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&me.Username, validation.Required, validation.Length(2, 50), validation.Match(usernameRegex).Error("only letters, numbers, and _ are allowed")),
		validation.Field(&me.Email, validation.Required, is.Email, validation.Length(0, 255)), // max len 255 because is.Email doesn't check the length
		validation.Field(&me.Bio, validation.Length(0, 255)),
	)
}

func (me *AuthService) Register(params RegisterParams) error {
	if err := params.cleanAndValidate(); err != nil {
		return utils.NewError(utils.InvalidData, err)
	}

	ctx := context.Background()
	tx, err := me.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	qtx := me.queries.WithTx(tx)

	if ok, err := qtx.CheckUsername(ctx, params.Username); err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	} else if ok {
		return utils.NewError(utils.UsernameConflict, nil)
	}

	if ok, err := qtx.CheckEmail(ctx, params.Email); err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	} else if ok {
		return utils.NewError(utils.EmailConflict, nil)
	}

	if err := qtx.InsertUser(ctx, repo.InsertUserParams{
		ID:       ulid.Make().String(),
		Name:     params.Name,
		Username: params.Username,
		Email:    params.Email,
		Bio:      params.Bio,
	}); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

type SendOtpParams struct {
	Channel   string // the communication channel to send opt through
	Identifer string // email/sms/... according to channel
	Purpose   string // login, password_reset, email_verify
}

func (me *SendOtpParams) cleanAndValidate() error {
	me.Channel = strings.ToLower(strings.TrimSpace(me.Channel))
	me.Identifer = strings.ToLower(strings.TrimSpace(me.Identifer))
	me.Purpose = strings.ToLower(strings.TrimSpace(me.Purpose))

	if err := validation.ValidateStruct(me,
		validation.Field(&me.Channel, validation.Required, validation.In("email")),
		validation.Field(&me.Identifer, validation.Required, validation.By(func(value any) error {
			switch me.Channel {
			case "email":
				return validation.Validate(value, is.Email)
			}
			return nil
		})),
		validation.Field(&me.Purpose, validation.Required, validation.In("login")),
	); err != nil {
		return err
	}

	return nil
}

// Sends an OTP and returns its ID
func (me *AuthService) SendOtp(params SendOtpParams) (string, error) {
	if err := params.cleanAndValidate(); err != nil {
		return "", utils.NewError(utils.InvalidData, err)
	}

	ctx := context.Background()

	var user repo.User
	switch params.Channel {
	case "email":
		var err error
		user, err = me.queries.GetUserByEmail(ctx, params.Identifer)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", utils.NewError(utils.EmailNotFound, nil)
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
		Channel:   params.Channel,
		Purpose:   params.Purpose,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}); err != nil {
		return "", fmt.Errorf("faield to insert otp: %w", err)
	}

	switch params.Channel {
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

type VerifyOtpParams struct {
	OtpID    string
	Otp      string
	Platform string
	OS       string
}

func (me *VerifyOtpParams) validate() error {
	me.Platform = strings.TrimSpace(me.Platform)
	me.OS = strings.TrimSpace(me.OS)

	return validation.ValidateStruct(me,
		validation.Field(&me.OtpID, validation.Required),
		validation.Field(&me.Otp, validation.Required),
		validation.Field(&me.Platform, validation.Required),
		validation.Field(&me.OS, validation.Required),
	)
}

type LoginSession struct {
	SessionToken string
	CsrfToken    string
	ExpiresAt    time.Time
}

// Verifies OTP and creates a session if OTP was requested for login, otherwise returns a nil session
func (me *AuthService) VerifyOtp(params VerifyOtpParams) (*LoginSession, error) {
	if err := params.validate(); err != nil {
		return nil, utils.NewError(utils.InvalidData, err)
	}

	ctx := context.Background()

	otp, err := me.queries.GetOtpByID(ctx, params.OtpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NewError(utils.InvalidOtp, nil)
		}
		return nil, fmt.Errorf("failed to get otp by id: %w", err)
	}

	if time.Since(otp.ExpiresAt) >= 0 || !verifyOtp(params.Otp, otp.OtpHash) {
		return nil, utils.NewError(utils.InvalidOtp, nil)
	}

	switch otp.Purpose {
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
			UserID:    otp.UserID,
			Platform:  params.Platform,
			Os:        params.OS,
			ExpiresAt: sessionExpiration,
		}); err != nil {
			return nil, fmt.Errorf("failed to insert session: %w", err)
		}

		if err := me.queries.MarkEmailAsVerified(ctx, otp.UserID); err != nil {
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
		return "", "", utils.NewError(utils.Unauthorized, nil)
	}

	ctx := context.Background()
	session, err := me.queries.GetSessionById(ctx, strings.Split(sessionToken, "_")[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", utils.NewError(utils.Unauthorized, nil)
		}
		return "", "", fmt.Errorf("failed to get session by id: %w", err)
	}

	if session.Token != sessionToken || !time.Now().Before(session.ExpiresAt) {
		return "", "", utils.NewError(utils.Unauthorized, nil)
	}

	return session.ID, session.UserID, nil
}

func (me *AuthService) ValidateSessionAndCsrfTokens(sessionToken, csrfToken string) (sessionID, userID string, error error) {
	if sessionToken == "" || csrfToken == "" {
		return "", "", utils.NewError(utils.Unauthorized, nil)
	}

	ctx := context.Background()
	session, err := me.queries.GetSessionById(ctx, strings.Split(sessionToken, "_")[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", utils.NewError(utils.Unauthorized, nil)
		}
		return "", "", fmt.Errorf("failed to get session by id: %w", err)
	}

	if session.Token != sessionToken || session.CsrfToken != csrfToken || !time.Now().Before(session.ExpiresAt) {
		return "", "", utils.NewError(utils.Unauthorized, nil)
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
		return utils.NewError(utils.NotFound, "session not found")
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
