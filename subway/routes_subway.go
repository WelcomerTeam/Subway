package internal

import (
	"net/http"

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

			err := ctx.BindJSON(&interaction)
			if err != nil {
				ctx.String(http.StatusBadRequest, err.Error())

				return
			}

			if interaction.Type == discord.InteractionTypePing {
				ctx.JSON(http.StatusOK, discord.InteractionResponse{
					Type: discord.InteractionCallbackTypePong,
				})

				return
			}

			response, err := subway.ProcessInteraction(interaction)

			if err != nil {
				subway.Logger.Warn().Err(err).Send()

				ctx.JSON(http.StatusInternalServerError, err.Error())
			} else {
				ctx.JSON(http.StatusOK, response)
			}
		})
	})
}
