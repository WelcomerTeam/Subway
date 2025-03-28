package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

var InteractionPongResponse = []byte(`{"type":1}`)

func (sub *Subway) HandleSubwayRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	start := time.Now()

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sub.Logger.Warn().Err(err).Msg("Failed to read body")

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	verified := sub.verifySignature(r, body)
	if !verified {
		sub.Logger.Warn().Msg("Sender passed invalid signature")

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	var interaction discord.Interaction

	defer func() {
		elapsed := float64(time.Since(start)) / float64(time.Second)

		var commandName string

		var guildID string

		var userID string

		if interaction.Data != nil {
			commandName = interaction.Data.Name
		}

		if interaction.GuildID != nil {
			guildID = strconv.FormatInt(int64(*interaction.GuildID), 10)
		}

		if interaction.User != nil {
			userID = strconv.FormatInt(int64(interaction.User.ID), 10)
		}

		subwayInteractionProcessingTimeName.WithLabelValues(commandName, guildID, userID).Observe(elapsed)
	}()

	err = json.Unmarshal(body, &interaction)
	if err != nil {
		sub.Logger.Warn().Err(err).Msg("Failed to parse interaction")

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	var guildID string

	var userID string

	if interaction.Type == discord.InteractionTypePing {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(InteractionPongResponse)

		return
	}

	var response *discord.InteractionResponse

	ctx := sub.Context
	ctx = AddURLToContext(ctx, *r.URL)

	switch interaction.Type {
	case discord.InteractionTypeApplicationCommand, discord.InteractionTypeApplicationCommandAutocomplete:
		response, err = sub.ProcessApplicationCommandInteraction(ctx, interaction)
	case discord.InteractionTypeMessageComponent:
		response, err = sub.ProcessMessageComponentInteraction(ctx, interaction)
	// case discord.InteractionTypeModalSubmit:
	// 	// not implemented
	default:
		sub.Logger.Warn().Int("interaction_type", int(interaction.Type)).Msg("Missing interaction handler")
	}

	if interaction.GuildID != nil {
		guildID = strconv.FormatInt(int64(*interaction.GuildID), 10)
	}

	if interaction.User != nil {
		userID = strconv.FormatInt(int64(interaction.User.ID), 10)
	}

	defer func() {
		if err != nil {
			subwayFailedInteractionTotal.Add(1)
		} else {
			subwaySuccessfulInteractionTotal.Add(1)
		}
	}()

	subwayInteractionTotal.WithLabelValues(interaction.Data.Name, guildID, userID).Add(1)

	if err != nil {
		sub.Logger.Error().Err(err).Msg("Failed to process interaction")

		w.WriteHeader(http.StatusNoContent)

		return
	}

	if response != nil {
		resp, err := json.Marshal(response)
		if err != nil {
			sub.Logger.Warn().Err(err).Msg("Failed to marshal response")

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}

		w.Header().Add("Content-Type", "application/json")

		_, err = w.Write(resp)
		if err != nil {
			sub.Logger.Warn().Err(err).Msg("Failed to write response")
		}
	} else {
		sub.Logger.Warn().Msg("No response to send")

		w.WriteHeader(http.StatusNoContent)
	}
}

func (sub *Subway) NewGRPCContext(ctx context.Context) *sandwich.GRPCContext {
	return &sandwich.GRPCContext{
		Context:        ctx,
		Logger:         sub.Logger,
		SandwichClient: sub.SandwichClient,
		GRPCInterface:  sub.GRPCInterface,
	}
}
