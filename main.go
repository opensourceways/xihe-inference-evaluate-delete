package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"container_manager/client"
	"container_manager/listen"
	"container_manager/route"
	"container_manager/service"
	"github.com/gin-gonic/gin"
	liboptions "github.com/opensourceways/community-robot-lib/options"
)

type options struct {
	service     liboptions.ServiceOptions
	enableDebug bool
}

func (o *options) Validate() error {
	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.BoolVar(
		&o.enableDebug, "enable_debug", false,
		"whether to enable debug model.",
	)

	fs.Parse(args)
	return o
}

func main() {
	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		log.Println("Invalid options, err:", err.Error())
	}

	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	s := service.NewK8sService()
	resource, err, _ := s.GetResource()
	if err != nil {
		log.Fatalln(err)
	}

	nConfig := new(listen.Config)
	if err := listen.LoadConfig(o.service.ConfigFile, nConfig); err != nil {
		log.Fatalln("load config failed:", err.Error())
	}

	l, err := listen.NewListen(client.GetClient(), client.GetK8sConfig(), client.GetDyna(), resource, nConfig)
	if err != nil {
		log.Fatalln(err)
	}
	go l.ListenResource()

	r := gin.Default()
	route.Route(r)

	err = r.Run(":" + strconv.Itoa(o.service.Port))
	if err != nil {
		log.Fatal(err)
	}
}
