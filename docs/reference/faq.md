# FAQ Técnico - Distributed Ledger Go
## Perspectiva de Staff Engineering: Alta Escalabilidade, Baixa Latência e Robustez

Este FAQ aborda as decisões arquiteturais de nível crítico e as estratégias de resiliência adotadas no projeto para garantir uma operação de classe mundial em produção.

**Uso recomendado**

Este FAQ é um documento de referência e suporte rápido para dúvidas técnicas. Use-o para esclarecer decisões e manter consistência entre equipes.

**Status do documento**

- **Status:** FAQ Ativo (Living FAQ)
- **Proprietário:** Staff Engineering
- **Última atualização:** 2026-06-23

---

### 1. Consistência e Integridade de Dados

#### **Q: Como o sistema garante a consistência absoluta dos saldos em um ambiente distribuído e altamente concorrente?**
**R:** Utilizamos uma estratégia combinada de três camadas:
1.  **Ordenação Rígida via Kafka:** Todas as transações de uma mesma conta (`account_id`) são enviadas para a mesma partição do Kafka (usando o ID da conta como chave). Isso garante que o processamento seja sequencial para aquela conta específica, eliminando condições de corrida no nível do consumidor.
2.  **Controle de Concorrência Otimista (OCC):** Na tabela `accounts_balance`, utilizamos uma coluna `version`. Cada atualização verifica se a versão lida ainda é a mesma no momento do `UPDATE`. Se houver conflito, a transação falha e o consumidor realiza um retry.
3.  **Partidas Dobradas (Double-Entry):** Seguindo padrões contábeis rigorosos, cada movimento financeiro gera entradas simétricas de débito e crédito, permitindo auditoria total e garantindo que o dinheiro não "desapareça" ou seja "criado" indevidamente.

#### **Q: Por que adotar o princípio "Append-Only" no Ledger?**
**R:** A imutabilidade é o alicerce da confiança em sistemas financeiros. Ao proibir `UPDATE` e `DELETE` em lançamentos (`ledger_entries`), garantimos um rastro de auditoria inalterável. Erros são corrigidos exclusivamente por estornos (lançamentos de compensação). Além disso, do ponto de vista de performance, operações de `INSERT` são significativamente mais rápidas e geram menos fragmentação em índices do que atualizações frequentes de linhas existentes.

---

### 2. Resiliência e Tolerância a Falhas

#### **Q: O que acontece se o cluster do Kafka ficar indisponível? O sistema para?**
**R:** Não. Implementamos o **Outbox Pattern** com degradação graciosa. Se o Gateway (`transaction-gw`) falha ao publicar no Kafka, ele persiste a intenção da transação em um cache local ou tabela de contingência ("Outbox"). Assim que a conectividade é restabelecida, um worker de background processa essas mensagens pendentes, garantindo que nenhuma requisição do cliente seja perdida (Durabilidade > Disponibilidade imediata).

#### **Q: Como evitamos o processamento duplicado de transações em caso de reentregas do Kafka (At-Least-Once)?**
**R:** A idempotência é garantida no nível do banco de dados. Cada transação possui uma `idempotency_key` única enviada pelo cliente ou gerada no ingresso. A tabela `transactions` possui uma restrição `UNIQUE` nesta chave. Se o Kafka reentregar uma mensagem que já foi processada (por exemplo, após um crash do consumidor antes do commit do offset), o banco rejeitará o insert duplicado, preservando a integridade do estado.

#### **Q: Como o sistema garante a entrega de notificações sem impactar a latência da transação principal?**
**R:** Utilizamos o padrão de **Eventos Assíncronos**. O processamento principal da transação no `ledger-core` foca apenas na integridade financeira (débito, crédito e persistência). Somente após o commit bem-sucedido da transação no banco de dados, um evento de notificação é publicado em um tópico separado do Kafka (`notifications`). O `notification-service` consome este tópico de forma independente. Isso garante que a latência da notificação (que pode envolver chamadas a APIs externas lentas) não atrase o processamento financeiro e que falhas no serviço de notificação não causem rollbacks em transações financeiras já validadas.

---

### 3. Escalabilidade e Performance

#### **Q: Por que utilizar um Worker Pool no Notification Service em vez de simplesmente disparar Goroutines conforme a demanda?**
**R:** Embora as Goroutines sejam baratas, elas não são gratuitas. Em um pico de tráfego, disparar milhares de goroutines simultâneas pode levar ao esgotamento de memória (OOM) ou sobrecarga do agendador do Go. O **Worker Pool** nos permite definir um "backpressure" natural. Limitamos o paralelismo de acordo com a capacidade de processamento e os limites de taxa (Rate Limits) dos provedores externos, garantindo que o serviço permaneça estável sob carga extrema.

