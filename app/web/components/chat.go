package components

import (
	"github.com/assaidy/gg"
)

type ChatPageParams struct {
	UserID    string
	SessionID string
}

func ChatPage(params ChatPageParams) gg.Node {
	return rootLayout(
		gg.Div(gg.KV{"class": "h-screen flex flex-col justify-center items-center bg-bg-primary"},
			gg.H1(gg.KV{"class": "text-3xl font-bold text-fg-primary mb-4"}, "User: ", params.UserID),
			gg.H1(gg.KV{"class": "text-3xl font-bold text-fg-primary"}, "Session ID: ", params.SessionID),
		),
	)
}
