package db

import (
	"os"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	// Cria um banco de dados temporário para testes
	tempFile := "test_db.sqlite"
	defer os.Remove(tempFile) // Remove o arquivo após o teste
	
	// Inicializa o banco de dados
	db, err := New(tempFile)
	if err != nil {
		t.Fatalf("Erro ao criar banco de dados de teste: %v", err)
	}
	defer db.Close()
	
	// Testa salvar mensagem
	t.Run("SalvarMensagem", func(t *testing.T) {
		msg := Mensagem{
			JID:       "5511999999999@s.whatsapp.net",
			Nome:      "Teste",
			Conteudo:  "Olá, como vai?",
			Timestamp: time.Now(),
			Entrada:   true,
		}
		
		id, err := db.SalvarMensagem(msg)
		if err != nil {
			t.Errorf("Erro ao salvar mensagem: %v", err)
		}
		
		if id <= 0 {
			t.Errorf("ID inválido retornado: %d", id)
		}
	})
	
	// Testa atualizar resposta
	t.Run("AtualizarResposta", func(t *testing.T) {
		// Primeiro salva uma mensagem
		msg := Mensagem{
			JID:       "5511999999999@s.whatsapp.net",
			Nome:      "Teste",
			Conteudo:  "Olá, como vai?",
			Timestamp: time.Now(),
			Entrada:   true,
		}
		
		id, err := db.SalvarMensagem(msg)
		if err != nil {
			t.Fatalf("Erro ao salvar mensagem para teste de atualização: %v", err)
		}
		
		// Agora atualiza a resposta
		err = db.AtualizarResposta(id, "Estou bem, obrigado!")
		if err != nil {
			t.Errorf("Erro ao atualizar resposta: %v", err)
		}
		
		// Verifica se a resposta foi atualizada usando BuscarMensagens
		msgs, err := db.BuscarMensagens(OpcoesConsulta{
			JID:    msg.JID,
			Limite: 10,
		})
		if err != nil {
			t.Fatalf("Erro ao buscar mensagens: %v", err)
		}
		
		encontrou := false
		for _, m := range msgs {
			if m.ID == id && m.Resposta == "Estou bem, obrigado!" {
				encontrou = true
				break
			}
		}
		
		if !encontrou {
			t.Errorf("Resposta não foi atualizada corretamente")
		}
	})
	
	// Testa buscar mensagens por JID
	t.Run("BuscarMensagens", func(t *testing.T) {
		// Limpa o banco para começar do zero
		_, err := db.conn.Exec("DELETE FROM mensagens")
		if err != nil {
			t.Fatalf("Erro ao limpar banco de dados: %v", err)
		}
		
		// Insere várias mensagens para o mesmo JID
		jid := "5511888888888@s.whatsapp.net"
		for i := 0; i < 5; i++ {
			msg := Mensagem{
				JID:       jid,
				Nome:      "Teste Busca",
				Conteudo:  "Mensagem de teste " + string(rune('A'+i)),
				Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
				Entrada:   i%2 == 0, // Alterna entre entrada e saída
			}
			
			_, err := db.SalvarMensagem(msg)
			if err != nil {
				t.Fatalf("Erro ao salvar mensagem para teste de busca: %v", err)
			}
		}
		
		// Busca as mensagens
		msgs, err := db.BuscarMensagens(OpcoesConsulta{
			JID:    jid,
			Limite: 10,
		})
		if err != nil {
			t.Errorf("Erro ao buscar mensagens por JID: %v", err)
		}
		
		if len(msgs) != 5 {
			t.Errorf("Esperava 5 mensagens, obteve %d", len(msgs))
		}
		
		// Testa limite
		msgsLimitadas, err := db.BuscarMensagens(OpcoesConsulta{
			JID:    jid,
			Limite: 2,
		})
		if err != nil {
			t.Errorf("Erro ao buscar mensagens limitadas: %v", err)
		}
		
		if len(msgsLimitadas) != 2 {
			t.Errorf("Esperava 2 mensagens limitadas, obteve %d", len(msgsLimitadas))
		}
		
		// Testa ObterUltimasMensagens
		ultimas, err := db.ObterUltimasMensagens(jid, 3)
		if err != nil {
			t.Errorf("Erro ao obter últimas mensagens: %v", err)
		}
		
		if len(ultimas) != 3 {
			t.Errorf("Esperava 3 últimas mensagens, obteve %d", len(ultimas))
		}
	})
	
	// Testa listar contatos únicos
	t.Run("BuscarContatos", func(t *testing.T) {
		// Limpa o banco para começar do zero
		_, err := db.conn.Exec("DELETE FROM mensagens")
		if err != nil {
			t.Fatalf("Erro ao limpar banco de dados: %v", err)
		}
		
		// Insere mensagens para diferentes contatos
		contatos := []struct {
			jid  string
			nome string
		}{
			{"5511111111111@s.whatsapp.net", "Contato 1"},
			{"5511222222222@s.whatsapp.net", "Contato 2"},
			{"5511333333333@s.whatsapp.net", "Contato 3"},
			{"5511111111111@s.whatsapp.net", "Contato 1"}, // Repete para testar deduplicação
		}
		
		for _, c := range contatos {
			msg := Mensagem{
				JID:       c.jid,
				Nome:      c.nome,
				Conteudo:  "Teste de contato único",
				Timestamp: time.Now(),
				Entrada:   true,
			}
			
			_, err := db.SalvarMensagem(msg)
			if err != nil {
				t.Fatalf("Erro ao salvar mensagem para teste de contatos: %v", err)
			}
		}
		
		// Lista os contatos únicos
		contatosUnicos, err := db.BuscarContatos()
		if err != nil {
			t.Errorf("Erro ao listar contatos únicos: %v", err)
		}
		
		// Deve ter 3 contatos únicos
		if len(contatosUnicos) != 3 {
			t.Errorf("Esperava 3 contatos únicos, obteve %d", len(contatosUnicos))
		}
		
		// Verifica se os JIDs estão corretos
		jids := make(map[string]bool)
		for _, c := range contatosUnicos {
			jids[c.JID] = true
		}
		
		if !jids["5511111111111@s.whatsapp.net"] || !jids["5511222222222@s.whatsapp.net"] || !jids["5511333333333@s.whatsapp.net"] {
			t.Errorf("Lista de contatos não contém todos os JIDs esperados")
		}
	})
	
	// Testa exclusão do histórico de um contato
	t.Run("ExcluirHistoricoContato", func(t *testing.T) {
		// Limpa o banco para começar do zero
		_, err := db.conn.Exec("DELETE FROM mensagens")
		if err != nil {
			t.Fatalf("Erro ao limpar banco de dados: %v", err)
		}
		
		// Adiciona mensagens para dois contatos diferentes
		jid1 := "5511444444444@s.whatsapp.net"
		jid2 := "5511555555555@s.whatsapp.net"
		
		// Adiciona mensagens para jid1
		for i := 0; i < 3; i++ {
			msg := Mensagem{
				JID:       jid1,
				Nome:      "Contato Exclusão",
				Conteudo:  "Mensagem para excluir",
				Timestamp: time.Now(),
				Entrada:   true,
			}
			
			_, err := db.SalvarMensagem(msg)
			if err != nil {
				t.Fatalf("Erro ao salvar mensagem: %v", err)
			}
		}
		
		// Adiciona mensagens para jid2
		for i := 0; i < 2; i++ {
			msg := Mensagem{
				JID:       jid2,
				Nome:      "Contato Manter",
				Conteudo:  "Mensagem para manter",
				Timestamp: time.Now(),
				Entrada:   true,
			}
			
			_, err := db.SalvarMensagem(msg)
			if err != nil {
				t.Fatalf("Erro ao salvar mensagem: %v", err)
			}
		}
		
		// Exclui histórico do primeiro contato
		err = db.ExcluirHistoricoContato(jid1)
		if err != nil {
			t.Errorf("Erro ao excluir histórico do contato: %v", err)
		}
		
		// Verifica se mensagens do jid1 foram excluídas
		msgs1, err := db.BuscarMensagens(OpcoesConsulta{JID: jid1})
		if err != nil {
			t.Errorf("Erro ao buscar mensagens após exclusão: %v", err)
		}
		
		if len(msgs1) > 0 {
			t.Errorf("Esperava 0 mensagens para o contato excluído, obteve %d", len(msgs1))
		}
		
		// Verifica se mensagens do jid2 foram mantidas
		msgs2, err := db.BuscarMensagens(OpcoesConsulta{JID: jid2})
		if err != nil {
			t.Errorf("Erro ao buscar mensagens do contato não excluído: %v", err)
		}
		
		if len(msgs2) != 2 {
			t.Errorf("Esperava 2 mensagens para o contato não excluído, obteve %d", len(msgs2))
		}
	})
	
	// Testa limpar todo o histórico
	t.Run("LimparHistorico", func(t *testing.T) {
		// Adiciona algumas mensagens
		for i := 0; i < 3; i++ {
			msg := Mensagem{
				JID:       "5511666666666@s.whatsapp.net",
				Nome:      "Teste Limpar",
				Conteudo:  "Mensagem para limpar",
				Timestamp: time.Now(),
				Entrada:   true,
			}
			
			_, err := db.SalvarMensagem(msg)
			if err != nil {
				t.Fatalf("Erro ao salvar mensagem: %v", err)
			}
		}
		
		// Limpa todo o histórico
		err = db.LimparHistorico()
		if err != nil {
			t.Errorf("Erro ao limpar histórico: %v", err)
		}
		
		// Verifica se todas as mensagens foram removidas
		msgs, err := db.BuscarMensagens(OpcoesConsulta{})
		if err != nil {
			t.Errorf("Erro ao buscar mensagens após limpar histórico: %v", err)
		}
		
		if len(msgs) > 0 {
			t.Errorf("Esperava 0 mensagens após limpar histórico, obteve %d", len(msgs))
		}
	})
}
