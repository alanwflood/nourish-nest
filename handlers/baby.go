package handlers

import (
	"NourishNestApp/db"
	"NourishNestApp/model"
	"NourishNestApp/views/pages"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func getBaby(c echo.Context) error {
	page := pages.AddNewBabyForm()
	return render(c, http.StatusOK, page)
}

func newBaby(c echo.Context) (err error) {
	bm := new(model.NewBaby)

	if err := c.Bind(bm); err != nil {
		log.Print(err.Error())
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "failed to create new bundle of joy"}
	}

	currentUser := getSessionUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/signin")
	}

	dob, err := time.Parse("2006-01-02", bm.DateOfBirth)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "could not parse date of birth for new baby"}
	}

	b := model.Baby{
		Gender:      bm.Gender,
		FirstName:   bm.FirstName,
		LastName:    bm.LastName,
		DateOfBirth: dob,
		User:        *currentUser,
	}

	err = db.InsertBaby(&b)

	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "something went wrong creating baby"}
	}

	return c.Redirect(http.StatusFound, "/baby")
}

func getAllBabies(c echo.Context) error {
	currentUser := getSessionUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/signin")
	}
	babies, _ := db.GetBabiesByUser(currentUser)

	page := pages.BabyListPage(babies)
	return render(c, http.StatusOK, page)
}

func babies(e *echo.Echo) {
	g := e.Group("/baby", CurrentUserMiddleware)
	g.GET("", getBaby)
	g.GET("/all", getAllBabies)
	g.POST("", newBaby)
}
