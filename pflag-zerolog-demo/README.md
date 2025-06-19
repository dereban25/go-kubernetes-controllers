# Viper + Environment Variables Demo

Демонстрационное приложение, показывающее использование **Viper** для работы с переменными окружения, конфигурационными файлами и приоритетом настроек.

## 🎯 Цель проекта

Показать как использовать Viper для:
- Работы с переменными окружения
- Загрузки конфигурационных файлов
- Управления приоритетом настроек
- Динамического обновления конфигурации
- Интеграции с pflag для CLI

## 🚀 Быстрый старт

```bash
# Сборка
make build

# Основная демонстрация
make demo

# Тестирование всех возможностей
make demo-all

# Показать примеры
make examples
```

## 📋 Возможности

### 1. Приоритет настроек (от высшего к низшему)
1. **Флаги командной строки** - `--log-level=debug`
2. **Переменные окружения** - `VIPER_LOG_LEVEL=debug`
3. **Конфигурационный файл** - `config.yaml`
4. **Значения по умолчанию** - встроенные в код

### 2. Переменные окружения
Все настройки могут быть заданы через переменные окружения с префиксом `VIPER_`:

```bash
# Настройки логирования
VIPER_LOG_LEVEL=debug              # Уровень логирования
VIPER_LOG_FORMAT=json              # Формат логов
VIPER_LOG_PRETTY=true              # Красивый вывод

# Настройки сервера
VIPER_SERVER_HOST=0.0.0.0          # Хост сервера
VIPER_SERVER_PORT=3000             # Порт сервера
VIPER_SERVER_TLS_ENABLED=true      # Включить TLS

# Настройки базы данных
VIPER_DATABASE_HOST=db.prod.com    # Хост БД
VIPER_DATABASE_PASSWORD=secret     # Пароль БД
VIPER_DATABASE_MAX_CONNECTIONS=50  # Макс. соединений

# Настройки приложения
VIPER_APP_ENVIRONMENT=production   # Окружение
VIPER_APP_DEBUG=false              # Режим отладки
```

### 3. Конфигурационный файл
Поддерживается `config.yaml`:

```yaml
log:
  level: info
  format: console
  pretty: true

server:
  host: localhost
  port: 8080
  tls:
    enabled: false

database:
  host: localhost
  port: 5432
  password: ""
  max_connections: 25

app:
  environment: development
  debug: false
  features:
    - logging
    - metrics
```

## 🎯 Примеры использования

### Базовые команды
```bash
# Запуск с настройками по умолчанию
./bin/viper-env-demo

# Показать текущую конфигурацию
./bin/viper-env-demo --show-config

# Показать переменные окружения
./bin/viper-env-demo --show-env

# Справка
./bin/viper-env-demo --help
```

### Переменные окружения
```bash
# Установить уровень логирования через переменную окружения
VIPER_LOG_LEVEL=debug ./bin/viper-env-demo

# Несколько переменных
VIPER_LOG_LEVEL=trace VIPER_SERVER_PORT=3000 ./bin/viper-env-demo

# Production конфигурация
VIPER_APP_ENVIRONMENT=production \
VIPER_LOG_FORMAT=json \
VIPER_SERVER_HOST=0.0.0.0 \
./bin/viper-env-demo
```

### Конфигурационный файл
```bash
# Использовать конфигурационный файл
./bin/viper-env-demo --config=config.yaml

# Переопределить настройки из файла флагами
./bin/viper-env-demo --config=config.yaml --log-level=debug --verbose
```

### Демонстрация приоритета
```bash
# 1. Только значения по умолчанию
./bin/viper-env-demo

# 2. Переменные окружения переопределяют значения по умолчанию
VIPER_LOG_LEVEL=debug ./bin/viper-env-demo

# 3. Флаги переопределяют переменные окружения
VIPER_LOG_LEVEL=info ./bin/viper-env-demo --log-level=debug
```

## 🔧 Команды Makefile

```bash
make help           # Показать справку
make build          # Собрать приложение
make demo           # Основная демонстрация
make test-env       # Тестировать переменные окружения
make test-config    # Тестировать конфигурационный файл
make test-priority  # Тестировать приоритет настроек
make examples       # Показать примеры команд
make create-env     # Создать .env файл из примера
make run-with-env   # Запустить с переменными из .env
make demo-all       # Полная демонстрация
make interactive    # Интерактивная демонстрация
make clean          # Очистить артефакты
```

## 📊 Что демонстрирует приложение

### 1. Приоритет настроек
```bash
# Демонстрирует как флаги переопределяют переменные окружения
VIPER_LOG_LEVEL=info ./bin/viper-env-demo --log-level=debug
# Результат: будет использован debug (флаг имеет приоритет)
```

### 2. Автоматическое чтение переменных окружения
```bash
# Viper автоматически находит переменные с префиксом VIPER_
export VIPER_SERVER_PORT=3000
./bin/viper-env-demo --show-config
# Результат: порт будет 3000
```

### 3. Вложенные структуры конфигурации
```bash
# Переменные с точками становятся вложенными структурами
VIPER_SERVER_TLS_ENABLED=true ./bin/viper-env-demo
# Соответствует: config.Server.TLS.Enabled = true
```

### 4. Динамическое обновление
```go
// Во время выполнения можно менять настройки
viper.Set("log.level", "debug")
newLevel := viper.GetString("log.level")
```

## 🎓 Применение в реальных проектах

Этот паттерн используется в:
- **Kubernetes** - все компоненты используют подобный подход
- **Docker** - переменные окружения для настройки
- **Prometheus** - конфигурационные файлы + переменные окружения
- **Grafana** - гибридная конфигурация
- **Microservices** - стандартный подход для 12-factor apps

### Преимущества:
1. **12-Factor App compliance** - стандартный подход
2. **Гибкость** - множество способов конфигурации
3. **Безопасность** - пароли через переменные окружения
4. **Удобство** - простое переопределение в разных окружениях
5. **Валидация** - структурированная конфигурация

## 🔐 Безопасность

### Рекомендации:
```bash
# ✅ Пароли через переменные окружения
VIPER_DATABASE_PASSWORD=secret ./bin/viper-env-demo

# ✅ Конфиденциальные данные не в git
echo ".env" >> .gitignore

# ✅ Маскирование паролей в логах
# Приложение автоматически маскирует пароли при выводе конфигурации
```

## 📖 Структура кода

### Конфигурационные структуры
```go
type Config struct {
    Log      LogConfig      `mapstructure:"log"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    App      AppConfig      `mapstructure:"app"`
}
```

### Инициализация Viper
```go
func initConfig() error {
    // Значения по умолчанию
    setDefaults()

    // Переменные окружения
    viper.SetEnvPrefix("VIPER")
    viper.AutomaticEnv()

    // Конфигурационный файл
    viper.ReadInConfig()

    // Переопределение флагами
    if *logLevel != "" {
        viper.Set("log.level", *logLevel)
    }

    return nil
}
```

## 🚀 Расширение

Легко добавить новые настройки:
```go
// 1. Добавить в структуру
type NewConfig struct {
    Feature string `mapstructure:"feature"`
}

// 2. Установить значение по умолчанию
viper.SetDefault("new.feature", "enabled")

// 3. Использовать переменную окружения
// VIPER_NEW_FEATURE=disabled
```

Это отличный пример современного подхода к конфигурации Go приложений! 🎯
