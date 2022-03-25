package api

import (
	"cdr-pusher/common/log"
	"cdr-pusher/service"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Cdr struct {
	cdrService service.ICdr
}

func APICdr(r *gin.Engine, cdr service.ICdr) {
	handler := &Cdr{
		cdrService: cdr,
	}
	Group := r.Group("v1/cdr")
	{
		Group.POST("", handler.CreateCDR)
	}
}
func (h *Cdr) CreateCDR(c *gin.Context) {
	if c.Request.Body == nil {
		log.Error("Cdr", "CreateCDR", "Request body null")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": http.StatusText(http.StatusUnprocessableEntity),
		})
		c.Abort()
		return
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Cdr", "CreateCDR", err.Error())
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": http.StatusText(http.StatusUnprocessableEntity),
		})
		c.Abort()
		return
	}
	go func() {
		if err := h.cdrService.HandlePostXmlToAPI(body); err != nil {
			log.Info("Post to API err : ", err)
		}
	}()
	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "success",
	})
}
