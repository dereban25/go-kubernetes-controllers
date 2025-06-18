# Переменные
APP_NAME := k8s-cli
VERSION := v1.0.0
BUILD_DIR := ./bin
GO_VERSION := 1.21

# Цвета для вывода
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
NC := \033[0m

.PHONY: help build install clean test run deps examples

# Помощь
help: ## Показать это сообщение помощи
	@echo "$(CYAN)$(APP_NAME) - Kubernetes CLI Tool$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  $(GREEN)%-15s$(NC) %s\n", $1, $2}' $(MAKEFILE_LIST)

# Установка зависимостей
deps: ## Установить зависимости
	@echo "$(YELLOW)Установка зависимостей...$(NC)"
	go mod download
	go mod tidy

# Сборка приложения
build: deps ## Собрать приложение
	@echo "$(YELLOW)Сборка $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "$(GREEN)✅ Сборка завершена: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# Установка в систему
install: build ## Установить в систему
	@echo "$(YELLOW)Установка $(APP_NAME)...$(NC)"
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ $(APP_NAME) установлен в /usr/local/bin/$(NC)"

# Запуск в режиме разработки
run: build ## Запустить приложение
	@echo "$(YELLOW)Запуск $(APP_NAME)...$(NC)"
	$(BUILD_DIR)/$(APP_NAME)

# Тестирование
test: ## Запустить тесты
	@echo "$(YELLOW)Запуск тестов...$(NC)"
	go test -v ./...

# Очистка
clean: ## Очистить артефакты сборки
	@echo "$(YELLOW)Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean -cache

# Проверка окружения
check-env: ## Проверить окружение
	@echo "$(YELLOW)Проверка окружения...$(NC)"
	@which kubectl > /dev/null || (echo "❌ kubectl не найден" && exit 1)
	@kubectl cluster-info > /dev/null || (echo "❌ Нет доступа к кластеру Kubernetes" && exit 1)
	@echo "$(GREEN)✅ Окружение готово$(NC)"

# Форматирование кода
fmt: ## Форматировать код
	@echo "$(YELLOW)Форматирование кода...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Код отформатирован$(NC)"

# Демонстрация
demo: build check-env ## Запустить демонстрацию
	@echo "$(CYAN)🚀 Демонстрация k8s-cli$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Показать текущий контекст:$(NC)"
	$(BUILD_DIR)/$(APP_NAME) context current
	@echo ""
	@echo "$(YELLOW)2. Список всех контекстов:$(NC)"
	$(BUILD_DIR)/$(APP_NAME) context list
	@echo ""
	@echo "$(YELLOW)3. Список подов:$(NC)"
	$(BUILD_DIR)/$(APP_NAME) list pods
	@echo ""
	@echo "$(YELLOW)4. Список сервисов:$(NC)"
	$(BUILD_DIR)/$(APP_NAME) list services
