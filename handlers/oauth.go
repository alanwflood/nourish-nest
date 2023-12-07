package handlers

import (
	"NourishNestApp/db"
	"NourishNestApp/logger"
	"NourishNestApp/model"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func OAuthHandlers(e *echo.Echo) {
	logger.Log.Info("Origin: " + os.Getenv("ORIGIN"))

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_SECRET"), "http://localhost:42069/auth/google/callback", "email", "profile"),
	)

	store := createGothicCookieStore()
	gothic.Store = store

	e.GET("/auth/:provider/callback", func(c echo.Context) error {
		assignProviderToRequestContext(c)
		gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		return onProviderUserFound(gothUser, c)
	})

	e.GET("/logout/:provider", func(c echo.Context) error {
		gothic.Logout(c.Response(), c.Request())
		_, err := deleteCookie(c)
		if err != nil {
			logger.Log.Error(err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		logger.Log.Info("Logging out user")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	e.GET("/auth/:provider", func(c echo.Context) error {
		assignProviderToRequestContext(c)
		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
			logger.Log.Info(gothUser.Provider)
			return onProviderUserFound(gothUser, c)
		} else {
			gothic.BeginAuthHandler(c.Response(), c.Request())
			return nil
		}
	})

	e.GET("/redirect", func(c echo.Context) error {
		logger.Log.Info("Redirecting user")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})
}

func onProviderUserFound(pu goth.User, c echo.Context) error {
	sesh, _ := getCookie(c)

	uid := pu.Provider + "_" + pu.UserID
	foundUser, err := db.GetUserByToken(pu.AccessToken)

	if foundUser == nil || err != nil {
		u := model.User{
			Id:        uid,
			Token:     pu.AccessToken,
			Name:      pu.Name,
			Email:     pu.Email,
			UpdatedAt: time.Now(),
		}

		logger.Log.Info("No user found Upserting user: " + u.Id)
		err := db.UpsertUser(u)
		if err != nil {
			logger.Log.Error(err.Error())
			return c.Redirect(http.StatusFound, "/")
		}
		foundUser = &u
	} else {
		logger.Log.Info("Found user: " + foundUser.Id)
	}

	sesh.Values["current_user"] = foundUser.Id
	err = sesh.Save(c.Request(), c.Response())

	if err != nil {
		logger.Log.Error("Error saving session: " + err.Error())
	}

	logger.Log.Info("Saved user to session: " + foundUser.Id)
	return c.Redirect(http.StatusFound, "/redirect")
}

func createGothicCookieStore() *sessions.CookieStore {
	key := []byte(os.Getenv("SESSION_SECRET"))
	store := sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   true,
	}
	return store
}

func getCookie(c echo.Context) (sesh *sessions.Session, err error) {
	sesh, err = session.Get("nn-session", c)
	return sesh, err
}

func deleteCookie(c echo.Context) (sesh *sessions.Session, err error) {
	sesh, err = session.Get("nn-session", c)
	if err != nil {
		return sesh, err
	}
	sesh.Options.MaxAge = -1
	err = sesh.Save(c.Request(), c.Response())
	return sesh, err
}

func assignProviderToRequestContext(c echo.Context) {
	nc := context.WithValue(c.Request().Context(), "provider", c.Param("provider"))
	nr := c.Request().WithContext(nc)
	c.SetRequest(nr)
}
