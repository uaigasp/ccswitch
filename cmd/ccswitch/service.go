package main

import (
	"fmt"
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
	stop     chan struct{}
	dir      string
	credPath string
	logger   service.Logger
}

func (p *program) Start(s service.Service) error {
	p.stop = make(chan struct{})
	p.logger, _ = s.Logger(nil)
	go p.loop()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(p.stop)
	return nil
}

func (p *program) loop() {
	dir := p.dir
	if dir == "" {
		var err error
		dir, err = appDataDir()
		if err != nil {
			log.Println("error obteniendo appdata dir:", err)
			return
		}
	}

	credPath := p.credPath
	if credPath == "" {
		var err error
		credPath, err = swap.DefaultCredentialsPath()
		if err != nil {
			log.Println("error obteniendo credentials path:", err)
			return
		}
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Println("error creando directorio de datos:", err)
	}

	logFile, err := os.OpenFile(filepath.Join(dir, "ccswitch.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		if p.logger != nil {
			p.logger.Error("no se pudo abrir el archivo de log: ", err)
		}
		return
	}
	log.SetOutput(logFile)
	defer logFile.Close()

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

func serviceConfig(extraArgs ...string) *service.Config {
	args := []string{"service", "run"}
	args = append(args, extraArgs...)
	return &service.Config{
		Name:        "ccswitch",
		DisplayName: "ccswitch auto-switch",
		Description: "cambia entre cuentas de claude code cuando el uso se acerca al limite",
		Arguments:   args,
	}
}

func runServiceInstall() error {
	dir, err := appDataDir()
	if err != nil {
		return fmt.Errorf("no se pudo resolver el directorio de datos: %w", err)
	}

	credPath, err := swap.DefaultCredentialsPath()
	if err != nil {
		return fmt.Errorf("no se pudo resolver la ruta de credenciales: %w", err)
	}

	s, err := service.New(&program{}, serviceConfig("--dir", dir, "--cred", credPath))
	if err != nil {
		return err
	}
	if err := s.Install(); err != nil {
		return err
	}

	if err := s.Start(); err != nil {
		fmt.Println("el servicio se instalo pero no arranco solo:", err)
		fmt.Println("iniciarlo a mano desde services.msc o correr: ccswitch service status")
		return nil
	}

	return nil
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

func runServiceRun(extra []string) error {
	dir, credPath := parseServiceRunFlags(extra)

	s, err := service.New(&program{dir: dir, credPath: credPath}, serviceConfig())
	if err != nil {
		return err
	}
	return s.Run()
}

func parseServiceRunFlags(args []string) (dir, credPath string) {
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--dir":
			dir = args[i+1]
		case "--cred":
			credPath = args[i+1]
		}
	}
	return dir, credPath
}
