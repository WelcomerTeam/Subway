package internal

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	jsoniter "github.com/json-iterator/go"
)

var InteractionPongResponse = []byte(`{"type":1}`)

func (subway *Subway) HandleSubwayRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	start := time.Now()

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		subway.Logger.Warn().Err(err).Msg("Failed to read body")

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	verified := subway.verifySignature(r, body)
	if !verified {
		subway.Logger.Warn().Msg("Sender passed invalid signature")

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

	err = jsoniter.Unmarshal(body, &interaction)
	if err != nil {
		subway.Logger.Warn().Err(err).Msg("Failed to parse interaction")

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if interaction.Type == discord.InteractionTypePing {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(InteractionPongResponse)

		return
	}

	response, err := subway.ProcessInteraction(
		subway.NewInteractionContext(subway.Context),
		interaction,
	)

	var guildID string

	var userID string

	if interaction.GuildID != nil {
		guildID = strconv.FormatInt(int64(*interaction.GuildID), 10)
	}

	if interaction.User != nil {
		userID = strconv.FormatInt(int64(interaction.User.ID), 10)
	}

	subwayInteractionTotal.WithLabelValues(interaction.Data.Name, guildID, userID).Add(1)

	if err != nil {
		subway.Logger.Warn().Err(err).Send()

		subwayFailedInteractionTotal.Add(1)

		w.WriteHeader(http.StatusNoContent)

		return
	}

	subwaySuccessfulInteractionTotal.Add(1)

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func (subway *Subway) NewInteractionContext(ctx context.Context) context.Context {
	return AddSubwayToContext(ctx, subway)
}

func (subway *Subway) NewGRPCContext(ctx context.Context) *sandwich.GRPCContext {
	return &sandwich.GRPCContext{
		Context:        ctx,
		Logger:         subway.Logger,
		SandwichClient: subway.SandwichClient,
		GRPCInterface:  subway.GRPCInterface,
	}
}
