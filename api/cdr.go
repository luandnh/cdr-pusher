package api

import (
	"cdr-pusher/common/log"
	"cdr-pusher/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Cdr struct {
	cdrService service.CDR
}

func APICdr(r *gin.Engine, cdr service.CDR) {
	handler := &Cdr{
		cdrService: cdr,
	}
	Group := r.Group("v1/sbclog")
	{
		Group.POST("", handler.CreateCDR)
	}
}
func (h *Cdr) CreateCDR(c *gin.Context) {
	if c.Request.Body == nil {
		log.Error("request body null")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": http.StatusText(http.StatusUnprocessableEntity),
		})
		c.Abort()
		return
	}
	body := map[string]any{}
	if err := c.BindJSON(&body); err != nil {
		log.Error(err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": http.StatusText(http.StatusUnprocessableEntity),
		})
		return
	}
	go func() {
		if err := h.cdrService.PostSBCLog(body); err != nil {
			log.Info("Post to API err : ", err)
		}
	}()
	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "success",
	})
}
