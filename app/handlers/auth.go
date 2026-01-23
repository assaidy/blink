package handlers

import (
	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/web/templates"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type CreateClientRequest struct {
	Platform string `json:"platform"`
	Os       string `json:"os"`
	App      string `json:"app"`
}

type CreateClientResponse struct {
	ClientID string `json:"clientID"`
}

func (me *AuthHandler) HandleCreateClient(c *fiber.Ctx) error {
	var request CreateClientRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	clientID, err := me.authService.CreateClient(services.CreateClientParams{
		Platform: request.Platform,
		Os:       request.Os,
		App:      request.App,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(CreateClientResponse{
		ClientID: clientID,
	})
}

type UpdateClientRequest struct {
	App string `json:"app"`
}

func (me *AuthHandler) HandleUpdateClient(c *fiber.Ctx) error {
	var request UpdateClientRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	if err := me.authService.UpdateClient(services.UpdateClientParams{
		ClientID: c.Params("client_id"),
		App:      request.App,
	}); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *AuthHandler) HandleRegister(c *fiber.Ctx) error {
	var request RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	if err := me.authService.Register(services.RegisterParams{
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
		Bio:      request.Bio,
	}); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

type RequestOtpRequest struct {
	Channel   string `json:"channel"`
	Identifer string `json:"identifier"`
	Purpose   string `json:"purpose"`
	ClientID  string `json:"clientID"` // the client app id
}

type RequestOtpResponse struct {
	OtpID string `json:"otpID"`
}

func (me *AuthHandler) HandleRequestOtp(c *fiber.Ctx) error {
	var request RequestOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	otpID, err := me.authService.SendOtp(services.SendOtpParams{
		Channel:   request.Channel,
		Identifer: request.Identifer,
		Purpose:   request.Purpose,
		ClientID:  request.ClientID,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(RequestOtpResponse{
		OtpID: otpID,
	})
}

type VerifyOtpRequest struct {
	OtpID string `json:"otpID"`
	Otp   string `json:"otp"`
}

func (me *AuthHandler) HandleVerifyOtp(c *fiber.Ctx) error {
	var request VerifyOtpRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	session, err := me.authService.VerifyOtp(services.VerifyOtpParams{
		OtpID: request.OtpID,
		Otp:   request.Otp,
	})
	if err != nil {
		return err
	}

	if session != nil {
		c.Cookie(&fiber.Cookie{
			Name:     "session_token",
			Value:    session.SessionToken,
			Expires:  session.ExpiresAt,
			HTTPOnly: true,
		})
		c.Cookie(&fiber.Cookie{
			Name:    "csrf_token",
			Value:   session.CsrfToken,
			Expires: session.ExpiresAt,
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

const (
	currentSessionID = "Auth.SessionID"
	currentUserID    = "Auth.UserID"
)

func (me *AuthHandler) WithSessionAndCSRFTokens(c *fiber.Ctx) error {
	sessionID, userID, err := me.authService.ValidateSession(c.Cookies("session_token"), c.Get("X-CSRF-Token"))
	if err != nil {
		return err
	}

	c.Locals(currentSessionID, sessionID)
	c.Locals(currentUserID, userID)
	return c.Next()
}

func getCurrentSessionID(c *fiber.Ctx) string {
	return c.Locals(currentSessionID).(string)
}

func getCurrentUserID(c *fiber.Ctx) string {
	return c.Locals(currentUserID).(string)
}

func (me *AuthHandler) HandleLogout(c *fiber.Ctx) error {
	if err := me.authService.Logout(getCurrentSessionID(c)); err != nil {
		return err
	}
	c.ClearCookie("session_token", "csrf_token")
	return c.SendStatus(fiber.StatusOK)
}

type GetActiveSessionsResponseItem struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Os       string `json:"os"`
	App      string `json:"app"`
}

func (me *AuthHandler) HandleGetActiveSessions(c *fiber.Ctx) error {
	sessions, err := me.authService.GetActiveSessionsForUser(getCurrentUserID(c))
	if err != nil {
		return err
	}

	response := make([]GetActiveSessionsResponseItem, 0, len(sessions))
	for _, s := range sessions {
		response = append(response, GetActiveSessionsResponseItem{
			ID:       s.ID,
			Platform: s.Platform,
			Os:       s.Os,
			App:      s.App,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (me *AuthHandler) HandleGetLoginPage(c *fiber.Ctx) error {
	return renderTempl(c, templates.LoginPage())
}

func (me *AuthHandler) HandleGetRegisterPage(c *fiber.Ctx) error {
	return renderTempl(c, templates.RegisterPage())
}

func (me *AuthHandler) HandleGetVerifyOtpPage(c *fiber.Ctx) error {
	return renderTempl(c, templates.VerifyOtpPage())
}
