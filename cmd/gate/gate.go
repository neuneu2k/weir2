/*
Copyright 2016 Assoba S.A.S.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/neuneu2k/weir2"
	"os"
	"os/signal"
	"syscall"
)

func run(c *cli.Context) {
	gateHandler, err := weir2.CreateGateHandler()
	if err != nil {
		log.WithError(err).Error("Connecting to hyena daemon")
		return
	}
	server := weir2.NewServer(c.Int("http"), c.Int("https"), c.String("certsDir"), gateHandler)
	server.Serve()
	log.Info("Gateway is listening")
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)
	signal.Notify(closeChan, syscall.SIGTERM)
	<-closeChan
	log.Info("Gateway shut down")
}

func main() {
	app := cli.NewApp()
	app.Name = "gate"
	app.Usage = "Hyena Http Gateway"
	app.Action = run
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "certsDir, c",
			Value: "",
			Usage: "Folder containing SSL certificates",
		},
		cli.IntFlag{
			Name:  "http, p",
			Value: 80,
			Usage: "Http listen port (0 to disable)",
		},
		cli.IntFlag{
			Name:  "https, s",
			Value: 443,
			Usage: "Https listen port (0 to disable)",
		},
	}
	app.Run(os.Args)
}
