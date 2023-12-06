package internal

// Accelerator to create a subcommand group.
func NewSubcommandGroup(name string, description string) *InteractionCommandable {
	return SetupInteractionCommandable(&InteractionCommandable{
		Name:        name,
		Description: description,

		Type: InteractionCommandableTypeSubcommandGroup,
	})
}
