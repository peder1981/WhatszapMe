package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	qrcode "github.com/mdp/qrterminal/v3"
)

func main() {
	// Define flags para teste
	content := flag.String("content", "https://github.com/peder/whatszapme", "Conteúdo para codificar no QR Code")
	forceMode := flag.String("mode", "auto", "Força o modo de renderização (auto, standard, halfblock)")
	flag.Parse()

	fmt.Printf("Sistema operacional: %s\n", runtime.GOOS)
	fmt.Printf("Arquitetura: %s\n", runtime.GOARCH)
	fmt.Printf("Conteúdo do QR Code: %s\n\n", *content)

	// Define o tipo de QR Code baseado no sistema operacional ou flag
	mode := *forceMode
	if mode == "auto" {
		if runtime.GOOS == "windows" {
			mode = "standard"
		} else {
			mode = "halfblock"
		}
	}

	fmt.Printf("Modo de renderização: %s\n\n", mode)
	fmt.Println("QR Code de teste:")

	// Opção de qualidade do QR Code (Low)

	// Gera o QR Code no modo selecionado
	switch mode {
	case "standard":
		qrcode.Generate(*content, qrcode.L, os.Stdout)
	case "halfblock":
		qrcode.GenerateHalfBlock(*content, qrcode.L, os.Stdout)
	default:
		fmt.Println("Modo inválido. Use 'standard' ou 'halfblock'.")
	}

	fmt.Println("\nTeste de QR Code concluído!")
	fmt.Println("\nInstruções para testar em diferentes plataformas:")
	fmt.Println("1. Compile o programa para cada plataforma:")
	fmt.Println("   - Windows: GOOS=windows GOARCH=amd64 go build -o test_qrcode_windows.exe ./cmd/test_qrcode/")
	fmt.Println("   - macOS: GOOS=darwin GOARCH=amd64 go build -o test_qrcode_macos ./cmd/test_qrcode/")
	fmt.Println("   - Linux: GOOS=linux GOARCH=amd64 go build -o test_qrcode_linux ./cmd/test_qrcode/")
	fmt.Println("2. Execute o binário na respectiva plataforma")
	fmt.Println("3. Escaneie o QR Code com qualquer leitor (câmera do celular, aplicativo QR Code)")
	fmt.Println("4. Verifique se o conteúdo é lido corretamente")
}
