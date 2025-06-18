package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel   string
	logFormat  string
	prettify   bool
)

// rootCmd основная команда
var rootCmd = &cobra.Command{
	Use:   "zerolog-demo",
	Short: "Демонстрация zerolog с различными уровнями логирования",
	Long: `Простое приложение для демонстрации zerolog с уровнями:
• trace - самый детальный уровень
• debug - отладочная информация
• info  - общая информация (по умолчанию)
• warn  - предупреждения
• error - ошибки`,
	Example: `  # Запуск с разными уровнями логирования
  zerolog-demo --log-level=trace
  zerolog-demo --log-level=debug
  zerolog-demo --log-level=info
  zerolog-demo --log-level=warn
  zerolog-demo --log-level=error

  # Разные форматы вывода
  zerolog-demo --log-format=json
  zerolog-demo --log-format=console
  zerolog-demo --prettify`,
	RunE: runDemo,
}

func init() {
	// Флаги для настройки логирования
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Уровень логирования (trace, debug, info, warn, error)")
	rootCmd.Flags().StringVar(&logFormat, "log-format", "console", "Формат логов (json, console)")
	rootCmd.Flags().BoolVar(&prettify, "prettify", false, "Красивый консольный вывод с цветами")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func runDemo(cmd *cobra.Command, args []string) error {
	// Настройка zerolog
	setupLogging()

	// Демонстрация всех уровней логирования
	demonstrateLogLevels()

	// Демонстрация структурированного логирования
	demonstrateStructuredLogging()

	// Демонстрация контекстного логирования
	demonstrateContextualLogging()

	return nil
}

// setupLogging настраивает zerolog
func setupLogging() {
	// Парсим уровень логирования
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		fmt.Printf("Неверный уровень логирования '%s', используется 'info'\n", logLevel)
		level = zerolog.InfoLevel
	}

	// Устанавливаем глобальный уровень
	zerolog.SetGlobalLevel(level)

	// Настраиваем формат вывода
	if logFormat == "console" || prettify {
		// Красивый консольный вывод
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}

		if prettify {
			// Добавляем цвета для уровней
			output.FormatLevel = func(i interface{}) string {
				level := fmt.Sprintf("%s", i)
				switch level {
				case "trace":
					return fmt.Sprintf("\033[90m%-5s\033[0m", "TRACE") // Серый
				case "debug":
					return fmt.Sprintf("\033[36m%-5s\033[0m", "DEBUG") // Циан
				case "info":
					return fmt.Sprintf("\033[32m%-5s\033[0m", "INFO")  // Зеленый
				case "warn":
					return fmt.Sprintf("\033[33m%-5s\033[0m", "WARN")  // Желтый
				case "error":
					return fmt.Sprintf("\033[31m%-5s\033[0m", "ERROR") // Красный
				default:
					return fmt.Sprintf("%-5s", level)
				}
			}
		}

		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// JSON формат
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Добавляем информацию о приложении
	log.Logger = log.With().
		Str("app", "zerolog-demo").
		Str("version", "1.0.0").
		Logger()
}

// demonstrateLogLevels демонстрирует все уровни логирования
func demonstrateLogLevels() {
	fmt.Println("\n🎯 Демонстрация уровней логирования:")
	fmt.Printf("Текущий уровень: %s\n\n", logLevel)

	// Trace - самый детальный уровень
	log.Trace().
		Str("function", "demonstrateLogLevels").
		Msg("Это TRACE сообщение - самый детальный уровень логирования")

	// Debug - отладочная информация
	log.Debug().
		Int("step", 1).
		Str("action", "demonstration").
		Msg("Это DEBUG сообщение - отладочная информация")

	// Info - общая информация (по умолчанию)
	log.Info().
		Str("status", "running").
		Msg("Это INFO сообщение - общая информация о работе приложения")

	// Warn - предупреждения
	log.Warn().
		Str("issue", "deprecated_function").
		Msg("Это WARN сообщение - предупреждение о потенциальной проблеме")

	// Error - ошибки
	log.Error().
		Str("error_type", "demo_error").
		Msg("Это ERROR сообщение - информация об ошибке")
}

