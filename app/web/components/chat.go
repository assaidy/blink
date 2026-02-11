package components

import (
	"fmt"
	"time"

	"github.com/assaidy/gg"
)

type ProfileBlockParams struct {
	ID       string
	Name     string
	Username string
}

func profileBlock(profile ProfileBlockParams) gg.Element {
	return gg.Div(gg.KV{"class": "flex items-center gap-3 p-3 hover:bg-bg-tertiary/50 hover:rounded-lg cursor-pointer transition-colors"},
		gg.Div(gg.KV{"class": "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
			getInitials(profile.Name),
		),
		gg.Div(gg.KV{"class": "flex-1 min-w-0"},
			gg.P(gg.KV{"class": "text-fg-primary font-medium truncate"}, profile.Name),
			gg.P(gg.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+profile.Username),
		),
	)
}

type SearchResultParams struct {
	Query          string
	HasMore        bool
	ProfileResults []ProfileBlockParams
}

func SearchResult(params SearchResultParams) gg.Node {
	if params.Query == "" {
		return gg.P(gg.KV{"class": "text-fg-secondary text-center py-3"}, "Search for users by name or username")
	}

	if len(params.ProfileResults) == 0 {
		return gg.Empty()
	}

	lastID := params.ProfileResults[len(params.ProfileResults)-1].ID

	return gg.Div(gg.KV{"class": "space-y-1"},
		gg.MapSlice(params.ProfileResults, func(profile ProfileBlockParams) gg.Node {
			return gg.IfElse(profile.ID == lastID,
				gg.Div(gg.KV{
					"hx-get":     "/search/users?query=" + params.Query + "&cursor=" + lastID,
					"hx-trigger": "intersect once",
					"hx-swap":    "afterend",
				},
					profileBlock(profile),
				),
				profileBlock(profile),
			)
		}),
	)
}

type PartnersListParams struct {
	Partners []ProfileBlockParams
	// This is the cursor. Empty when no more partners
	LastMessageWithLastPartnerID string
}

func PartnersList(params PartnersListParams) gg.Node {
	if len(params.Partners) == 0 {
		return gg.Empty()
	}

	lastID := params.Partners[len(params.Partners)-1].ID

	return gg.Div(gg.KV{"class": "space-y-1"},
		gg.MapSlice(params.Partners, func(partner ProfileBlockParams) gg.Node {
			attrs := gg.IfElse(partner.ID == lastID && params.LastMessageWithLastPartnerID != "", gg.KV{
				"hx-get":       "/partners?cursor=" + params.LastMessageWithLastPartnerID,
				"hx-trigger":   "intersect once",
				"hx-swap":      "afterend",
				"hx-indicator": "#partners-indicator",
				// Disable the sidebar indicator for requests in blocks
				"hx-disinherit": "hx-indicator",
			}, nil)

			return gg.Div(attrs,
				gg.Div(gg.KV{
					"hx-get":     "/chat/" + partner.ID,
					"hx-trigger": "click",
					"hx-target":  "#chat-container",
					"hx-swap":    "innerHTML",
					// Cancel the request if clicking the active partner
					"hx-on::config-request": `
						const oldActive = document.getElementById('active-chat-partner');
						if (oldActive) {
							if (oldActive == this) {
								event.preventDefault();
								return;
							}
							oldActive.removeAttribute('id');
							oldActive.classList.remove('bg-bg-tertiary', 'rounded-lg');
							oldActive.classList.add('hover:bg-bg-tertiary/50', 'hover:rounded-lg')
						}
						console.log('setting active', this)
						this.setAttribute('id', 'active-chat-partner');
						this.classList.remove('hover:bg-bg-tertiary/50', 'hover:rounded-lg');
						this.classList.add('bg-bg-tertiary', 'rounded-lg')
					`,
				},
					profileBlock(partner),
				),
			)
		}),
	)
}

type ChatContainerParams struct {
	Partner ProfileBlockParams
}

func ChatContainer(params ChatContainerParams) gg.Node {
	return gg.Div(gg.KV{"class": "flex-1 flex flex-col bg-bg-primary h-full"},
		gg.Div(gg.KV{"class": "px-4 py-3 bg-bg-secondary border-b border-bg-tertiary flex items-center gap-3"},
			gg.Div(gg.KV{"class": "w-10 h-10 rounded-full bg-blue flex items-center justify-center text-bg-primary font-bold"},
				getInitials(params.Partner.Name),
			),
			gg.Div(gg.KV{"class": "flex-1 min-w-0"},
				gg.P(gg.KV{"class": "text-fg-primary font-medium truncate"}, params.Partner.Name),
				gg.P(gg.KV{"class": "text-fg-secondary text-sm truncate"}, "@"+params.Partner.Username),
			),
		),
		gg.Div(gg.KV{
			"id":    "messages-container",
			"class": "flex-1 overflow-y-auto px-4 py-4 flex flex-col-reverse gap-3",
		},
			gg.Div(gg.KV{
				"hx-get":              fmt.Sprintf("/chat/%s/messages", params.Partner.ID),
				"hx-trigger":          "load",
				"hx-swap":             "afterend",
				"hx-indicator":        "#messages-indicator",
				"hx-on::after-settle": "this.remove()",
			}),
			gg.Div(gg.KV{"class": "flex justify-center"},
				spinner("messages-indicator"),
			),
		),
	)
}

type ChatMessagesListParams struct {
	PartnerID string
	Messages  []ChatMessageParams
	HasMore   bool
}

func ChatMessagesList(params ChatMessagesListParams) gg.Node {
	if len(params.Messages) == 0 {
		return gg.Empty()
	}

	cursorMessageID := params.Messages[len(params.Messages)-1].ID

	return gg.MapSlice(params.Messages, func(msg ChatMessageParams) gg.Node {
		return gg.Div(
			gg.IfElse(params.HasMore && msg.ID == cursorMessageID,
				gg.KV{
					"hx-get":       fmt.Sprintf("/chat/%s/messages?cursor=%s", params.PartnerID, cursorMessageID),
					"hx-trigger":   "intersect once",
					"hx-swap":      "afterend",
					"hx-indicator": "#messages-indicator",
					// For internal requests not to use this indicator
					"hx-disinherit": "hx-indicator",
				},
				nil,
			),
			chatMessage(msg),
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

func chatMessage(msg ChatMessageParams) gg.Node {
	return gg.Div(gg.KV{"class": "flex w-full " + gg.IfElse(msg.FromMe, "justify-end", "justify-start")},
		gg.Div(gg.KV{"class": "flex flex-col " + gg.IfElse(msg.FromMe, "items-end", "items-start") + " max-w-[70%]"},
			gg.Div(gg.KV{"class": "px-4 py-2 " + gg.IfElse(msg.FromMe, "bg-blue text-bg-primary rounded-l-2xl rounded-tr-2xl", "bg-bg-tertiary text-fg-primary rounded-r-2xl rounded-tl-2xl")},
				gg.P(gg.KV{"class": "break-words"}, msg.Content),
			),
			gg.Div(gg.KV{"class": "flex items-center gap-1 mt-1 px-1"},
				gg.P(gg.KV{"class": "text-xs text-fg-secondary"}, msg.SentAt.Format("Jan 2, 3:04 PM")),
				gg.If(msg.FromMe && msg.IsRead,
					gg.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue"><polyline points="20 6 9 17 4 12"></polyline></svg>`),
				),
			),
		),
	)
}
