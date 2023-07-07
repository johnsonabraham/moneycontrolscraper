package auth

import (
	"strings"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"github.com/sirupsen/logrus"
)

type UserClaims struct {
	User string `json:"user"`
}

// Auth godoc
//
//	@Summary		API Auth
//	@Description	Auth endpoint using x-api-key to get token with time bound validation
//	@Accept			json
//	@Produce		text/plain
//	@Success		200	{string}	string
//	@Failure		400	{object}	models.FailedResponse
//	@Failure		401	{object}	models.FailedResponse
//	@Failure		404	{object}	models.FailedResponse
//	@Failure		500	{object}	models.FailedResponse
//	@Security		ApiKeyAuth
//	@in	 header
//	@name x-api-key
//	@Router	/auth [get]
func GenerateToken(signer *jwt.Signer, cfg *config.AppEnvVars) iris.Handler {
	return func(ctx iris.Context) {
		switch ctx.Method() {
		case "GET":
			key := ctx.GetHeader("x-api-key")
			if key == "" {
				failedResp := models.FailedResponse{
					Status:   iris.StatusInvalidToken,
					ErrorMsg: "x-api-key is missing",
				}
				ctx.StopWithJSON(iris.StatusInvalidToken, failedResp)

				return
			}

			key = strings.Trim(key, "\"")

			if key != cfg.APIKey {
				ctx.StopWithStatus(iris.StatusInvalidToken)

				return
			}

			claims := UserClaims{User: "X-API-KEY"}

			token, err := signer.Sign(claims)
			if err != nil {
				failedresp := models.FailedResponse{
					Status:   iris.StatusInternalServerError,
					ErrorMsg: "token expired",
				}
				ctx.StopWithJSON(iris.StatusInternalServerError, failedresp)

				return
			}

			ctx.Write(token)
		default:
			failedResp := models.FailedResponse{
				Status:   iris.StatusMethodNotAllowed,
				ErrorMsg: "http method not allowed, only get is supported",
			}
			ctx.StopWithJSON(iris.StatusMethodNotAllowed, failedResp)
		}
	}
}

func isAuthorized(ctx iris.Context) {
	claims, ok := jwt.Get(ctx).(*UserClaims)
	if !ok {
		logrus.Error("invalid type passed, instead of UserClaims")
	}

	standardClaims := jwt.GetVerifiedToken(ctx).StandardClaims
	expiresAtString := standardClaims.ExpiresAt().
		Format(ctx.Application().ConfigurationReadOnly().GetTimeFormat())
	timeLeft := standardClaims.Timeleft()

	ctx.Writef("User=%s\nexpires at: %s\ntime left: %s\n", claims.User, expiresAtString, timeLeft)
}
