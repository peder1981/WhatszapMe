package whatsapp

import (
	"os"
	"path/filepath"
)

// createDirIfNotExists cria um diretório se ele não existir
func createDirIfNotExists(dir string) error {
	if dir == "" {
		return nil
	}

	// Verifica se o diretório já existe
	info, err := os.Stat(dir)
	if err == nil {
		if info.IsDir() {
			return nil // Diretório já existe
		}
		return os.ErrExist // Existe, mas não é um diretório
	}

	// Cria o diretório e seus pais se necessário
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}

	return err
}

// ensureDBDirectory garante que o diretório do banco de dados existe
func ensureDBDirectory(dbPath string) error {
	// Extrai o diretório do caminho do banco de dados
	dbDir := filepath.Dir(dbPath)
	return createDirIfNotExists(dbDir)
}

// formatJID formata um JID para exibição amigável
func formatJID(jid string) string {
	return jid
}

// extractPhoneNumber extrai o número de telefone de um JID
func extractPhoneNumber(jid string) string {
	return jid
}