// demonstrateStructuredLogging демонстрирует структурированное логирование
func demonstrateStructuredLogging() {
	fmt.Println("\n📊 Демонстрация структурированного логирования:")

	// Логирование с различными типами данных
	log.Info().
		Str("user_id", "user123").
		Int("age", 25).
		Float64("balance", 1234.56).
		Bool("is_premium", true).
		Time("login_time", time.Now()).
		Dur("session_duration", 45*time.Minute).
		Strs("roles", []string{"user", "admin"}).
		Msg("Пользователь вошел в систему")

	// Логирование HTTP запроса
	log.Info().
		Str("method", "GET").
		Str("path", "/api/users").
		Int("status_code", 200).
		Dur("response_time", 150*time.Millisecond).
		Int("response_size", 1024).
		Str("user_agent", "Mozilla/5.0").
		Msg("HTTP запрос обработан")

	// Логирование ошибки с контекстом
	log.Error().
		Err(fmt.Errorf("connection timeout")).
		Str("service", "database").
		Str("host", "db.example.com").
		Int("port", 5432).
		Int("retry_count", 3).
		Msg("Ошибка подключения к базе данных")
}

// demonstrateContextualLogging демонстрирует контекстное логирование
func demonstrateContextualLogging() {
	fmt.Println("\n🔗 Демонстрация контекстного логирования:")

	// Создаем логгер с контекстом
	contextLogger := log.With().
		Str("request_id", "req-12345").
		Str("user_id", "user456").
		Logger()

	contextLogger.Info().Msg("Начало обработки запроса")

	// Функция с собственным контекстом
	processOrder(contextLogger)

	contextLogger.Info().Msg("Завершение обработки запроса")

	// Демонстрация подкомпонентов
	demonstrateSubComponents()
}

// processOrder симулирует обработку заказа
func processOrder(logger zerolog.Logger) {
	orderLogger := logger.With().
		Str("component", "order_processor").
		Str("order_id", "order-789").
		Logger()

	orderLogger.Debug().Msg("Валидация заказа")

	orderLogger.Info().
		Float64("amount", 99.99).
		Str("currency", "USD").
		Msg("Заказ валиден, начинаем обработку")

	// Симуляция обработки
	time.Sleep(100 * time.Millisecond)

	orderLogger.Warn().
		Int("inventory_level", 5).
		Msg("Низкий уровень запасов")

	orderLogger.Info().
		Str("status", "completed").
		Msg("Заказ успешно обработан")
}

// demonstrateSubComponents демонстрирует логирование компонентов
func demonstrateSubComponents() {
	// База данных
	dbLogger := log.With().
		Str("component", "database").
		Str("table", "users").
		Logger()

	dbLogger.Debug().
		Str("query", "SELECT * FROM users WHERE id = ?").
		Interface("params", []interface{}{123}).
		Msg("Выполнение SQL запроса")

	dbLogger.Info().
		Int("rows_affected", 1).
		Dur("query_time", 25*time.Millisecond).
		Msg("Запрос выполнен успешно")

	// Кеш
	cacheLogger := log.With().
		Str("component", "cache").
		Str("key", "user:123").
		Logger()

	cacheLogger.Debug().Msg("Поиск в кеше")
	cacheLogger.Info().
		Bool("hit", false).
		Msg("Промах кеша, данные загружены из БД")

	// Внешний сервис
	apiLogger := log.With().
		Str("component", "external_api").
		Str("service", "payment_gateway").
		Logger()

	apiLogger.Info().
		Str("endpoint", "https://api.payments.com/charge").
		Str("method", "POST").
		Msg("Отправка запроса к внешнему API")

	apiLogger.Error().
		Int("status_code", 503).
		Str("error", "Service Unavailable").
		Msg("Внешний сервис недоступен")
}
