package main

import (
	"container_manager/client"
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

	s := new(service.K8sService)
	resource, err, _ := s.GetResource()
	if err != nil {
		log.Fatalln(err)
	}

	listen := service.NewListen(client.GetClient(), client.GetK8sConfig(), client.GetDyna(), resource)
	go listen.ListenResource()

	r := gin.Default()
	route.Route(r)

	err = r.Run(":8000")
	if err != nil {
		log.Fatal(err)
	}
}
