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
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	_ = args
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
