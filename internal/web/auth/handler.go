package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rbrick/aoc-runner/internal/session"
)

type Handler struct {
	session *session.Session
}

func (h *Handler) status(reqCtx *echo.Context) error {
	if !h.session.LoggedIn() {
		return reqCtx.JSON(http.StatusOK, &SessionDTO{
			LoggedIn: false,
			LoginURL: session.LoginURL,
		})
	}

	return reqCtx.JSON(http.StatusOK, &SessionDTO{
		LoggedIn: true,
	})
}

// reads the Advent of Code session from the browsers on this machine
func (h *Handler) login(reqCtx *echo.Context) error {
	ctx := reqCtx.Request().Context()

	err := h.session.Login(ctx)

	if errors.Is(err, session.ErrNotLoggedIn) {
		// the browser client should send the user to the login url, then try again
		return reqCtx.JSON(http.StatusUnauthorized, &SessionDTO{
			LoggedIn: false,
			LoginURL: session.LoginURL,
			Message:  err.Error(),
		})
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return reqCtx.JSON(http.StatusOK, &SessionDTO{
		LoggedIn: true,
	})
}

func NewHandler(session *session.Session) *Handler {
	return &Handler{
		session: session,
	}
}
