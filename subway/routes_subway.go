package internal

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	verified := subway.verifySignature(r, body)
	if !verified {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	var interaction discord.Interaction

	defer func() {
		elapsed := float64(time.Since(start)) / float64(time.Second)

		var guildID string

		var userID string

		if interaction.GuildID != nil {
			guildID = strconv.FormatInt(int64(*interaction.GuildID), 10)
		}

		if interaction.User != nil {
			userID = strconv.FormatInt(int64(interaction.User.ID), 10)
		}

		subwayInteractionProcessingTimeName.WithLabelValues(interaction.Data.Name, guildID, userID).Observe(elapsed)
	}()

	err = jsoniter.Unmarshal(body, &interaction)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if interaction.Type == discord.InteractionTypePing {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(InteractionPongResponse)

		return
	}

	response, err := subway.ProcessInteraction(interaction)

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
