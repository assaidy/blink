package components

import (
	"fmt"
	"time"

	"github.com/assaidy/hyper"
)

type ProfileBlockParams struct {
	ID       string
	Name     string
	Username string
	IsOnline bool
}

func ProfileBlockPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		"id": "profile-block-presence-indicator-" + partnerID,
		"class": "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			h.IfElse(!isOnline, "hidden", ""),
	})
}

func ChatContainerPresenceIndicator(partnerID string, isOnline bool) h.Node {
	return h.Div(h.KV{
		"id": "chat-container-presence-indicator-" + partnerID,
		"class": "absolute bottom-0 right-0 w-3.5 h-3.5 bg-green rounded-full border-2 border-bg-primary " +
			h.IfElse(!isOnline, "hidden", ""),
	})
}

func profileBlock(profile ProfileBlockParams) h.Node {
	return h.Div(h.KV{"class": "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		h.Div(h.KV{"class": "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(profile.Name),
			ProfileBlockPresenceIndicator(profile.ID, profile.IsOnline),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, profile.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+profile.Username),
		),
	)
}

type SearchResultParams struct {
	Query          string
	HasMore        bool
	ProfileResults []ProfileBlockParams
}

func SearchResult(params SearchResultParams) h.Node {
	if params.Query == "" {
		return h.P(h.KV{"class": "text-fg-secondary text-center py-3"}, "Search for users by name or username")
	}

	if len(params.ProfileResults) == 0 {
		return h.Empty()
	}

	lastID := params.ProfileResults[len(params.ProfileResults)-1].ID

	return h.Div(h.KV{"class": "space-y-1"},
		h.MapSlice(params.ProfileResults, func(profile ProfileBlockParams) h.Node {
			return h.Div(h.KV{
				"hx-get":     "/search/users/select/" + profile.ID,
				"hx-trigger": "click",
				"hx-target":  "#chat-container",
				"hx-swap":    "innerHTML",
			},
				h.IfElse(profile.ID == lastID,
					h.Div(h.KV{
						"hx-get":     "/search/users?query=" + params.Query + "&cursor=" + lastID,
						"hx-trigger": "intersect once",
						"hx-swap":    "afterend",
					},
						profileBlock(profile),
					),
					profileBlock(profile),
				),
			)
		}),
	)
}

type PartnersListParams struct {
	Partners []ProfileBlockParams
	// This is the cursor. Empty when no more partners
	LastMessageWithLastPartnerID string
}

func PartnersList(params PartnersListParams) h.Node {
	if len(params.Partners) == 0 {
		return h.Empty()
	}

	lastID := params.Partners[len(params.Partners)-1].ID

	return h.Div(h.KV{
		"id":    "partners-list",
		"class": "space-y-1",
	},
		h.MapSlice(params.Partners, func(partner ProfileBlockParams) h.Node {
			attrs := h.IfElse(partner.ID == lastID && params.LastMessageWithLastPartnerID != "", h.KV{
				"hx-get":       "/partners?cursor=" + params.LastMessageWithLastPartnerID,
				"hx-trigger":   "intersect once",
				"hx-swap":      "afterend",
				"hx-indicator": "#partners-indicator",
				// Disable the sidebar indicator for requests in blocks
				"hx-disinherit": "hx-indicator",
			}, nil)

			return h.Div(attrs, PartnersListItem(partner))
		}),
	)
}

func PartnersListItem(partner ProfileBlockParams) h.Node {
	return h.Div(h.KV{
		"id":         "partner-" + partner.ID,
		"class":      "cursor-pointer transition-colors",
		"hx-get":     "/chat/" + partner.ID,
		"hx-trigger": "click",
		"hx-target":  "#chat-container",
		"hx-swap":    "innerHTML",
		// Cancel the request if clicking the active partner
		"hx-on::config-request": `
			if (window.currentActivePartnerId === "` + partner.ID + `") {
				event.preventDefault();
				return;
			}
		`,
		// Re-applies data-active attribute.
		// This prevents losing the active styles for newly instered element if it was the active
		// as the old one which has the attribute is deleted by the oob-swap response.
		"hx-on::load": `
			document.getElementById("partner-"+window.currentActivePartnerId)?.setAttribute("data-active", "");
		`,
	},
		profileBlock(partner),
	)
}

type ChatContainerParams struct {
	Partner ProfileBlockParams
}

