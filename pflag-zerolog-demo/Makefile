APP_NAME := viper-env-demo
VERSION := v1.0.0
BUILD_DIR := ./bin

# Цвета для вывода
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RED := \033[31m
NC := \033[0m

.PHONY: help build clean demo test-env test-config test-priority examples

# Помощь
help: ## Показать это сообщение помощи
	@echo "$(CYAN)$(APP_NAME) - Viper + Environment Variables Demo$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  $(GREEN)%-20s$(NC) %s\n", $1, $2}' $(MAKEFILE_LIST)

# Установка зависимостей
deps: ## Установить зависимости
	@echo "$(YELLOW)Установка зависимостей...$(NC)"
	go mod download
	go mod tidy

# Сборка приложения
build: deps ## Собрать приложение
	@echo "$(YELLOW)Сборка $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.appVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "$(GREEN)✅ Сборка завершена: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# Очистка
clean: ## Очистить артефакты сборки
	@echo "$(YELLOW)Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean -cache

# Основная демонстрация
demo: build ## Запустить основную демонстрацию
	@echo "$(CYAN)🚀 Демонстрация Viper + Environment Variables$(NC)"
	@echo ""
	@$(BUILD_DIR)/$(APP_NAME)

# Тестирование переменных окружения
test-env: build ## Протестировать переменные окружения
	@echo "$(CYAN)🌍 Тестирование переменных окружения$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Базовый запуск (без переменных):$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --show-config
	@echo ""
	@echo "$(YELLOW)2. С переменными окружения:$(NC)"
	@VIPER_LOG_LEVEL=debug \
	 VIPER_SERVER_PORT=3000 \
	 VIPER_APP_ENVIRONMENT=staging \
	 $(BUILD_DIR)/$(APP_NAME) --show-config
	@echo ""
	@echo "$(YELLOW)3. Показать переменные окружения:$(NC)"
	@VIPER_LOG_LEVEL=trace \
	 VIPER_SERVER_HOST=0.0.0.0 \
	 VIPER_DATABASE_PASSWORD=secret123 \
	 $(BUILD_DIR)/$(APP_NAME) --show-env

# Тестирование конфигурационного файла
test-config: build ## Протестировать конфигурационный файл
	@echo "$(CYAN)📋 Тестирование конфигурационного файла$(NC)"
	@echo ""
	@echo "$(YELLOW)1. С конфигурационным файлом:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --config=config.yaml --show-config
	@echo ""
	@echo "$(YELLOW)2. Переопределение флагами:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --config=config.yaml --log-level=trace --debug --show-config

# Тестирование приоритета настроек
test-priority: build ## Протестировать приоритет настроек
	@echo "$(CYAN)🎯 Тестирование приоритета настроек$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Только значения по умолчанию:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME)
	@echo ""
	@echo "$(YELLOW)2. Переменные окружения переопределяют значения по умолчанию:$(NC)"
	@VIPER_LOG_LEVEL=debug \
	 VIPER_SERVER_PORT=4000 \
	 $(BUILD_DIR)/$(APP_NAME)
	@echo ""
	@echo "$(YELLOW)3. Флаги переопределяют переменные окружения:$(NC)"
	@VIPER_LOG_LEVEL=debug \
	 VIPER_SERVER_PORT=4000 \
	 $(BUILD_DIR)/$(APP_NAME) --log-level=trace --verbose

# Примеры использования
examples: build ## Показать примеры использования
	@echo "$(CYAN)📖 Примеры использования Viper с переменными окружения$(NC)"
	@echo ""
	@echo "$(YELLOW)Базовые команды:$(NC)"
	@echo "  $(BUILD_DIR)/$(APP_NAME)                    # Значения по умолчанию"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --show-config      # Показать конфигурацию"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --show-env         # Показать переменные окружения"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --help             # Справка"
	@echo ""
	@echo "$(YELLOW)Переменные окружения:$(NC)"
	@echo "  VIPER_LOG_LEVEL=debug $(BUILD_DIR)/$(APP_NAME)"
	@echo "  VIPER_SERVER_PORT=3000 $(BUILD_DIR)/$(APP_NAME)"
	@echo "  VIPER_APP_ENVIRONMENT=production $(BUILD_DIR)/$(APP_NAME)"
	@echo ""
	@echo "$(YELLOW)Комбинированные примеры:$(NC)"
	@echo "  VIPER_LOG_LEVEL=debug VIPER_SERVER_PORT=8080 $(BUILD_DIR)/$(APP_NAME) --verbose"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --config=config.yaml --log-level=trace"
	@echo ""
	@echo "$(YELLOW)Приоритет настроек (от высшего к низшему):$(NC)"
	@echo "  1. Флаги командной строки"
	@echo "  2. Переменные окружения (префикс VIPER_)"
	@echo "  3. Конфигурационный файл"
	@echo "  4. Значения по умолчанию"

# Создать .env файл из примера
create-env: ## Создать .env файл из .env.example
	@echo "$(YELLOW)Создание .env файла...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(GREEN)✅ .env файл создан из .env.example$(NC)"; \
		echo "$(YELLOW)Отредактируйте .env файл под свои нужды$(NC)"; \
	else \
		echo "$(RED)❌ .env файл уже существует$(NC)"; \
	fi

# Загрузить переменные из .env и запустить
run-with-env: build create-env ## Запустить с переменными из .env файла
	@echo "$(CYAN)🌍 Запуск с переменными из .env файла$(NC)"
	@echo ""
	@if [ -f .env ]; then \
		set -a && source .env && set +a && $(BUILD_DIR)/$(APP_NAME); \
	else \
		echo "$(RED)❌ .env файл не найден. Запустите 'make create-env' сначала.$(NC)"; \
	fi

# Демонстрация всех возможностей
demo-all: build ## Запустить полную демонстрацию
	@echo "$(CYAN)🎯 Полная демонстрация Viper + Environment Variables$(NC)"
	@echo ""
	@make test-env
	@echo ""
	@make test-config
	@echo ""
	@make test-priority

# Интерактивная демонстрация
interactive: build ## Интерактивная демонстрация
	@echo "$(CYAN)🎮 Интерактивная демонстрация$(NC)"
	@echo ""
	@echo "Попробуйте следующие команды:"
	@echo ""
	@make examples
	@echo ""
	@echo "Или запустите 'make demo-all' для полной демонстрации"
