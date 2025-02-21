package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	cancelExecution bool
	mu              sync.Mutex
	logger          *log.Logger
)

func setupLogger() {
	logFile, err := os.OpenFile("program.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("[ERROR] Falha ao configurar o logger: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("[INFO] Logger inicializado com sucesso.")
}

func isAdmin() bool {
	_, err := exec.Command("cmd", "/C", "net session").Output()
	if err != nil {
		logger.Printf("[WARN] O programa não está sendo executado como administrador: %v", err)
		return false
	}
	logger.Println("[INFO] O programa está sendo executado como administrador.")
	return true
}

func runCommand(name string, args ...string) error {
	logger.Printf("[INFO] Executando comando: %s %v", name, args)
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // Esconde a janela do terminal
	err := cmd.Run()
	if err != nil {
		logger.Printf("[ERROR] Falha ao executar comando %s: %v", name, err)
		return err
	}
	logger.Printf("[INFO] Comando concluído com sucesso: %s", name)
	return nil
}

func main() {
	setupLogger()

	if !isAdmin() {
		logger.Println("[ERROR] Este programa precisa ser executado como administrador.")
		dialog.ShowError(fmt.Errorf("este programa precisa ser executado como administrador"), nil)
		return
	}

	// Inicializa a aplicação usando o ID personalizado
	myApp := app.NewWithID("win-fixer")
	myWindow := myApp.NewWindow("Monitoria do Sistema")

	// Elementos da interface gráfica
	logText := widget.NewMultiLineEntry()
	logText.Disabled()
	logText.SetText("Iniciando a execução dos comandos...\n")

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	statusLabel := widget.NewLabel("Status: Aguardando início...")

	cancelButton := widget.NewButton("Cancelar Execução", func() {
		mu.Lock()
		cancelExecution = true
		mu.Unlock()
		statusLabel.SetText("Status: Execução cancelada pelo usuário.")
		progressBar.Hide()
	})
	cancelButton.Disable()

	var startButton *widget.Button
	startButton = widget.NewButton("Iniciar Manutenção", func() {
		startButton.Disable()
		cancelButton.Enable()
		progressBar.Show()
		statusLabel.SetText("Status: Executando comandos...")
		cancelExecution = false

		go func() {
			defer func() {
				startButton.Enable()
				cancelButton.Disable()
				progressBar.Hide()
			}()

			commands := []struct {
				name string
				args []string
			}{
				{"sfc", []string{"/scannow"}},
				{"DISM", []string{"/Online", "/Cleanup-Image", "/RestoreHealth"}},
				{"winget", []string{"upgrade", "--all"}},
				{"cmd", []string{"/C", "del /S /Q %TEMP%\\*"}},
				{"ipconfig", []string{"/flushdns"}},
				{"ipconfig", []string{"/release"}},
				{"ipconfig", []string{"/renew"}},
				{"defrag", []string{"C:", "/U", "/V"}},
				{"chkdsk", []string{"C:", "/F", "/R"}},
				{"wevtutil", []string{"cl", "Application"}},
			}

			for _, command := range commands {
				mu.Lock()
				if cancelExecution {
					mu.Unlock()
					logger.Println("[INFO] Execução cancelada pelo usuário.")
					logText.SetText(logText.Text + "Execução cancelada pelo usuário.\n")
					return
				}
				mu.Unlock()

				logger.Printf("[INFO] Executando comando: %s %v", command.name, command.args)
				logText.SetText(logText.Text + fmt.Sprintf("Executando: %s %v\n", command.name, command.args))
				err := runCommand(command.name, command.args...)
				if err != nil {
					logText.SetText(logText.Text + fmt.Sprintf("Erro ao executar %s: %v\n", command.name, err))
				} else {
					logText.SetText(logText.Text + fmt.Sprintf("%s concluído com sucesso.\n", command.name))
				}

				time.Sleep(1 * time.Second)
			}

			dialog.ShowConfirm("Reinicialização Necessária", "Todos os comandos foram concluídos. Deseja reiniciar o sistema agora?", func(restart bool) {
				if restart {
					logger.Println("[INFO] Reiniciando o sistema...")
					err := runCommand("shutdown", "/r", "/t", "0")
					if err != nil {
						return
					}
				} else {
					logger.Println("[INFO] Reinicialização cancelada pelo usuário.")
					logText.SetText(logText.Text + "Reinicialização cancelada pelo usuário.\n")
					statusLabel.SetText("Status: Concluído.")
				}
			}, myWindow)
		}()
	})

	// Componentes da interface
	content := container.NewVBox(
		widget.NewLabelWithStyle("Monitoria do Sistema", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		logText,
		progressBar,
		statusLabel,
		container.NewHBox(startButton, cancelButton),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.SetIcon(theme.ComputerIcon()) // Ícone do sistema
	myWindow.ShowAndRun()
}
