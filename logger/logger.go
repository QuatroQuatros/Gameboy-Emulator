package logger

import (
	"fmt"
	"strings"
	"time"
)

var logChan = make(chan string, 1000) // Buffer de 10000 mensagens

// Inicia a goroutine de impressão de logs
func init() {
	go logPrinter()
}

// Goroutine que imprime os logs
func logPrinter() {
	var logBuffer strings.Builder

	// Intervalo de impressão (1 segundo)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message := <-logChan:
			// Adiciona a mensagem ao buffer
			logBuffer.WriteString(message + "\n")

			// Se o buffer estiver muito grande, imprime e limpa
			if logBuffer.Len() > 4096 { // 4 KB
				fmt.Print(logBuffer.String())
				logBuffer.Reset()
			}

		case <-ticker.C:
			// Imprime o buffer a cada segundo
			if logBuffer.Len() > 0 {
				fmt.Print(logBuffer.String())
				logBuffer.Reset()
			}
		}
	}
}

// Função para enviar logs ao canal
func LogMessage(message string) {
	select {
	case logChan <- message: // Envia a mensagem para o canal
	default:
		// Se o canal estiver cheio, descarta a mensagem para evitar bloqueio
	}
}

// Função para obter os logs restantes no canal
func GetRemainingLogs() string {
	var logBuffer strings.Builder

	// Esvazia o canal e adiciona as mensagens ao buffer
	for {
		select {
		case message := <-logChan:
			logBuffer.WriteString(message + "\n")
		default:
			return logBuffer.String()
		}
	}
}

// var (
// 	LogBuffer strings.Builder
// 	mutex     sync.Mutex // Para garantir acesso seguro ao buffer em concorrência
// )

// // Função para adicionar logs ao buffer
// func LogMessage(message string) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	LogBuffer.WriteString(message + "\n")
// }

// // Função para exibir o buffer no final
// func DisplayLogs() {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	fmt.Println("\n=== Logs do programa ===")
// 	fmt.Print(LogBuffer.String())
// 	fmt.Println("========================")
// }
