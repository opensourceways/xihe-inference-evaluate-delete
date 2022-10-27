package main

import (
	"container_manager/client"
	"container_manager/route"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()
	route.Route(r)

	err = r.Run(":8000")
	if err != nil {
		log.Fatal(err)
	}
}
