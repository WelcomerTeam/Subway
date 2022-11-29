package internal

import "sync"

type CogInfo struct {
	Name        string
	Description string

	Meta interface{}
}

// Cog is the basic interface for any cog. This must provide information about the cog
// such as its name and description.
type Cog interface {
	CogInfo() *CogInfo
	RegisterCog(sub *Subway) error
}

// CogWithInteractionCommands is an interface for any cog that implements methods that return interaction commands.
type CogWithInteractionCommands interface {
	GetInteractionCommandable() *InteractionCommandable
}

// CogWithBotLoad is an interface for any cog that implements methods that run when a bot loads.
type CogWithBotLoad interface {
	BotLoad(sub *Subway)
}

// CogWithBotUnload is an interface for any cog that implements methods that run when a bot unloads.
type CogWithBotUnload interface {
	BotUnload(sub *Subway, wg *sync.WaitGroup)
}
