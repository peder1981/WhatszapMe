package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB representa uma conexão com o banco de dados SQLite
type DB struct {
	conn *sql.DB
}

// Contato representa um contato do WhatsApp
type Contato struct {
	ID         int64
	JID        string    // ID do contato no WhatsApp
	Nome       string    // Nome do contato
	Telefone   string    // Número de telefone formatado
	UltimaSync time.Time // Último momento de sincronização
}

// Mensagem representa uma mensagem no histórico
type Mensagem struct {
	ID        int64
	JID       string    // ID do contato no WhatsApp
	Nome      string    // Nome do contato ou remetente
	Conteudo  string    // Conteúdo da mensagem
	Resposta  string    // Resposta gerada pelo LLM
	Timestamp time.Time // Momento do recebimento/envio
	Entrada   bool      // True = recebida do contato, False = enviada pelo sistema
}

// Opções para consulta de mensagens
type OpcoesConsulta struct {
	JID       string    // Filtrar por contato específico
	DataInicio time.Time // Filtrar a partir desta data
	DataFim    time.Time // Filtrar até esta data
	Limite     int       // Limitar quantidade de resultados
	Ordem      string    // "asc" (mais antigas primeiro) ou "desc" (mais recentes primeiro)
}

// New cria uma nova instância de DB
func New(caminhoDB string) (*DB, error) {
	// Garante que o diretório existe
	dirPath := filepath.Dir(caminhoDB)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório para banco de dados: %w", err)
	}

	// Abre a conexão com o banco
	conn, err := sql.Open("sqlite3", caminhoDB)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco de dados: %w", err)
	}

	// Testa a conexão
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	db := &DB{conn: conn}

	// Inicializa as tabelas
	if err := db.inicializarTabelas(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao inicializar tabelas: %w", err)
	}

	return db, nil
}

// Fecha a conexão com o banco de dados
func (db *DB) Close() error {
	return db.conn.Close()
}

// Inicializa as tabelas do banco de dados
func (db *DB) inicializarTabelas() error {
	// Tabela para armazenar o histórico de mensagens
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS mensagens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			jid TEXT NOT NULL,
			nome TEXT NOT NULL,
			conteudo TEXT NOT NULL,
			resposta TEXT,
			timestamp DATETIME NOT NULL,
			entrada BOOLEAN NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_mensagens_jid ON mensagens(jid);
		CREATE INDEX IF NOT EXISTS idx_mensagens_timestamp ON mensagens(timestamp);
		
		-- Nova tabela para armazenar contatos
		CREATE TABLE IF NOT EXISTS contatos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			jid TEXT NOT NULL UNIQUE,
			nome TEXT NOT NULL,
			telefone TEXT,
			ultima_sync DATETIME NOT NULL,
			auto_responder BOOLEAN DEFAULT true
		);
		
		CREATE INDEX IF NOT EXISTS idx_contatos_jid ON contatos(jid);
		CREATE INDEX IF NOT EXISTS idx_contatos_nome ON contatos(nome);
	`)

	if err != nil {
		return fmt.Errorf("erro ao criar tabelas: %w", err)
	}

	return nil
}

// SalvarMensagem salva uma nova mensagem no histórico
func (db *DB) SalvarMensagem(msg Mensagem) (int64, error) {
	query := `
		INSERT INTO mensagens (jid, nome, conteudo, resposta, timestamp, entrada)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		msg.JID,
		msg.Nome,
		msg.Conteudo,
		msg.Resposta,
		msg.Timestamp,
		msg.Entrada,
	)

	if err != nil {
		return 0, fmt.Errorf("erro ao salvar mensagem: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("erro ao obter ID da mensagem inserida: %w", err)
	}

	return id, nil
}

// AtualizarResposta atualiza a resposta associada a uma mensagem
func (db *DB) AtualizarResposta(id int64, resposta string) error {
	_, err := db.conn.Exec("UPDATE mensagens SET resposta = ? WHERE id = ?", resposta, id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar resposta da mensagem %d: %w", id, err)
	}
	return nil
}

