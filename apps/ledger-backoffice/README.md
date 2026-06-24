

# **Software Design Document (SDD) — Componente:** ledger-backoffice

## **1\. Escopo e Propósito**

Este componente atua como a interface administrativa, de observabilidade e de auditoria contábil do ecossistema distributed-ledger-go. Ele provê uma visão unificada para o time de operações financeiras (Backoffice) inspecionar o rastro imutável de transações (*Audit Trail*), monitorar a saúde da consistência eventual gerada pelo reconciliador em lote e visualizar o estado das filas de falha (Dead Letter Queue \- DLQ).

## **2\. Padrões de Implementação e Stack Técnica**

* **Linguagem:** Go 1.22+  
* **Renderização:** Server-Side Rendering (SSR) nativo utilizando o pacote padrão html/template (sem frameworks SPA/Node.js para otimizar o tempo de desenvolvimento).  
* **Estilo Visual:** Interface limpa e responsiva utilizando o framework utilitário **Tailwind CSS** (via CDN).  
* **Arquitetura:** Padrão MVC Simples (Model-View-Controller) ou Portas/Adaptadores simplificados, mapeando queries HTTP diretamente para a réplica de leitura do banco de dados PostgreSQL e Redis.  
* **Roteador HTTP:** Pacote padrão net/http ou driver leve (ex: go-chi/chi ou gofiber/fiber).

## **3\. Especificação das Rotas HTTP e Visualizações**

### **Rota 1: Dashboard Geral e Saúde da Reconciliação**

* **Endpoint:** GET /admin/dashboard  
* **Propósito:** Exibir os indicadores macro de integridade do ecossistema.  
* **Dados necessários (Queries SQL):**  
  * Contagem total de contas cadastradas (SELECT COUNT(\*) FROM accounts\_balance).  
  * Volume total financeiro transacionado somando créditos da tabela imutável.  
  * Exibição do status do último Job do ledger-reconciler (buscando uma chave de metadados de controle persistida no Redis, ex: reconciler:last\_run\_status).  
* **Seção de Alertas:** Uma listagem em tabela vermelha exibindo contas que falharam na validação matemática do reconciliador em lote (onde $\\sum \\text{Créditos} \- \\sum \\text{Débitos} \\neq \\text{Saldo Corrente}$).

### **Rota 2: Busca por Conta e Rastro de Auditoria (Audit Trail)**

* **Endpoint:** GET /admin/accounts/{id}  
* **Propósito:** Inspecionar o extrato detalhado de um cliente para validação legal e suporte técnico.  
* **Contrato da View:**  
  * **Bloco Superior:** Exibe os dados correntes da tabela accounts\_balance (ID, Saldo Atual, Versão de Concorrência Otimista, Última Atualização).  
  * **Tabela Principal (O Rastro Imutável):** Listagem paginada por período contendo o histórico bruto extraído da tabela ledger\_entries. Cada linha deve exibir: ID da Entrada, ID da Transação Relacionada, Tipo (CREDIT ou DEBIT), Valor Formatado em Reais (convertendo o BIGINT de centavos para decimal) e Data de Criação com precisão de milissegundos.

### **Rota 3: Gerenciador de Mensagens e DLQ**

* **Endpoint:** GET /admin/dlq  
* **Propósito:** Visualizar anomalias de integração geradas por falhas de latência ou indisponibilidade de APIs de terceiros no notification-motor.  
* **Dados necessários:** Exibir o volume de mensagens represadas na fila de Dead Letter Queue (DLQ) do Kafka.

## **4\. Estrutura de Diretórios Sugerida (Para colar no prompt da IA)**

Plaintext  
/apps/ledger-backoffice  
├── /handlers         \# Controladores HTTP que tratam requisições e renderizam templates  
│   └── admin.go  
├── /repository       \# Queries puras de leitura do Postgres e conexões de metadados do Redis  
│   └── reader.go  
├── /web              \# Estrutura de visualização Server-Side  
│   ├── /templates  
│   │   ├── layout.html    \# Base estrutural contendo injeção do Tailwind CSS via CDN  
│   │   ├── dashboard.html \# View da Rota 1  
│   │   └── account.html   \# View da Rota 2  
│   └── /static            \# Assets estáticos opcionais  
├── main.go           \# Inicialização do servidor, graceful shutdown e injeção de dependências  
└── go.mod

## **5\. Exemplo de Código do Template Base (**layout.html**)**

Forneça este trecho para a IA entender como estruturar as páginas usando Tailwind nativo via CDN sem precisar compilar nada localmente:  
HTML  
{{define "layout"}}  
\<\!DOCTYPE **html**\>  
\<html lang\="pt-BR"\>  
\<head\>  
    \<meta charset\="UTF-8"\>  
    \<meta name\="viewport" content\="width=device-width, initial-scale=1.0"\>  
    \<title\>Distributed Ledger — Backoffice de Operações\</title\>  
    \<script src\="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"\>\</script\>  
\</head\>  
\<body class\="bg-slate-900 text-slate-100 font-sans"\>  
    \<nav class\="bg-slate-950 border-b border-slate-800 px-6 py-4"\>  
        \<div class\="max-w-7xl mx-auto flex justify-between items-center"\>  
            \<span class\="text-lg font-bold text-sky-400"\>Ledger Engine Admin\</span\>  
            \<div class\="space-x-4 text-sm font-medium"\>  
                \<a href\="/admin/dashboard" class\="text-slate-300 hover:text-white transition"\>Dashboard\</a\>  
                \<a href\="/admin/dlq" class\="text-slate-300 hover:text-white transition"\>Mensagens & DLQ\</a\>  
            \</div\>  
        \</div\>  
    \</nav\>  
    \<main class\="max-w-7xl mx-auto px-6 py-8"\>  
        {{template "content" .}}  
    \</main\>  
\</body\>  
\</html\>  
{{end}}

## **🚀 Como instruir a IA para gerar o código:**

Copie o conteúdo acima e use o seguinte prompt complementar:  
*"Com base no SDD fornecido acima, escreva a implementação do microserviço* ledger-backoffice *em Go. Use o pacote nativo* html/template *para renderização. O repositório deve ler diretamente as tabelas do PostgreSQL conforme a DDL definida no projeto. Garanta que o tratamento de erros em Go seja explícito e limpo, sem pânico em tempo de execução."*  
Isso vai acelerar o seu desenvolvimento em pouquíssimos minutos, mantendo a coerência técnica intocável para as suas revisões de código offline.  
