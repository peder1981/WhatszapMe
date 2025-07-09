package ui

import (
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
}

// NewQRCodeGenerator cria um novo gerador de QR Code para a UI
func NewQRCodeGenerator() *QRCodeGenerator {
	statusText := widget.NewLabel("Aguardando código QR")
	imageCanvas := canvas.NewImageFromImage(createEmptyQRImage(256))
	imageCanvas.FillMode = canvas.ImageFillOriginal
	
	content := container.NewVBox(
		container.NewCenter(imageCanvas),
		container.NewCenter(statusText),
	)
	
	return &QRCodeGenerator{
		container:   content,
		imageCanvas: imageCanvas,
		statusText:  statusText,
	}
}

// Container retorna o container Fyne para o QR Code
func (qr *QRCodeGenerator) Container() *fyne.Container {
	return qr.container
}

// UpdateQRCode atualiza a imagem do QR Code na interface
func (qr *QRCodeGenerator) UpdateQRCode(data string) error {
	qrImg, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		qr.statusText.SetText("Erro ao gerar QR Code: " + err.Error())
		return err
	}
	
	qr.statusText.SetText("Escaneie este QR Code com o WhatsApp no seu celular")
	img := qrImg.Image(256)
	qr.imageCanvas.Image = img
	qr.imageCanvas.Refresh()
	
	return nil
}

// ClearQRCode limpa a imagem do QR Code
func (qr *QRCodeGenerator) ClearQRCode(message string) {
	qr.imageCanvas.Image = createEmptyQRImage(256)
	qr.imageCanvas.Refresh()
	qr.statusText.SetText(message)
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
