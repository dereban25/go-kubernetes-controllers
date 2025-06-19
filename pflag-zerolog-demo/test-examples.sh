#!/bin/bash

# test-examples.sh - Примеры тестирования Viper + Environment Variables

echo "🧪 Примеры тестирования Viper + Environment Variables"
echo ""

APP="./bin/viper-env-demo"

if [ ! -f "$APP" ]; then
    echo "❌ Приложение не найдено. Запустите 'make build' сначала."
    exit 1
fi

echo "1. 📋 Базовая конфигурация (значения по умолчанию):"
echo "   $APP --show-config"
echo ""

echo "2. 🌍 Переменные окружения:"
echo "   VIPER_LOG_LEVEL=debug $APP"
echo "   VIPER_SERVER_PORT=3000 VIPER_APP_ENVIRONMENT=staging $APP"
echo "   VIPER_DATABASE_PASSWORD=secret123 $APP --show-env"
echo ""

echo "3. 📄 Конфигурационный файл:"
echo "   $APP --config=config.yaml --show-config"
echo ""

echo "4. 🎯 Приоритет настроек:"
echo "   # Переменная окружения переопределяет значение по умолчанию:"
echo "   VIPER_LOG_LEVEL=debug $APP"
echo ""
echo "   # Флаг переопределяет переменную окружения:"
echo "   VIPER_LOG_LEVEL=info $APP --log-level=debug"
echo ""

echo "5. 🔄 Комбинированные примеры:"
echo "   VIPER_LOG_LEVEL=debug VIPER_SERVER_PORT=8080 $APP --verbose --show-config"
echo "   $APP --config=config.yaml --log-level=trace --debug"
echo ""

echo "6. 🔍 Полная демонстрация:"
echo "   make demo-all"
echo ""

echo "Запустите любую из команд выше для тестирования!"
echo ""
echo "💡 Совет: используйте 'make examples' для быстрого просмотра всех команд"
