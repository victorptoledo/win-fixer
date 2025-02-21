package main

import (
	"log"
	"os/exec"
)

// Verifica se o programa está sendo executado como administrador.
func isAdmin() bool {
	_, err := exec.Command("cmd", "/C", "net session").Output()
	if err != nil {
		log.Println("[WARN] O programa não está sendo executado como administrador.")
		return false
	}
	log.Println("[INFO] O programa está sendo executado como administrador.")
	return true
}

func main() {
	if !isAdmin() {
		log.Println("[ERROR] Este programa precisa ser executado como administrador.")
		return
	}
	log.Println("[INFO] Programa iniciado com sucesso.")
}