#### **Q: Como o sistema lida com a leitura de milhões de registros para reconciliação sem causar lentidão no banco de dados?**
**R:** Utilizamos o **Cursor Pagination** (Paginação por Cursor). Em vez de `OFFSET`, que obriga o banco a escanear todos os registros anteriores, buscamos registros onde `id > last_processed_id`. Isso mantém a performance constante em $O(1)$ independentemente do tamanho da tabela. Combinamos isso com o `errgroup` do Go para processar lotes em paralelo, respeitando sempre um limite estrito de conexões com o banco.

---

### 4. Arquitetura e Manutenibilidade

#### **Q: Qual a justificativa para o uso de Arquitetura Hexagonal, SOLID e Clean Architecture?**
**R:** A combinação desses padrões visa o **desacoplamento tecnológico** e a **manutenibilidade extrema**.
1.  **Arquitetura Hexagonal:** Isola o domínio (negócio) das dependências externas via Portas e Adaptadores.
2.  **SOLID:** Garante que o código seja fácil de estender sem quebrar funcionalidades existentes (OCP) e que as dependências sejam invertidas (DIP), facilitando a injeção de mocks em testes.
3.  **Clean Architecture:** Organiza o código em camadas concêntricas onde a dependência flui sempre para dentro (Domain).

Isso nos permite:
*   **Testabilidade:** Podemos testar 100% da lógica de negócio usando mocks, sem subir Docker ou bancos reais.
*   **Flexibilidade:** Se precisarmos trocar o PostgreSQL pelo Cassandra ou o Kafka pelo Pulsar, alteramos apenas os adaptadores (`internal/adapters`), mantendo a lógica de negócio intacta.

#### **Q: Como é tratada a observabilidade em um fluxo distribuído?**
**R:** Utilizamos **Logs Estruturados (slog)** com injeção de `correlation_id` (ou `trace_id`). Desde o momento que a requisição entra no `transaction-gw`, ela recebe um ID único que é propagado nos headers do Kafka e logs de todos os microserviços. Isso permite que, em caso de erro, possamos rastrear toda a jornada de uma transação específica através de múltiplos serviços em segundos.

---

### 5. Gestão de Recursos

#### **Q: Como o Rate Limiter garante atomicidade sem se tornar um gargalo de performance?**
**R:** Utilizamos **Scripts Lua no Redis**. Ao executar a lógica de incremento e verificação dentro de um script Lua, o Redis garante que essa operação seja atômica. Nenhuma outra requisição pode interferir entre a leitura do contador e o seu incremento. Para evitar latência de rede excessiva, o middleware é leve e possui um fallback para memória local (`sync.Map`) caso o Redis apresente instabilidade.

#### **Q: Como o sistema lida com "Hot Partitions" (muitas transações para a mesma conta)?**
**R:** Esta é uma limitação física do modelo sequencial. Se uma única conta recebe um volume de transações maior do que um único worker consegue processar, teremos um atraso (lag) naquela partição. Para mitigar isso em nível de Staff:
1.  Otimizamos o processamento do consumidor para ser o mais rápido possível (poucas operações de IO).
2.  Monitoramos o lag por partição.
3.  Em casos extremos de contas institucionais ("super contas"), implementamos estratégias de *sharding* de saldo ou subcontas, embora a complexidade contábil aumente.

#### **Q: Por que utilizar Server-Side Rendering (SSR) com `html/template` no Backoffice em vez de um framework moderno como React ou Vue?**
**R:** Em ferramentas internas de backoffice, priorizamos **Velocidade de Desenvolvimento, Segurança e Simplicidade Operacional**.
1.  **Simplicidade:** O uso do pacote nativo `html/template` elimina a necessidade de um pipeline complexo de build (Node.js, Webpack, NPM) e diminui o tamanho da imagem Docker.
2.  **Segurança:** SSR reduz a superfície de ataque ao não expor excessivamente a lógica de negócio e as APIs internas para o cliente (browser).
3.  **Performance Percebida:** Para ferramentas de auditoria e listagem de dados, o envio do HTML pronto para o browser resulta em uma interação rápida e previsível, sem o overhead de "hydrating" e chamadas assíncronas múltiplas no carregamento inicial.
4.  **Consistência:** O uso de Tailwind CSS via CDN nos permite manter uma interface profissional e responsiva sem a necessidade de compilação de CSS local.

#### **Q: Como o sistema gerencia configurações sensíveis e segrega ambientes (Dev, Staging, Prod)?**
**R:** Adotamos o padrão **12-Factor App** para configurações, utilizando variáveis de ambiente. Implementamos o suporte a arquivos `.env` através de um carregador centralizado no pacote `shared/config`. Isso permite:
1.  **Segurança:** Credenciais nunca são versionadas no código (via `.gitignore`).
2.  **Portabilidade:** O mesmo binário pode rodar localmente com um arquivo `.env` ou em um cluster Kubernetes onde as configurações são injetadas via `ConfigMaps` ou `Secrets`.
3.  **Padrão de Indústria:** O uso de arquivos `.env` facilita a integração com ferramentas de CI/CD e ambientes de desenvolvimento local padronizados.
