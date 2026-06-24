package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Load carrega as variáveis de ambiente de um arquivo .env, se existir.
func Load() {
	err := godotenv.Load()
	if err != nil {
		// Não é um erro fatal se o arquivo .env não existir, pois as variáveis
		// podem estar definidas diretamente no ambiente (ex: Docker, K8s).
		slog.Debug("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente do sistema.")
	}
}

// GetEnv retorna o valor de uma variável de ambiente ou um valor padrão se estiver vazia.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
