package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
	case "add":
		alias, email := parseAddFlags(args)
		if err := runAdd(alias, email); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("cuenta agregada:", alias)
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "status":
		if err := runStatus(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "switch":
		var alias string
		if len(args) > 0 {
			alias = args[0]
		}
		if err := runSwitch(alias); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("cambiado a:", alias)
	case "service":
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "uso: ccswitch service <install|uninstall|status|run>")
			os.Exit(1)
		}
		var err error
		switch args[0] {
		case "install":
			err = runServiceInstall()
		case "uninstall":
			err = runServiceUninstall()
		case "status":
			err = runServiceStatus()
		case "run":
			err = runServiceRun()
		default:
			fmt.Fprintln(os.Stderr, "subcomando desconocido:", args[0])
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ccswitch - maneja varias cuentas de Claude Code

Comandos:
  add                agrega la cuenta logueada ahora mismo
  list                lista las cuentas guardadas
  status               muestra la cuenta activa y su uso
  switch <alias>       cambia a otra cuenta guardada
  service install       instala el servicio de auto-switch
  service uninstall     lo desinstala
  service status         dice si esta corriendo`)
}
