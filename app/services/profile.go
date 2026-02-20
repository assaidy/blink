package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/assaidy/blink/app/repo"
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
			return Profile{}, ErrNotFound
		}
		return Profile{}, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func validateProfileUpdateParams(name, username, email, bio string) error {
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
		// Max len 255 because is.Email doesn't check the length
		validation.Field(&params.Email, validation.Required, is.Email, validation.Length(0, 255)),
		validation.Field(&params.Bio, validation.Length(0, 255)),
	)
}

const ProfileWasUpdatedEvent = "ProfileWasUpdatedEvent"

type ProfileWasUpdatedEventPayload struct {
	UserID    string `json:"userID"`
	PartnerID string `json:"partnerID"` // Empty when notifying self
	Name      string `json:"name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
}

func (me *ProfileService) UpdateProfile(userID, name, username, email, bio string) error {
	name = strings.TrimSpace(name)
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	bio = strings.TrimSpace(bio)

	if err := validateProfileUpdateParams(name, username, email, bio); err != nil {
		return fmt.Errorf("%w: %w", ErrValidation, err)
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
			return ErrNotFound
		}
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if user.Username != username {
		if ok, err := qtx.CheckUsername(ctx, username); err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		} else if ok {
			return ErrUsernameTaken
		}
	}

	if user.Email != email {
		if ok, err := qtx.CheckEmail(ctx, email); err != nil {
			return fmt.Errorf("failed to check email: %w", err)
		} else if ok {
			return ErrEmailTaken
		}
	}

	if err := qtx.UpdateUser(ctx, repo.UpdateUserParams{
		ID:       userID,
		Name:     name,
		Username: username,
		Email:    email,
		Bio:      bio,
	}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if err := me.pubsub.Publish(ctx,
		ProfileWasUpdatedEvent,
		pubsub.JsonMessageGenerator,
		ProfileWasUpdatedEventPayload{
			UserID:   userID,
			Name:     name,
			Username: username,
			Email:    email,
			Bio:      bio,
		},
	); err != nil {
		return fmt.Errorf("failed to publish event %s: %w", ProfileWasUpdatedEvent, err)
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
				Name:      name,
				Username:  username,
				Email:     email,
				Bio:       bio,
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
		return ErrUnauthorized
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
