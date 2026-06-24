# Módulo Shared

O módulo `shared` é um pacote compartilhado que fornece funcionalidades comuns a todos os microserviços do monorepo. Ele visa reduzir a duplicação de código e garantir a consistência em aspectos transversais como logging, configuração e comunicação.

## Estrutura do Módulo

O módulo está organizado nas seguintes subpastas:

### 1. `config`
Utilitários para carregamento de configurações a partir de variáveis de ambiente.
- Fornece uma estrutura base de configuração que pode ser estendida pelos serviços.
- Suporta leitura automática de `.env` (se presente).

### 2. `events`
Definição de tipos comuns para eventos de domínio.
- `Transaction`: Estrutura base para eventos de intenção de transação.
- `Notification`: Estrutura para eventos de notificação.

### 3. `health`
Implementação de health checks padronizados.
- Facilita a exposição de endpoints `/healthz` que verificam a saúde do serviço e suas dependências.

### 4. `kafka`
Abstrações para produtores e consumidores Kafka.
- Configuração simplificada utilizando a biblioteca `segmentio/kafka-go`.
- Implementação de produtor com suporte a retries e logs integrados.

### 5. `logger`
Configuração padronizada do logger utilizando `slog` (Structured Logging).
- Garante que todos os logs sigam o mesmo formato (JSON em produção).
- Inclui campos padrão como `service_name`.

## Como Utilizar

Para utilizar o módulo `shared` em um novo serviço, adicione a dependência no arquivo `go.mod` do serviço e certifique-se de que o serviço está incluído no `go.work` na raiz do projeto.

### Exemplo: Inicializando o Logger

```go
import "github.com/distributed-ledger-go/shared/logger"

func main() {
    log := logger.New("meu-servico", "info")
    log.Info("Serviço inicializado")
}
```

### Exemplo: Carregando Configuração

```go
import "github.com/distributed-ledger-go/shared/config"

type MyConfig struct {
    config.BaseConfig
    CustomField string `env:"CUSTOM_FIELD"`
}

func main() {
    var cfg MyConfig
    config.Load(&cfg)
}
```

## Diretrizes de Contribuição

- O módulo `shared` não deve conter lógica de negócio específica de um domínio.
- Evite adicionar dependências pesadas que não serão utilizadas pela maioria dos serviços.
- Mantenha a compatibilidade com versões anteriores sempre que possível.
