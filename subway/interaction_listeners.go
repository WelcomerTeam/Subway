package internal

import (
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

// Creates a permanent component listener with no expiration.
// This is useful for components that are expected to be interacted with indefinitely, such as a persistent menu or a long-running interactive message.
func (sub *Subway) RegisterComponentListener(customID string, handler InteractionHandler) {
	sub.ComponentListenersMu.Lock()
	sub.ComponentListeners[customID] = &ComponentListener{
		Channel:            nil,
		InitialInteraction: discord.Interaction{},
		Handler:            handler,
	}
	sub.ComponentListenersMu.Unlock()
}

type ComponentListener struct {
	Channel            chan *discord.Interaction
	InitialInteraction discord.Interaction
	Handler            InteractionHandler

	createdAt time.Time
	expiresAt time.Time

	// Internal for easier cancellation
	subway *Subway
	key    string
}

// Cancel stops listening for a component and closes the channel, if one is present.
func (listener *ComponentListener) Cancel() {
	if listener.Channel != nil {
		close(listener.Channel)

		listener.Channel = nil
	}

	listener.subway.ComponentListenersMu.Lock()
	delete(listener.subway.ComponentListeners, listener.key)
	listener.subway.ComponentListenersMu.Unlock()
}

// WaitForComponent allows you to wait for a specific component interaction. You can either
// use a callback function which is automatically handled or use a channel.
func (sub *Subway) HandleComponent(interaction discord.Interaction, customID string, timeout time.Duration, handler InteractionHandler) *ComponentListener {
	now := time.Now()

	listener := &ComponentListener{
		Channel:            nil,
		InitialInteraction: interaction,
		Handler:            handler,
		createdAt:          now,
		expiresAt:          now.Add(timeout),
		subway:             sub,
		key:                customID,
	}

	if handler == nil {
		listener.Channel = make(chan *discord.Interaction)
	}

	sub.ComponentListenersMu.RLock()

	existing, ok := sub.ComponentListeners[customID]
	if ok {
		existing.Cancel()
	}

	sub.ComponentListenersMu.RUnlock()

	sub.ComponentListenersMu.Lock()
	sub.ComponentListeners[customID] = listener
	sub.ComponentListenersMu.Unlock()

	return listener
}
