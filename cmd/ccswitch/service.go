package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"

	"github.com/uaigasp/ccswitch/internal/config"
	"github.com/uaigasp/ccswitch/internal/serviceloop"
	"github.com/uaigasp/ccswitch/internal/swap"
)

type program struct {
	stop chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.stop = make(chan struct{})
	go p.loop()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(p.stop)
	return nil
}

func (p *program) loop() {
	dir, err := appDataDir()
	if err != nil {
		log.Println("error obteniendo appdata dir:", err)
		return
	}

	credPath, err := swap.DefaultCredentialsPath()
	if err != nil {
		log.Println("error obteniendo credentials path:", err)
		return
	}

	logFile, err := os.OpenFile(filepath.Join(dir, "ccswitch.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	var lastSwitch time.Time

	for {
		cfg, err := config.Load(dir)
		if err != nil {
			log.Println("error leyendo config:", err)
		} else {
			decision, err := serviceloop.RunOnce(dir, credPath, cfg, lastSwitch, false)
			if err != nil {
				log.Println("error en el ciclo de auto-switch:", err)
			} else if decision.ShouldSwitch {
				lastSwitch = time.Now()
				log.Println("cambio de cuenta a:", decision.TargetAlias)
			}
		}

		select {
		case <-p.stop:
			return
		case <-time.After(time.Duration(mustInterval(dir)) * time.Second):
		}
	}
}

func mustInterval(dir string) int {
	cfg, err := config.Load(dir)
	if err != nil {
		return config.Default().IntervalSeconds
	}
	return cfg.IntervalSeconds
}

func serviceConfig() *service.Config {
	return &service.Config{
		Name:        "ccswitch",
		DisplayName: "ccswitch auto-switch",
		Description: "cambia entre cuentas de claude code cuando el uso se acerca al limite",
	}
}

func runServiceInstall() error {
	s, err := service.New(&program{}, serviceConfig())
	if err != nil {
		return err
	}
	return s.Install()
}

func runServiceUninstall() error {
	s, err := service.New(&program{}, serviceConfig())
	if err != nil {
		return err
	}
	return s.Uninstall()
}

func runServiceStatus() error {
	s, err := service.New(&program{}, serviceConfig())
	if err != nil {
		return err
	}
	status, err := s.Status()
	if err != nil {
		return err
	}
	switch status {
	case service.StatusRunning:
		log.Println("corriendo")
	case service.StatusStopped:
		log.Println("detenido")
	default:
		log.Println("desconocido")
	}
	return nil
}

func runServiceRun() error {
	s, err := service.New(&program{}, serviceConfig())
	if err != nil {
		return err
	}
	return s.Run()
}