// BuscarMensagens busca mensagens no histórico com base nas opções fornecidas
func (db *DB) BuscarMensagens(opcoes OpcoesConsulta) ([]Mensagem, error) {
	query := "SELECT id, jid, nome, conteudo, resposta, timestamp, entrada FROM mensagens WHERE 1=1"
	args := []interface{}{}

	// Adiciona filtros à consulta
	if opcoes.JID != "" {
		query += " AND jid = ?"
		args = append(args, opcoes.JID)
	}

	// Filtro por data inicial
	if !opcoes.DataInicio.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, opcoes.DataInicio)
	}

	// Filtro por data final
	if !opcoes.DataFim.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, opcoes.DataFim)
	}

	// Ordenação
	if opcoes.Ordem == "asc" {
		query += " ORDER BY timestamp ASC"
	} else {
		query += " ORDER BY timestamp DESC" // padrão é mais recente primeiro
	}

	// Limite
	if opcoes.Limite > 0 {
		query += " LIMIT ?"
		args = append(args, opcoes.Limite)
	}

	// Executa a consulta
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagens: %w", err)
	}
	defer rows.Close()

	// Processa os resultados
	var mensagens []Mensagem
	for rows.Next() {
		var msg Mensagem
		var timestamp string // SQLite retorna timestamp como string

		if err := rows.Scan(&msg.ID, &msg.JID, &msg.Nome, &msg.Conteudo, &msg.Resposta, &timestamp, &msg.Entrada); err != nil {
			return nil, fmt.Errorf("erro ao ler mensagem: %w", err)
		}

		// Converte o timestamp para time.Time
		// Tenta primeiro no formato RFC3339 (formato padrão do Go)
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			// Se falhar, tenta no formato SQLite padrão
			t, err = time.Parse("2006-01-02 15:04:05", timestamp)
			if err != nil {
				return nil, fmt.Errorf("erro ao processar timestamp: %w", err)
			}
		}
		msg.Timestamp = t

		mensagens = append(mensagens, msg)
	}

	return mensagens, nil
}

// BuscarContatos retorna uma lista de contatos do banco de dados ordenados por atividade mais recente
func (db *DB) BuscarContatos() ([]struct{ JID, Nome string }, error) {
	// Nova implementação: busca contatos ordenados pela última mensagem (mais recente primeiro)
	// Isso simula o comportamento do WhatsApp Web, onde contatos mais ativos aparecem no topo
	query := `
		WITH UltimasMensagens AS (
			SELECT jid, MAX(timestamp) as ultima_atividade
			FROM mensagens
			GROUP BY jid
		)
		SELECT c.jid, c.nome, um.ultima_atividade
		FROM contatos c
		LEFT JOIN UltimasMensagens um ON c.jid = um.jid
		UNION
		SELECT m.jid, m.nome, um.ultima_atividade
		FROM mensagens m
		JOIN UltimasMensagens um ON m.jid = um.jid
		WHERE m.jid NOT IN (SELECT jid FROM contatos)
		GROUP BY m.jid
		ORDER BY ultima_atividade DESC NULLS LAST
	`

	// Tenta executar a query com NULLS LAST (SQLite >= 3.30.0)
	rows, err := db.conn.Query(query)
	
	// Se der erro (versões mais antigas do SQLite não suportam NULLS LAST), use uma query alternativa
	if err != nil {
		query = `
			WITH UltimasMensagens AS (
				SELECT jid, MAX(timestamp) as ultima_atividade
				FROM mensagens
				GROUP BY jid
			)
			SELECT c.jid, c.nome, um.ultima_atividade
			FROM contatos c
			LEFT JOIN UltimasMensagens um ON c.jid = um.jid
			UNION
			SELECT m.jid, m.nome, um.ultima_atividade
			FROM mensagens m
			JOIN UltimasMensagens um ON m.jid = um.jid
			WHERE m.jid NOT IN (SELECT jid FROM contatos)
			GROUP BY m.jid
			ORDER BY CASE WHEN ultima_atividade IS NULL THEN 0 ELSE 1 END DESC, ultima_atividade DESC
		`
		
		rows, err = db.conn.Query(query)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar contatos: %w", err)
		}
	}
	defer rows.Close()

	var contatos []struct{ JID, Nome string }
	for rows.Next() {
		var contato struct{ JID, Nome string }
		var ultimaAtividade interface{} // Usamos interface{} para aceitar qualquer tipo retornado pelo SQLite
		if err := rows.Scan(&contato.JID, &contato.Nome, &ultimaAtividade); err != nil {
			return nil, fmt.Errorf("erro ao ler contato: %w", err)
		}
		contatos = append(contatos, contato)
	}

	// Se não há contatos, cria um contato de teste para demonstração
	if len(contatos) == 0 {
		// Cria contato de teste
		testeJID := "123456789@s.whatsapp.net"
		testeNome := "Contato Teste"
		
		_, err = db.conn.Exec(
			"INSERT OR IGNORE INTO contatos (jid, nome, telefone, ultima_sync, auto_responder) VALUES (?, ?, ?, ?, ?)",
			testeJID, testeNome, "+123456789", time.Now(), true,
		)
		
		if err != nil {
			// Apenas log, não retorna erro
			fmt.Printf("Erro ao criar contato de teste: %v\n", err)
		} else {
			// Adiciona o contato de teste à lista
			contatos = append(contatos, struct{ JID, Nome string }{testeJID, testeNome})
		}
	}

	return contatos, nil
}

