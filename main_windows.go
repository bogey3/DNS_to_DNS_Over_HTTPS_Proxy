package main

import (
	"github.com/kardianos/service"
	"log"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	DNSServer()
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func runAsService(){
	svcConfig := &service.Config{
		Name:        "Local DNS over HTTPS Server",
		DisplayName: "DNS-HTTPS Server",
		Description: "This server will listen for dns requests, and forward them to a public dns over https server.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}


func main() {
	runAsService()
}
