.PHONY: help install build run clean test docker-up docker-down docker-reset docker-logs migrate-up migrate-down migrate-drop migrate-force migrate-version migrate-create swagger dev

# Carregar variáveis de ambiente
include .env
export

# Variáveis
APP_NAME=potential-idiomas-api
BINARY_NAME=api
DOCKER_COMPOSE=docker-compose
MIGRATE=migrate
MIGRATIONS_PATH=./migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Alvo padrão
help:
	@echo "Comandos disponíveis:"
	@echo "  make install          Instalar dependências do Go"
	@echo "  make build            Compilar a aplicação"
	@echo "  make run              Executar a aplicação"
	@echo "  make dev              Executar com hot reload (requer air)"
	@echo "  make clean            Limpar arquivos de build"
	@echo "  make test             Executar testes"
	@echo ""
	@echo "Comandos Docker:"
	@echo "  make docker-up        Iniciar todos os containers Docker"
	@echo "  make docker-down      Parar todos os containers"
	@echo "  make docker-reset     Resetar ambiente Docker (remove volumes)"
	@echo "  make docker-logs      Exibir logs do Docker"
	@echo "  make docker-ps        Exibir containers em execução"
	@echo ""
	@echo "Comandos de migração:"
	@echo "  make migrate-up       Aplicar migrations pendentes"
	@echo "  make migrate-down     Reverter última migration"
	@echo "  make migrate-drop     Apagar todas as migrations (PERIGOSO)"
	@echo "  make migrate-force    Forçar versão da migration (ex: make migrate-force VERSION=1)"
	@echo "  make migrate-version  Mostrar versão atual da migration"
	@echo "  make migrate-create   Criar nova migration (ex: make migrate-create NAME=add_users)"
	@echo ""
	@echo "Documentação Swagger:"
	@echo "  make swagger          Gerar documentação Swagger"

# Instalar dependências
install:
	@echo "Instalando dependências..."
	go mod download
	go mod tidy
	@echo "Dependências instaladas com sucesso"

# Compilar a aplicação
build:
	@echo "Compilando aplicação..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Build concluído: bin/$(BINARY_NAME)"

# Rodar a aplicação
run: build
	@echo "Iniciando aplicação..."
	./bin/$(BINARY_NAME)

# Rodar com hot reload
dev:
	@echo "Iniciando servidor de desenvolvimento com hot reload..."
	air

# Limpar arquivos de build
clean:
	@echo "Limpando arquivos de build..."
	rm -rf bin/
	rm -rf tmp/
	go clean
	@echo "Limpeza concluída"

# Executar testes
test:
	@echo "Executando testes..."
	go test -v -cover ./...

# Testes com cobertura
test-coverage:
	@echo "Executando testes com relatório de cobertura..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório de cobertura gerado: coverage.html"

# Comandos Docker
docker-up:
	@echo "Iniciando containers Docker..."
	$(DOCKER_COMPOSE) up -d
	@echo "Aguardando serviços iniciarem..."
	@sleep 5
	@echo "Containers em execução"
	@$(MAKE) docker-ps

docker-down:
	@echo "Parando containers Docker..."
	$(DOCKER_COMPOSE) down
	@echo "Containers parados"

docker-reset:
	@echo "Resetando ambiente Docker..."
	$(DOCKER_COMPOSE) down -v
	@echo "Removendo volumes e containers órfãos..."
	docker volume prune -f
	@echo "Iniciando containers novamente..."
	$(DOCKER_COMPOSE) up -d
	@echo "Aguardando serviços iniciarem..."
	@sleep 5
	@echo "Ambiente Docker resetado com sucesso"
	@$(MAKE) migrate-up

docker-logs:
	@echo "Exibindo logs do Docker (Ctrl+C para sair)..."
	$(DOCKER_COMPOSE) logs -f

docker-logs-api:
	@echo "Exibindo logs da API (Ctrl+C para sair)..."
	$(DOCKER_COMPOSE) logs -f api

docker-logs-db:
	@echo "Exibindo logs do banco de dados (Ctrl+C para sair)..."
	$(DOCKER_COMPOSE) logs -f postgres

docker-ps:
	@echo "Containers em execução:"
	@$(DOCKER_COMPOSE) ps

docker-restart:
	@echo "Reiniciando containers Docker..."
	$(DOCKER_COMPOSE) restart
	@echo "Containers reiniciados"

# Comandos de migração
migrate-up:
	@echo "Aplicando migrations..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "Migrations aplicadas com sucesso"

migrate-down:
	@echo "Revertendo última migration..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "Migration revertida"

migrate-drop:
	@echo "ATENÇÃO: Isso irá apagar todas as tabelas e dados"
	@read -p "Tem certeza? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" drop -f; \
		echo "Todas as migrations foram apagadas"; \
	else \
		echo "Operação cancelada"; \
	fi

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERRO: VERSION é obrigatória"; \
		echo "Uso: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@echo "Forçando versão da migration para $(VERSION)..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(VERSION)
	@echo "Versão forçada para $(VERSION)"

migrate-version:
	@echo "Versão atual da migration:"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "ERRO: NAME é obrigatório"; \
		echo "Uso: make migrate-create NAME=add_users_table"; \
		exit 1; \
	fi
	@echo "Criando migration: $(NAME)"
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)
	@echo "Arquivos de migration criados"

# Acesso ao banco
db-connect:
	@echo "Conectando ao banco..."
	docker exec -it potential_db psql -U $(DB_USER) -d $(DB_NAME)

db-dump:
	@echo "Gerando dump do banco..."
	docker exec potential_db pg_dump -U $(DB_USER) $(DB_NAME) > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Dump criado com sucesso"

db-restore:
	@if [ -z "$(FILE)" ]; then \
		echo "ERRO: FILE é obrigatório"; \
		echo "Uso: make db-restore FILE=backup.sql"; \
		exit 1; \
	fi
	@echo "Restaurando banco a partir de $(FILE)..."
	docker exec -i potential_db psql -U $(DB_USER) -d $(DB_NAME) < $(FILE)
	@echo "Banco restaurado com sucesso"

# Swagger
swagger:
	@echo "Gerando documentação Swagger..."
	swag init -g cmd/api/main.go -o docs
	@echo "Documentação gerada em docs/"

# Helpers de desenvolvimento
setup: install docker-up migrate-up
	@echo "Ambiente de desenvolvimento configurado com sucesso"
	@echo "Banco de dados pronto com todas as migrations aplicadas"

reset: clean docker-reset
	@echo "Ambiente resetado com sucesso"

# Lint e formatação
lint:
	@echo "Executando linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatando código..."
	go fmt ./...
	goimports -w .

# Instalar ferramentas de desenvolvimento
install-tools:
	@echo "Instalando ferramentas de desenvolvimento..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Ferramentas instaladas"

# Build de produção
build-prod:
	@echo "Compilando para produção..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Build de produção concluído"

# Build Docker da API
docker-build:
	@echo "Construindo imagem Docker..."
	docker build -t $(APP_NAME):latest .
	@echo "Imagem Docker construída com sucesso"

# Reset total (cuidado!)
nuke: docker-down
	@echo "ATENÇÃO: Isso irá remover TODOS os containers, volumes e dados"
	@read -p "Tem certeza? Digite 'yes' para confirmar: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose down -v --remove-orphans; \
		docker system prune -f; \
		rm -rf bin/ tmp/; \
		echo "Ambiente completamente removido"; \
		echo "Execute 'make setup' para reinicializar"; \
	else \
		echo "Operação cancelada"; \
	fi

