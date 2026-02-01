package handlers

import (
	"fmt"
	"time"

	"github.com/assaidy/blink/app/services"
	"github.com/assaidy/blink/app/utils"
	"github.com/assaidy/blink/app/web/components"
	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

type GetChatsCursor struct {
	LastMessageIDWithLastPartner string `json:"lastMessageIDWithLastPartner"`
}

type GetChatPartnersResponseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	IsOnline bool   `json:"isOnline"`
}

type GetChatPartnersResponse CursoredResponse[GetChatPartnersResponseItem]

// TODO: add filter/search option
func (me *ChatHandler) HandleApiGetChatPartners(c *fiber.Ctx) error {
	var requestCursor GetChatsCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, nil)
		}
	}

	limit := 15
	partners, err := me.chatService.GetChatPartners(getCurrentUserID(c), requestCursor.LastMessageIDWithLastPartner, limit)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(partners) {
		if encodedResponseCursor, err = encodeCursor(GetChatsCursor{
			LastMessageIDWithLastPartner: partners[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]GetChatPartnersResponseItem, 0, len(partners))
	for _, p := range partners {
		responseItems = append(responseItems, GetChatPartnersResponseItem{
			ID:       p.ID,
			Name:     p.Name,
			Username: p.Username,
			IsOnline: p.IsOnline,
		})
	}

	return c.Status(fiber.StatusOK).JSON(GetChatPartnersResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

func (me *ChatHandler) HandleApiDeleteChat(c *fiber.Ctx) error {
	if err := me.chatService.DeleteChat(getCurrentUserID(c), c.Params("partner_id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

type GetChatMessagesCursor struct {
	LastMessageID string `json:"lastMessageID"`
}

type GetChatMessagesResponseItem struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	SentAt  time.Time `json:"sentAt"`
	IsRead  bool      `json:"isRead"`
	FromMe  bool      `json:"fromMe"`
}

type GetChatMessagesResponse CursoredResponse[GetChatMessagesResponseItem]

func (me *ChatHandler) HandleApiGetChatMessages(c *fiber.Ctx) error {
	var requestCursor GetChatMessagesCursor
	if qc := c.Query("cursor"); qc != "" {
		if err := decodeCursor(qc, &requestCursor); err != nil {
			return utils.NewError(utils.InvalidCursor, nil)
		}
	}

	limit := 15
	messages, err := me.chatService.GetChatMessages(getCurrentUserID(c), c.Params("partner_id"), requestCursor.LastMessageID, limit)
	if err != nil {
		return err
	}

	var encodedResponseCursor string
	if limit == len(messages) {
		if encodedResponseCursor, err = encodeCursor(GetChatMessagesCursor{
			LastMessageID: messages[limit-1].ID,
		}); err != nil {
			return fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	responseItems := make([]GetChatMessagesResponseItem, 0, len(messages))
	for _, m := range messages {
		responseItems = append(responseItems, GetChatMessagesResponseItem{
			ID:      m.ID,
			Content: m.Content,
			SentAt:  m.SentAt,
			IsRead:  m.IsRead,
			FromMe:  m.FromMe,
		})
	}

	return c.Status(fiber.StatusOK).JSON(GetChatMessagesResponse{
		Items:  responseItems,
		Cursor: encodedResponseCursor,
	})
}

func (me *ChatHandler) HandleApiMarkMessagesAsRead(c *fiber.Ctx) error {
	if err := me.chatService.MarkMessagesAsRead(getCurrentUserID(c), c.Params("partner_id"), c.Query("upto_message_id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func (me *ChatHandler) HandleChatPage(c *fiber.Ctx) error {
	return render(c, components.ChatPage(components.ChatPageParams{
		UserID:    getCurrentUserID(c),
		SessionID: getCurrentSessionID(c),
	}))
}
