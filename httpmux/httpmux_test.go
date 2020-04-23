package httpmux

import (
	"aicode"
	"log"
	"net/http"
	"testing"
)

type Body struct {
	Opt     string
	Channel string `json:"chan" valid:"required"`
}

func Index(c Context) aicode.HTTPError {
	// c.String(http.StatusOK, "heelo")
	st := new(Body)
	if err := c.Bind(st); err != nil {
		log.Println(err)
	}
	if err := c.Validate(st); err != nil {
		log.Println(err)
	}
	return aicode.ComBadParam
	v := make(map[string]string)
	v["body"] = "hello"
	c.JSON(http.StatusOK, v)
	return nil
}

func Test_Http(t *testing.T) {
	srv := New()
	srv.POST("/", Index)

	http.ListenAndServe(":9000", srv)
}
