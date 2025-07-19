package ui

import (
	"fyne.io/fyne/v2"
)

// Trunca o texto para exibição em logs limitando o tamanho
func truncarTexto(texto string, tamanho int) string {
	if len(texto) <= tamanho {
		return texto
	}
	return texto[:tamanho] + "..."
}

// Layout personalizado para texto com quebra de linha
type textoLayout struct{}

func (tl *textoLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width float32
	var height float32
	
	for _, obj := range objects {
		size := obj.MinSize()
		if size.Width > width {
			width = size.Width
		}
		height += size.Height
	}
	
	return fyne.NewSize(width, height)
}

func (tl *textoLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	
	for _, obj := range objects {
		size := obj.MinSize()
		obj.Resize(fyne.NewSize(containerSize.Width, size.Height))
		obj.Move(pos)
		pos = pos.Add(fyne.NewPos(0, size.Height))
	}
}
