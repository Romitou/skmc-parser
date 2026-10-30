package handlers

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/romitou/skriptmc-parser/structures"
)

func Parse(req structures.HTTPRequest) {
	req.Gin.Header("Access-Control-Allow-Headers", "Content-Type")
	req.Gin.Status(http.StatusOK)
}

func PostParse(req structures.HTTPRequest) {
	requestBody, err := io.ReadAll(req.Gin.Request.Body)
	if err != nil {
		req.Gin.JSON(http.StatusBadRequest, gin.H{
			"code":  2002,
			"error": "Your request body is badly formed or invalid. Correct your request and try again.",
		})
		return
	}

	containers := req.Ctx.Cache.Containers

	if len(containers) == 0 {
		req.Gin.JSON(http.StatusServiceUnavailable, gin.H{
			"code":  2003,
			"error": "No parser is currently able to process your request.",
		})
		return
	}

	container := containers[rand.Intn(len(containers))]

	req.Gin.Header("X-Parser", container.Names[0])

	parserNetwork, ok := container.NetworkSettings.Networks["skmc-parser"]
	if !ok {
		req.Gin.JSON(http.StatusInternalServerError, gin.H{
			"code":  2004,
			"error": "A parser was found, but seems to be misconfigured. Retry your request.",
		})
		return
	}

	response, err := http.Post("http://"+parserNetwork.IPAddress+"/parse", "text/plain", bytes.NewBuffer(requestBody))
	if err != nil || response.StatusCode != http.StatusOK {
		req.Gin.JSON(http.StatusInternalServerError, gin.H{
			"code":  2005,
			"error": "Your request has been sent to a parser but it seems to return an error. Try your request again.",
		})
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		req.Gin.JSON(http.StatusInternalServerError, gin.H{
			"code":  2006,
			"error": "An error occurred while reading the parser response. Try your request again.",
		})
		return
	}

	req.Gin.Data(http.StatusOK, "application/json", body)
}
