package ui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/skip2/go-qrcode"
)

// QRCodeGenerator gera e atualiza o QR Code na interface gráfica
type QRCodeGenerator struct {
	container   *fyne.Container
	imageCanvas *canvas.Image
	statusText  *widget.Label
	progress    *widget.ProgressBar // Barra de progresso para sincronização
}

// NewQRCodeGenerator cria um novo gerador de QR Code para a UI
func NewQRCodeGenerator() *QRCodeGenerator {
	statusText := widget.NewLabel("Aguardando código QR")
	imageCanvas := canvas.NewImageFromImage(createEmptyQRImage(256))
	imageCanvas.FillMode = canvas.ImageFillOriginal
	
	// Cria a barra de progresso inicialmente oculta (valor=0)
	progressBar := widget.NewProgressBar()
	progressBar.Hide() // Inicialmente oculta
	
	content := container.NewVBox(
		container.NewCenter(imageCanvas),
		container.NewCenter(statusText),
		container.NewPadded(progressBar),
	)
	
	return &QRCodeGenerator{
		container:   content,
		imageCanvas: imageCanvas,
		statusText:  statusText,
		progress:    progressBar,
	}
}

// Container retorna o container Fyne para o QR Code
func (qr *QRCodeGenerator) Container() *fyne.Container {
	return qr.container
}

// UpdateQRCode atualiza a imagem do QR Code na interface
func (qr *QRCodeGenerator) UpdateQRCode(data string) error {
	// Log debug para verificar se a função está sendo chamada
	fmt.Printf("QRCodeGenerator.UpdateQRCode chamado com dados de %d bytes\n", len(data))
	
	qrImg, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		fmt.Printf("Erro ao gerar QR Code: %v\n", err)
		qr.statusText.SetText("Erro ao gerar QR Code: " + err.Error())
		return err
	}
	
	// Gera a imagem do QR Code
	img := qrImg.Image(256)
	
	// Atualiza a imagem e o texto
	qr.statusText.SetText("Escaneie este QR Code com o WhatsApp no seu celular")
	qr.imageCanvas.Image = img
	
	// Força a atualização visual do componente
	qr.imageCanvas.Refresh()
	
	// Log de sucesso
	fmt.Println("QR Code atualizado na interface!")
	
	return nil
}

// ClearQRCode limpa a imagem do QR Code
func (qr *QRCodeGenerator) ClearQRCode(message string) {
	qr.imageCanvas.Image = createEmptyQRImage(256)
	qr.imageCanvas.Refresh()
	qr.statusText.SetText(message)
	// Oculta a barra de progresso quando limpa o QR code
	qr.progress.Hide()
}

// StartProgress inicia a exibição da barra de progresso
func (qr *QRCodeGenerator) StartProgress(message string) {
	// Define a mensagem de status
	qr.statusText.SetText(message)
	
	// Exibe a barra de progresso indefinida (mostrando atividade)
	qr.progress.Show()
	qr.progress.SetValue(0.0) // Indeterminado
}

// UpdateProgress atualiza o valor da barra de progresso
func (qr *QRCodeGenerator) UpdateProgress(value float64, message string) {
	// Se valor for maior que 0, passa para modo determinado
	if value > 0 {
		qr.progress.SetValue(value)
	}
	
	// Atualiza a mensagem se fornecida
	if message != "" {
		qr.statusText.SetText(message)
	}
}

// StopProgress esconde a barra de progresso
func (qr *QRCodeGenerator) StopProgress(message string) {
	qr.progress.Hide()
	
	// Atualiza a mensagem se fornecida
	if message != "" {
		qr.statusText.SetText(message)
	}
}

// createEmptyQRImage cria uma imagem vazia para o QR Code
func createEmptyQRImage(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Preenche com fundo branco e borda cinza
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Borda de 1 pixel
			if x == 0 || x == size-1 || y == 0 || y == size-1 {
				img.Set(x, y, color.RGBA{200, 200, 200, 255}) // Cinza claro
			} else {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // Branco
			}
		}
	}
	
	return img
}
