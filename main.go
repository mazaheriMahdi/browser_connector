package main

import (
	"browser-connector/models"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"log"
	"maps"
	"net/http"
)

var sessions = map[uuid.UUID]models.Session{}

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	//cdp, err := pw.Chromium.ConnectOverCDP("ws://185.220.227.26:9222/devtools/browser/08d387f7-6bb1-41c0-b2ff-4e8d0ed9c6e5")
	//if err != nil {
	//	return
	//}
	//defer cdp.Close()

	cdp, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer cdp.Close()

	r := gin.Default()
	r.POST("/Session", func(c *gin.Context) {
		id := uuid.New()
		context, err2 := cdp.NewContext()
		if err2 != nil {
			_ = c.AbortWithError(500, err2)
		}
		sessions[id] = models.Session{
			SessionId: id,
			Context:   context,
		}
		c.JSON(200, gin.H{
			"sessionId": id,
		})
	})

	r.POST("/Session/:id/Goto", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))
		dto := models.GotoDto{}
		if errA := c.ShouldBind(&dto); errA == nil {

			session, ok := sessions[id]
			if !ok {
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			}
			page, err := session.Context.NewPage()
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			_, err = page.Goto(dto.Url)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			err = page.WaitForLoadState()
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			content, err := page.Content()
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}

			c.JSON(200, gin.H{
				"content": content,
			})

		} else {
			c.AbortWithError(http.StatusBadRequest, errA)
		}
	})

	r.POST("/Session/:id/ImplicitWait", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))
		dto := models.ImplicitWaitDto{}
		if errA := c.ShouldBind(&dto); errA == nil {

			session, ok := sessions[id]
			if !ok {
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			}
			session.Context.Pages()[0].WaitForTimeout(dto.Seconds * 1000)
			c.String(200, "done")

		} else {
			c.AbortWithError(http.StatusBadRequest, errA)
		}
	})

	r.GET("/Session/:id/Content", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))

		session, ok := sessions[id]
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
		}
		content, err := session.Context.Pages()[0].Content()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.JSON(200, gin.H{
			"content": content,
		})

	})

	r.GET("health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/Session", func(c *gin.Context) {
		result := []string{}
		for sessionId := range maps.Keys(sessions) {
			result = append(result, sessionId.String())
		}
		c.JSON(http.StatusOK, gin.H{
			"sessions": result,
			"contexts": len(cdp.Contexts()),
		})
	})

	r.Run("0.0.0.0:8081") // listen and serve on 0.0.0.0:8080
}
