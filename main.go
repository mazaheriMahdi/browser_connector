package main

import (
	"browser-connector/models"
	"errors"
	"fmt"
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
	//cdp, err := pw.Chromium.ConnectOverCDP("ws://127.0.0.1:9222/devtools/browser/816bcf2b-4ae3-4e8c-86b1-a2632a86df5b")
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
		_, err := context.NewPage()
		if err != nil {
			_ = c.AbortWithError(500, err2)
			return
		}
		if err2 != nil {
			_ = c.AbortWithError(500, err2)
			return
		}
		sessions[id] = models.Session{
			SessionId:  id,
			Context:    context,
			ActivePage: 0,
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
			page := session.Context.Pages()[session.ActivePage]
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
			session.Context.Pages()[session.ActivePage].WaitForTimeout(dto.Seconds * 1000)
			c.String(200, "done")

		} else {
			c.AbortWithError(http.StatusBadRequest, errA)
		}
	})

	r.POST("/Session/:id/Clean", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))

		session, ok := sessions[id]
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			return
		}
		session.Context.Pages()[session.ActivePage].Close()
		session.ActivePage--
		_, err := session.Context.NewPage()
		if err != nil {
			return
		}
		c.String(200, "done")

	})

	r.POST("/Session/:id/Scroll", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))
		dto := models.ScrollDto{}
		if errA := c.ShouldBind(&dto); errA == nil {
			session, ok := sessions[id]
			if !ok {
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
				return
			}
			_, err := session.Context.Pages()[session.ActivePage].Evaluate(fmt.Sprintf("window.scroll(%d,%d)", dto.X, dto.Y))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "can't scroll to given position",
				})
				return
			}
			c.String(200, "done")

		} else {
			c.AbortWithError(http.StatusBadRequest, errA)
		}
	})

	r.GET("/Session/:id/Screenshot", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))

		session, ok := sessions[id]
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			return
		}
		screenshot, err := session.Context.Pages()[session.ActivePage].Screenshot()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "can't scroll to given position",
			})
			return
		}
		c.Data(200, "image/png", screenshot)

	})

	r.GET("/Session/:id/Content", func(c *gin.Context) {
		id := uuid.MustParse(c.Param("id"))

		session, ok := sessions[id]
		if !ok {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session not found"))
			return
		}
		content, err := session.Context.Pages()[session.ActivePage].Content()
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

		page := session.Context.Pages()[session.ActivePage]
		if page.IsClosed() {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("session is already closed"))
			return
		} else {
			err := page.Close()
			delete(sessions, id)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errors.New("can't close session"))
				return
			}
			err = session.Context.Close()
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, errors.New("can't close context"))
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
