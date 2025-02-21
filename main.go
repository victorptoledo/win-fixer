package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var logger *log.Logger

func setupLogger() {
	logFile, err := os.OpenFile("win-fixer.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.Fatalf("[ERROR] Falha ao criar o arquivo de log: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func isAdmin() bool {
	_, err := exec.Command("cmd", "/C", "net session").Output()
	if err != nil {
		log.Println("[WARN] O programa não está sendo executado como administrador.")
		return false
	}
	logger.Println("[INFO] O programa está sendo executado como administrador.")
	return true
}

func runCommand(name string, args ...string) error {
	logger.Printf("[INFO] Executando comando: %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	if err != nil {
		logger.Printf("[ERROR] Falha ao executar comando: %s %v\n", name, args)
		return err
	}
	logger.Printf("[INFO] Comando executado com sucesso: %s %v\n", name, args)
	return nil
}

func main() {
	setupLogger()

	if !isAdmin() {
		logger.Println("[ERROR] Este programa precisa ser executado como administrador.")
		return
	}

	myApp := app.NewWithID("win-fixer")
	myWindow := myApp.NewWindow("Win Fixer")

	logText := widget.NewMultiLineEntry()
	logText.SetText("Iniciando a execução dos comandos...\n")

	startButton := widget.NewButton("Iniciar manutenção", func() {
		logger.Println("[INFO] Iniciando a execução dos comandos via interface gráfica...")
		logText.Append("Executando comandos...\n")

		commands := []struct {
			name string
			args []string
		}{
			{"sfc", []string{"/scannow"}},
			{"DISM", []string{"/Online", "/Cleanup-Image", "/RestoreHealth"}},
		}

		for _, command := range commands {
			logText.Append(fmt.Sprintf("[INFO] Executando comando: %s %v\n", command.name, command.args))
			err := runCommand(command.name, command.args...)
			if err != nil {
				logText.Append(fmt.Sprintf("[ERROR] Falha ao executar comando: %s %v\n", command.name, command.args))
				logger.Printf("[ERROR] Falha ao executar comando: %s %v\n", command.name, command.args)
			} else {
				logText.Append(fmt.Sprintf("[INFO] Comando executado com sucesso: %s %v\n", command.name, command.args))
			}
			time.Sleep(1 * time.Second)
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Win Fixer"),
		logText,
		startButton,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 100))
	myWindow.ShowAndRun()
}
