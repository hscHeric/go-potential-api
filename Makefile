.PHONY: help run build install clean docker-up docker-down docker-logs docker-reset swagger

APP_NAME=potential-api
MAIN_PATH=./cmd/api

help:
	@echo "Available commands:"
	@echo "  make install      - Instala as dependências do projeto"
	@echo "  make run          - Roda a aplicação"
	@echo "  make build        - Gera o binário da aplicação"
	@echo "  make docker-up    - Inicia o container do banco de dados"
	@echo "  make docker-down  - Derruba o container do banco de dados"
	@echo "  make docker-logs  - Ver os logs do container do banco de dados"
	@echo "  make docker-reset - Reseta o container do banco de dados"
	@echo "  make swagger      - Gera a documentação Swagger"
	@echo "  make clean        - Limpa os arquivos binários gerados"

install:
	@go mod download
	@go mod tidy

run:
	@go run $(MAIN_PATH)/main.go

build:
	@go build -o $(APP_NAME) $(MAIN_PATH)/main.go

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-reset:
	@docker-compose down -v
	@docker-compose up -d

swagger:
	@swag init -g cmd/api/main.go -o docs

clean:
	@rm -f $(APP_NAME)
	@go clean