func ChatContainer(params ChatContainerParams) h.Node {
	return h.Div(h.KV{
		"class": "flex-1 flex flex-col bg-bg-primary h-full",
		"hx-on::load": `
			window.currentActivePartnerId = "` + params.Partner.ID + `";
		  document.querySelector("#partners-list [data-active]")?.removeAttribute("data-active");
		  document.getElementById("partner-" + "` + params.Partner.ID + `").setAttribute("data-active", "");
		`,
	},
		ChatContainerHeader(params.Partner),
		h.Div(h.KV{
			"id":    "messages-container",
			"class": "flex-1 overflow-y-auto px-6 sm:px-10 py-4 flex flex-col-reverse gap-3",
		},
			h.Div(h.KV{
				"id": "new-message-inserter",
				"hx-on::oob-before-swap": `
					// don't insert the new message if it doesn't come from the active partner
					if (window.currentActivePartnerId !== "` + params.Partner.ID + `") {
						event.preventDefault();
						return;
					}
				`,
			}),
			h.Div(h.KV{
				"hx-get":              fmt.Sprintf("/chat/%s/messages", params.Partner.ID),
				"hx-trigger":          "load",
				"hx-swap":             "afterend",
				"hx-indicator":        "#messages-indicator",
				"hx-on::after-settle": "this.remove()",
			}),
			h.Div(h.KV{"class": "flex justify-center"},
				spinner("messages-indicator"),
			),
		),
		ChatInputForm(ChatInputFormParams{PartnerID: params.Partner.ID}),
	)
}

func ChatContainerHeader(partner ProfileBlockParams) h.Node {
	return h.Div(h.KV{
		"id":    "chat-container-header-" + partner.ID,
		"class": "h-16 px-4 bg-bg-secondary border-b border-bg-tertiary flex items-center gap-3",
	},
		h.Div(h.KV{"class": "relative w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(partner.Name),
			ChatContainerPresenceIndicator(partner.ID, partner.IsOnline),
		),
		h.Div(h.KV{"class": "flex-1 min-w-0"},
			h.P(h.KV{"class": "text-fg-primary font-medium truncate"}, partner.Name),
			h.P(h.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+partner.Username),
		),
	)
}

type ChatMessagesListParams struct {
	PartnerID string
	Messages  []ChatMessageParams
	HasMore   bool
}

func ChatMessagesList(params ChatMessagesListParams) h.Node {
	if len(params.Messages) == 0 {
		return h.Empty()
	}

	cursorMessageID := params.Messages[len(params.Messages)-1].ID

	return h.MapSlice(params.Messages, func(msg ChatMessageParams) h.Node {
		return h.Div(
			h.IfElse(params.HasMore && msg.ID == cursorMessageID,
				h.KV{
					"hx-get":       fmt.Sprintf("/chat/%s/messages?cursor=%s", params.PartnerID, cursorMessageID),
					"hx-trigger":   "intersect once",
					"hx-swap":      "afterend",
					"hx-indicator": "#messages-indicator",
					// For internal requests not to use this indicator
					"hx-disinherit": "hx-indicator",
				},
				nil,
			),
			ChatMessage(msg),
		)
	})
}

type ChatMessageParams struct {
	ID      string
	Content string
	SentAt  time.Time
	IsRead  bool
	FromMe  bool
}

func ChatMessage(msg ChatMessageParams) h.Node {
	return h.Div(h.KV{"class": "flex w-full " + h.IfElse(msg.FromMe, "justify-end", "justify-start")},
		h.Div(h.KV{"class": "flex flex-col " + h.IfElse(msg.FromMe, "items-end", "items-start") + " max-w-[70%]"},
			h.Div(h.KV{"class": "px-4 py-2 " + h.IfElse(msg.FromMe, "bg-blue text-bg-primary rounded-l-2xl rounded-tr-2xl", "bg-bg-tertiary text-fg-primary rounded-r-2xl rounded-tl-2xl")},
				h.P(h.KV{"class": "whitespace-pre-wrap"}, msg.Content),
			),
			h.Div(h.KV{"class": "flex items-center gap-1 mt-1 px-1"},
				// TODO: display the year in a sticky widget like telegram/whatsapp,
				// and move this paragraph inside the message box
				h.P(h.KV{"class": "text-xs text-fg-secondary"}, msg.SentAt.Format("Jan 2, 3:04 PM")),
				h.If(msg.FromMe && msg.IsRead,
					h.RawText(`<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue"><polyline points="20 6 9 17 4 12"></polyline></svg>`),
				),
			),
		),
	)
}
