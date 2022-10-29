package main

import (
	"container_manager/client"
	"container_manager/listen"
	"container_manager/route"
	"container_manager/service"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	s := service.NewK8sService()
	resource, err, _ := s.GetResource()
	if err != nil {
		log.Fatalln(err)
	}

	l := listen.NewListen(client.GetClient(), client.GetK8sConfig(), client.GetDyna(), resource)
	go l.ListenResource()

	r := gin.Default()
	route.Route(r)

	err = r.Run(":8000")
	if err != nil {
		log.Fatal(err)
	}
}
