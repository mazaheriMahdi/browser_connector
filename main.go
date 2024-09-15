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
	cdp, err := pw.Chromium.Launch()
	//cdp, err := pw.Chromium.ConnectOverCDP("ws://127.0.0.1:9222/devtools/browser/2204f766-3ca9-448e-9c31-5aa9561c93e8")
	if err != nil {
		log.Fatalf("could not start chrome: %v", err)
	}
	defer cdp.Close()

	//cdp, err := pw.Chromium.Launch()
	//if err != nil {
	//	log.Fatalf("could not launch browser: %v", err)
	//}
	//defer cdp.Close()

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
				return
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
			return
		}
		content, err := session.Context.Pages()[0].Content()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.JSON(200, gin.H{
			"content": content,
		})
	})

	r.DELETE("/Session/:id", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))

		session, ok := sessions[id]
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			return
		}

		page := session.Context.Pages()[0]
		if page.IsClosed() {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session is already closed"))
			return
		} else {
			err := page.Close()
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errors.New("can't close session"))
				return
			}
		}

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.JSON(200, gin.H{
			"status": "done",
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
