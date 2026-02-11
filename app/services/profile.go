package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/utils/pubsub"
	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ProfileService struct {
	db      *sql.DB
	queries *repo.Queries
	pubsub  pubsub.Pubsub
}

func NewProfileService(db *sql.DB, pubsub pubsub.Pubsub) *ProfileService {
	return &ProfileService{
		db:      db,
		queries: repo.New(db),
		pubsub:  pubsub,
	}
}

type Profile = repo.User

func (me *ProfileService) SearchProfiles(query string, limit int, lastUserID string) ([]Profile, error) {
	ctx := context.Background()
	users, err := me.queries.SearchUsers(ctx, repo.SearchUsersParams{
		LastID: lastUserID,
		Query:  query,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	return users, nil
}

func (me *ProfileService) GetProfile(userID string) (Profile, error) {
	ctx := context.Background()
	user, err := me.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, utils.NewError(utils.NotFound, "user not found")
		}
		return Profile{}, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

type UpdateProfileParams struct {
	Name     string
	Username string
	Email    string
	Bio      string
}

func (me *UpdateProfileParams) cleanAndValidate() error {
	me.Name = strings.TrimSpace(me.Name)
	me.Username = strings.TrimSpace(me.Username)
	me.Email = strings.ToLower(strings.TrimSpace(me.Email))
	me.Bio = strings.TrimSpace(me.Bio)

	return validation.ValidateStruct(me,
		validation.Field(&me.Name, validation.Required, validation.Length(2, 50)),
		validation.Field(&me.Username, validation.Required, validation.Length(2, 50),
			validation.Match(usernameRegex).Error("only letters, numbers, and _ are allowed"),
		),
		// Max len 255 because is.Email doesn't check the length
		validation.Field(&me.Email, validation.Required, is.Email, validation.Length(0, 255)),
		validation.Field(&me.Bio, validation.Length(0, 255)),
	)
}

const ProfileWasUpdatedEvent = "ProfileWasUpdatedEvent"

type ProfileWasUpdatedEventPayload struct {
	UserID    string `json:"userID"`
	PartnerID string `json:"partnerID"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
}

func (me *ProfileService) UpdateProfile(userID string, params UpdateProfileParams) error {
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

	user, err := qtx.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NewError(utils.NotFound, "user not found")
		}
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if user.Username != params.Username {
		if ok, err := qtx.CheckUsername(ctx, params.Username); err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		} else if ok {
			return utils.NewError(utils.UsernameConflict, nil)
		}
	}

	if user.Email != params.Email {
		if ok, err := qtx.CheckEmail(ctx, params.Email); err != nil {
			return fmt.Errorf("failed to check email: %w", err)
		} else if ok {
			return utils.NewError(utils.EmailConflict, nil)
		}
	}

	if err := qtx.UpdateUser(ctx, repo.UpdateUserParams{
		ID:       userID,
		Name:     params.Name,
		Username: params.Username,
		Email:    params.Email,
		Bio:      params.Bio,
	}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	partnerIDs, err := me.queries.GetAllChatPartnerIDs(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get chat partner IDs: %w", err)
	}

	for _, id := range partnerIDs {
		if err := me.pubsub.Publish(ctx,
			ProfileWasUpdatedEvent,
			pubsub.JsonMessageGenerator,
			ProfileWasUpdatedEventPayload{
				UserID:    id,
				PartnerID: userID,
				Name:      params.Name,
				Email:     params.Email,
				Bio:       params.Bio,
			},
		); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", ProfileWasUpdatedEvent, err)
		}
	}

	return nil
}

const ProfileWasDeletedEvent = "ProfileWasDeletedEvent"

type ProfileWasDeletedEventPayload struct {
	UserID    string `json:"userID"`
	PartnerID string `json:"partnerID"`
}

func (me *ProfileService) DeleteProfile(userID string) error {
	ctx := context.Background()
	if ok, err := me.queries.CheckUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to check user id: %w", err)
	} else if !ok {
		return utils.NewError(utils.Unauthorized, nil)
	}

	if err := me.queries.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}

	partnerIDs, err := me.queries.GetAllChatPartnerIDs(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get chat partner IDs: %w", err)
	}

	for _, id := range partnerIDs {
		if err := me.pubsub.Publish(ctx,
			ProfileWasDeletedEvent,
			pubsub.JsonMessageGenerator,
			ProfileWasDeletedEventPayload{
				UserID:    id,
				PartnerID: userID,
			},
		); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", ProfileWasDeletedEvent, err)
		}
	}

	return nil
}
