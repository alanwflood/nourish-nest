package handlers

import (
	"NourishNestApp/db"
	"NourishNestApp/logger"
	"NourishNestApp/views/components"
	"NourishNestApp/views/pages"
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func renderAllEntries(c echo.Context) error {
	pageSize := 50
	pageNumber := getPaginationFromParams(c)
	nextPage := pageNumber + 1

	nextSessionTime := db.GetNextSessionStartTime()

	entries := db.GetAllEntries(pageSize, (pageNumber-1)*pageSize)
	component := pages.ViewAllEntries(entries, nextPage, nextSessionTime)

	return render(c, http.StatusOK, component)
}

func dailySummaries(e *echo.Echo) {
	e.GET("/daily", func(c echo.Context) error {
		pageSize := 7
		pageNumber := getPaginationFromParams(c)
		nextPage := pageNumber + 1
		summaries := db.GetDailySummaries(pageSize, (pageNumber-1)*pageSize)
		component := pages.ViewDailySummaries(summaries, nextPage)
		return render(c, http.StatusOK, component)
	})
}

func getPaginationFromParams(c echo.Context) int {
	page := c.QueryParam("p")
	pageNumber, _ := strconv.Atoi(page)
	if pageNumber == 0 {
		pageNumber = 1
	}

	return pageNumber
}

func entries(e *echo.Echo) {
	e.GET("/", renderAllEntries, CurrentUserMiddleware)

	g := e.Group("/entry", CurrentUserMiddleware)

	g.GET("/all", renderAllEntries)

	g.GET("/:id/edit", func(c echo.Context) error {
		id := c.Param("id")
		entry := db.GetEntryById(id)
		if entry == nil {
			return c.String(http.StatusNotFound, "failed to find entry")
		}
		component := components.EditEntryDialog(entry)
		return render(c, http.StatusOK, component)
	})

	g.GET("/:id/feed/:feedId/edit", func(c echo.Context) error {
		id := c.Param("id")
		entry := db.GetEntryById(id)
		if entry == nil {
			return c.String(http.StatusNotFound, "failed to find entry")
		}

		feedId, _ := strconv.Atoi(c.Param("feedId"))
		feed := entry.GetFeedByFeedId(feedId)
		if feed == nil {
			return c.String(http.StatusNotFound, "failed to find feed")
		}

		component := components.EditEntryFeedDialog(entry, feed)
		return render(c, http.StatusOK, component)
	})

	g.DELETE("/:id", func(c echo.Context) error {
		id := c.Param("id")
		ok := db.DeleteEntryById(id)
		if ok {
			return c.HTML(http.StatusOK, "")
		} else {
			return c.NoContent(http.StatusExpectationFailed)
		}
	})

	g.GET("", func(c echo.Context) error {
		component := pages.NewEntry()
		return render(c, http.StatusOK, component)
	})

	g.POST("", func(c echo.Context) error {
		action := c.QueryParam("action")
		var dirtyStateInt int

		if c.FormValue("dirty") == "" {
			dirtyStateInt = 0
		} else {
			dirtyStateInt, _ = strconv.Atoi(c.FormValue("dirty"))
		}

		id := uuid.New().String()
		newEntry := db.Entry{
			Id:              id,
			Notes:           c.FormValue("notes"),
			NappyStateWet:   c.FormValue("wet") != "",
			NappyStateDirty: dirtyStateInt,
		}

		db.UpsertEntry(newEntry)

		if action == "finish" {
			return c.Redirect(http.StatusFound, "/entry/all#entry-"+id)
		}

		return c.Redirect(http.StatusFound, "/entry/"+id+"/feed")
	})

	g.PUT("/:id", func(c echo.Context) error {
		entryId := c.Param("id")
		entry := db.GetEntryById(entryId)
		if entry == nil {
			return c.String(http.StatusNotFound, "failed to find entry")
		}

		var dirtyStateInt int

		if c.FormValue("dirty") == "" {
			dirtyStateInt = 0
		} else {
			dirtyStateInt, _ = strconv.Atoi(c.FormValue("dirty"))
		}

		entry.Notes = c.FormValue("notes")
		entry.NappyStateWet = c.FormValue("wet") != ""
		entry.NappyStateDirty = dirtyStateInt

		db.UpsertEntry(*entry)

		component := components.EntryCard(*entry)
		return render(c, http.StatusOK, component)
	})

	g.GET("/:id/feed", func(c echo.Context) error {
		id := c.Param("id")
		entry := db.GetEntryById(id)
		lastLoggedSide := db.GetLastLoggedSide()
		component := pages.NewEntryFeed(entry, lastLoggedSide)
		return render(c, http.StatusOK, component)
	})

	g.POST("/:id/feed", func(c echo.Context) error {
		action := c.QueryParam("action")
		id := c.Param("id")

		entry := db.GetEntryById(id)
		if entry == nil {
			return c.String(http.StatusNotFound, "failed to find entry")
		}

		newFeed := createFeedFromRequestContext(c)
		log.Printf("Creating new feed with time: " + newFeed.StartTime.String())
		db.CreateFeedForEntry(entry, newFeed)

		if action == "finish" {
			return c.Redirect(http.StatusFound, "/entry/all#entry-"+id)
		}

		return c.Redirect(http.StatusFound, "/entry/"+id+"/feed")
	})

	g.PUT("/:id/feed/:feedId", func(c echo.Context) error {
		entryId := c.Param("id")
		feedId, _ := strconv.Atoi(c.Param("feedId"))
		feed := db.GetFeedByEntryIdAndFeedId(entryId, feedId)

		if feed == nil {
			return c.String(http.StatusNotFound, "failed to find feed")
		}

		assignFormValuesToFeed(feed, c)
		db.UpdateFeed(feed)

		entry := db.GetEntryById(entryId)
		component := components.FeedCardsList(*entry)
		return render(c, http.StatusOK, component)
	})

	g.DELETE("/:id/feed/:feedId", func(c echo.Context) error {
		entryId := c.Param("id")
		feedId, _ := strconv.Atoi(c.Param("feedId"))
		db.DeleteFeedByEntryIdAndFeedId(entryId, feedId)
		ok := db.DeleteFeedByEntryIdAndFeedId(entryId, feedId)
		if ok {
			entry := db.GetEntryById(entryId)
			component := components.FeedCardsList(*entry)
			return render(c, http.StatusOK, component)
		} else {
			return c.NoContent(http.StatusExpectationFailed)
		}
	})
}

func convertTimeFormValueToTime(t time.Time, timeValue string) time.Time {
	timeValues := strings.Split(timeValue, ":")
	timeHour, _ := strconv.Atoi(timeValues[0])
	timeMinute, _ := strconv.Atoi(timeValues[1])
	return time.Time(time.Date(t.Year(), t.Month(), t.Day(), timeHour, timeMinute, 0, 0, t.Location()))
}

func convertStringFormValueToTime(fv string) time.Time {
	i, _ := strconv.ParseInt(fv, 10, 64)
	return time.UnixMilli(i)
}

func assignFormValuesToFeed(feed *db.Feed, c echo.Context) {
	feed.Side = c.FormValue("side")

	dateStarted := convertStringFormValueToTime(c.FormValue("dateStarted"))
	startTime := convertTimeFormValueToTime(dateStarted, c.FormValue("timeStarted"))
	endTime := convertTimeFormValueToTime(dateStarted, c.FormValue("timeStopped"))

	// If Endtime rolled the clock over to midnight, just add a day to the date started
	if endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	feed.CreatedAt = time.Now()
	feed.StartTime = startTime
	feed.EndTime = endTime
}

func createFeedFromRequestContext(c echo.Context) db.Feed {
	var newFeed db.Feed

	newFeed.Id = 0
	assignFormValuesToFeed(&newFeed, c)
	log.Printf("New feed parsed\n\nStarted: %s\nStopped: %s", newFeed.StartTime.String(), newFeed.EndTime.String())

	return newFeed
}

func users(e *echo.Echo) {
	e.GET("/signin", func(c echo.Context) error {
		user := getSessionUser(c)
		if user != nil {
			return c.Redirect(http.StatusFound, "/")
		} else {
			logger.Log.Info("No user found from session")
		}

		hasErr := c.QueryParam("invalid")
		showErr := hasErr == "1"
		page := pages.UserSignIn(showErr)
		return render(c, http.StatusOK, page)
	})
}

func HandleRoutes(e *echo.Echo) {
	OAuthHandlers(e)
	dailySummaries(e)
	entries(e)
	users(e)
	babies(e)
}

func render(ctx echo.Context, status int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(status)

	err := t.Render(context.Background(), ctx.Response().Writer)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "failed to render response template")
	}

	return nil
}
