package internal

import (
	"net/http"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/gin-gonic/gin"
)

func registerRoutes(g *gin.Engine) {
	// GET / returns subway information.
	g.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Subway VERSION "+VERSION)
	})

	// POST / handles interactions.
	g.POST("/", func(ctx *gin.Context) {
		verifySignature(ctx, subway.publicKey, func(ctx *gin.Context) {
			var interaction discord.Interaction
			start := time.Now()

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

			err := ctx.BindJSON(&interaction)
			if err != nil {
				ctx.String(http.StatusBadRequest, err.Error())

				return
			}

			if interaction.Type == discord.InteractionTypePing {
				ctx.JSON(http.StatusOK, discord.InteractionResponse{
					Type: discord.InteractionCallbackTypePong,
					Data: nil,
				})

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
				ctx.JSON(http.StatusInternalServerError, err.Error())
			} else {
				subwaySuccessfulInteractionTotal.Add(1)
				ctx.JSON(http.StatusOK, response)
			}
		})
	})
}
