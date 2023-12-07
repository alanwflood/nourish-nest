package handlers

import (
	"NourishNestApp/db"
	"NourishNestApp/logger"
	"NourishNestApp/model"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func getSessionUser(c echo.Context) *model.User {
	sesh, err := getCookie(c)
	if err != nil {
		logger.Log.Error("Failed to get session: " + err.Error())
		return nil
	}

	userId, ok := sesh.Values["current_user"].(string)
	logger.Log.Info("User ID: " + userId)
	if !ok {
		logger.Log.Info(fmt.Sprintf("Authenticated session not found: ", sesh.Values))
		return nil
	}

	user, err := db.GetUserById(userId)
	if err != nil {
		logger.Log.Info(fmt.Sprintf("Authenticated user not found for user with Id: %s", userId))
		return nil
	}

	logger.Log.Info(fmt.Sprintf("Authenticated user found: %s", user.Email))
	return user
}

type CustomContext struct {
	echo.Context
	CurrentUser *model.User
}

func CurrentUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := getSessionUser(c)
		if user == nil {
			logger.Log.Info("User is not authorized for route: " + c.Request().RequestURI)
			return c.Redirect(http.StatusFound, "/signin")
		} else {
			logger.Log.Info("User is currently connected with id: " + user.Id)
		}

		cc := &CustomContext{c, user}
		return next(cc)
	}
}

func UseSessionMiddleware(e *echo.Echo) {
	key := []byte(os.Getenv("SESSION_SECRET"))
	store := sessions.NewCookieStore([]byte(key))
	e.Use(session.Middleware(store))
}
