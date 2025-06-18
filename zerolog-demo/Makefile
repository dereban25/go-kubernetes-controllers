APP_NAME := zerolog-demo
VERSION := v1.0.0
BUILD_DIR := ./bin

# Цвета для вывода
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RED := \033[31m
NC := \033[0m

.PHONY: help build clean run-all test-levels test-formats

# Помощь
help: ## Показать это сообщение помощи
	@echo "$(CYAN)$(APP_NAME) - Zerolog Demo$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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

# Очистка
clean: ## Очистить артефакты сборки
	@echo "$(YELLOW)Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean -cache

# Запуск с разными уровнями
run-all: build ## Запустить демо со всеми уровнями логирования
	@echo "$(CYAN)🎯 Демонстрация всех уровней логирования$(NC)"
	@echo ""
	@echo "$(YELLOW)1. TRACE уровень (самый детальный):$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=trace --prettify
	@echo ""
	@echo "$(YELLOW)2. DEBUG уровень:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=debug --prettify
	@echo ""
	@echo "$(YELLOW)3. INFO уровень (по умолчанию):$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=info --prettify
	@echo ""
	@echo "$(YELLOW)4. WARN уровень:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=warn --prettify
	@echo ""
	@echo "$(YELLOW)5. ERROR уровень (только ошибки):$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=error --prettify

# Тестирование уровней
test-levels: build ## Протестировать все уровни логирования
	@echo "$(CYAN)🧪 Тестирование уровней логирования$(NC)"
	@echo ""
	@for level in trace debug info warn error; do \
		echo "$(YELLOW)Тестирование уровня: $$level$(NC)"; \
		$(BUILD_DIR)/$(APP_NAME) --log-level=$$level --prettify; \
		echo ""; \
	done

# Тестирование форматов
test-formats: build ## Протестировать форматы вывода
	@echo "$(CYAN)📊 Тестирование форматов вывода$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Консольный формат с цветами:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-format=console --prettify
	@echo ""
	@echo "$(YELLOW)2. Консольный формат без цветов:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-format=console
	@echo ""
	@echo "$(YELLOW)3. JSON формат:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-format=json

# Демонстрация
demo: build ## Запустить полную демонстрацию
	@echo "$(CYAN)🚀 Полная демонстрация zerolog$(NC)"
	@echo ""
	@echo "$(YELLOW)Запуск с красивым консольным выводом:$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --log-level=debug --prettify

# Примеры использования
examples: build ## Показать примеры использования
	@echo "$(CYAN)📖 Примеры использования$(NC)"
	@echo ""
	@echo "$(YELLOW)Основные команды:$(NC)"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --log-level=trace"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --log-level=debug --prettify"
	@echo "  $(BUILD_DIR)/$(APP_NAME) --log-format=json"
	@echo ""
	@echo "$(YELLOW)Уровни логирования:$(NC)"
	@echo "  trace - Самая детальная информация"
	@echo "  debug - Отладочная информация"
	@echo "  info  - Общая информация (по умолчанию)"
	@echo "  warn  - Предупреждения"
	@echo "  error - Только ошибки"
	@echo ""
	@echo "$(YELLOW)Форматы вывода:$(NC)"
	@echo "  console - Читаемый консольный формат"
	@echo "  json    - Структурированный JSON"
	@echo ""
	@echo "$(YELLOW)Дополнительные опции:$(NC)"
	@echo "  --prettify - Красивый вывод с цветами"
