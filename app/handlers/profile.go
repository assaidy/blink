package handlers

import (
	"fmt"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/gofiber/fiber/v2"
)

type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

type SearchProfilesCursor struct {
	LastUserID string `json:"lastUserID"`
}

type SearchProfileResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type SearchProfileResponse CursoredResponse[SearchProfileResponseItem]

func (me *ProfileHandler) HandleApiSearchProfiles(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
			Items: []SearchProfileResponseItem{},
		})
	}

	var requestCursor SearchProfilesCursor
	if cq := c.Query("cursor"); cq != "" {
		if err := decodeCursor(cq, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, err)
		}
	}

	limit := 15
	profiles, err := me.profileService.SearchProfiles(query, limit, requestCursor.LastUserID)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(profiles) {
		if encodedResponseCursor, err = encodeCursor(SearchProfilesCursor{
			LastUserID: profiles[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]SearchProfileResponseItem, 0, len(profiles))
	for _, p := range profiles {
		responseItems = append(responseItems, SearchProfileResponseItem{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SearchProfileResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

type GetProfileResponse struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Bio      string    `json:"bio"`
	JoinedAt time.Time `json:"joinedAt"`
}

func (me *ProfileHandler) HandleApiGetProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(c.Params("user_id"))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(GetProfileResponse{
		ID:       profile.ID,
		Name:     profile.Name,
		Username: profile.Username,
		Bio:      profile.Bio,
		JoinedAt: profile.JoinedAt,
	})
}

type GetMyProfileResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailIsVerified bool      `json:"emailIsVerified"`
	Bio             string    `json:"bio"`
	JoinedAt        time.Time `json:"joinedAt"`
}

func (me *ProfileHandler) HandleApiGetMyProfile(c *fiber.Ctx) error {
	profile, err := me.profileService.GetProfile(getCurrentUserID(c))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(GetMyProfileResponse{
		ID:              profile.ID,
		Name:            profile.Name,
		Username:        profile.Username,
		Email:           profile.Email,
		EmailIsVerified: profile.EmailIsVerified,
		Bio:             profile.Bio,
		JoinedAt:        profile.JoinedAt,
	})
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
}

func (me *ProfileHandler) HandleApiUpdateProfile(c *fiber.Ctx) error {
	var request UpdateProfileRequest
	if err := c.BodyParser(&request); err != nil {
		return utils.NewError(utils.InvalidJson, err)
	}

	if err := me.profileService.UpdateProfile(getCurrentUserID(c), services.UpdateProfileParams{
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
		Bio:      request.Bio,
	}); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (me *ProfileHandler) HandleApiDeleteProfile(c *fiber.Ctx) error {
	if err := me.profileService.DeleteProfile(getCurrentUserID(c)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}
