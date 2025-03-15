package internal

// Accelerator to create a subcommand group.
func NewSubcommandGroup(name, description string) *InteractionCommandable {
	return SetupInteractionCommandable(&InteractionCommandable{
		Name:        name,
		Description: description,

		Type: InteractionCommandableTypeSubcommandGroup,
	})
}