// SincronizarContato adiciona ou atualiza um contato no banco de dados
func (db *DB) SincronizarContato(jid, nome, telefone string) error {
	query := `
		INSERT INTO contatos (jid, nome, telefone, ultima_sync) 
		VALUES (?, ?, ?, ?)
		ON CONFLICT(jid) DO UPDATE SET 
			nome = ?,
			telefone = ?,
			ultima_sync = ?
	`
	agora := time.Now()
	
	_, err := db.conn.Exec(query, jid, nome, telefone, agora, nome, telefone, agora)
	if err != nil {
		return fmt.Errorf("erro ao sincronizar contato: %w", err)
	}
	
	return nil
}

// ListarTodosContatos retorna todos os contatos cadastrados na tabela de contatos
func (db *DB) ListarTodosContatos() ([]Contato, error) {
	query := `SELECT id, jid, nome, telefone, ultima_sync FROM contatos ORDER BY nome`
	
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contatos: %w", err)
	}
	defer rows.Close()
	
	var contatos []Contato
	for rows.Next() {
		var contato Contato
		if err := rows.Scan(&contato.ID, &contato.JID, &contato.Nome, &contato.Telefone, &contato.UltimaSync); err != nil {
			return nil, fmt.Errorf("erro ao ler contato: %w", err)
		}
		contatos = append(contatos, contato)
	}
	
	return contatos, nil
}

// ObterUltimasMensagens retorna as últimas N mensagens de um contato específico
func (db *DB) ObterUltimasMensagens(jid string, quantidade int) ([]Mensagem, error) {
	return db.BuscarMensagens(OpcoesConsulta{
		JID:    jid,
		Limite: quantidade,
		Ordem:  "desc",
	})
}

// ExcluirHistoricoContato exclui todo o histórico de um contato específico
func (db *DB) ExcluirHistoricoContato(jid string) error {
	_, err := db.conn.Exec("DELETE FROM mensagens WHERE jid = ?", jid)
	if err != nil {
		return fmt.Errorf("erro ao excluir histórico do contato %s: %w", jid, err)
	}
	return nil
}

// LimparHistorico exclui todos os registros de mensagens
func (db *DB) LimparHistorico() error {
	_, err := db.conn.Exec("DELETE FROM mensagens")
	if err != nil {
		return fmt.Errorf("erro ao limpar histórico: %w", err)
	}
	return nil
}
