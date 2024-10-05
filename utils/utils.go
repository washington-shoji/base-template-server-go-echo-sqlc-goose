package utils

import (
	"github.com/labstack/echo/v4"
)

func RespondWithError(ctx echo.Context, code int, msg string) error {

	type errResponse struct {
		Error string `json:"error"`
	}

	return ctx.JSON(code, errResponse{
		Error: msg,
	})

}

func RespondWithJSON(ctx echo.Context, code int, payload interface{}) error {
	return ctx.JSON(code, payload)
}
