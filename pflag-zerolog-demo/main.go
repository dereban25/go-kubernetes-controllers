package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Структура конфигурации
type Config struct {
	// Логирование
	Log LogConfig `mapstructure:"log"`

	// Сервер
	Server ServerConfig `mapstructure:"server"`

	// База данных
	Database DatabaseConfig `mapstructure:"database"`

	// Приложение
	App AppConfig `mapstructure:"app"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Caller     bool   `mapstructure:"caller"`
	Timestamp  bool   `mapstructure:"timestamp"`
	Pretty     bool   `mapstructure:"pretty"`
	NoColor    bool   `mapstructure:"no_color"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	MaxConns int    `mapstructure:"max_connections"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type AppConfig struct {
	Name        string   `mapstructure:"name"`
	Version     string   `mapstructure:"version"`
	Environment string   `mapstructure:"environment"`
	Debug       bool     `mapstructure:"debug"`
	Features    []string `mapstructure:"features"`
}

// Глобальные переменные
var (
	config *Config

	// Флаги командной строки
	configFile = pflag.StringP("config", "c", "", "Path to config file")
	logLevel   = pflag.StringP("log-level", "l", "", "Log level (trace, debug, info, warn, error)")
	verbose    = pflag.BoolP("verbose", "v", false, "Enable verbose logging")
	debug      = pflag.Bool("debug", false, "Enable debug mode")
	help       = pflag.BoolP("help", "h", false, "Show help")
	showEnv    = pflag.Bool("show-env", false, "Show environment variables")
	showConfig = pflag.Bool("show-config", false, "Show current configuration")
)

const (
	appName = "viper-env-demo"
)

func main() {
	// Парсим флаги
	pflag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// Инициализируем конфигурацию
	if err := initConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Настраиваем логирование
	if err := setupLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up logging: %v\n", err)
		os.Exit(1)
	}

	// Показать переменные окружения если запрошено
	if *showEnv {
		showEnvironmentVariables()
		return
	}

	// Показать конфигурацию если запрошено
	if *showConfig {
		showCurrentConfig()
		return
	}

	// Запускаем демонстрацию
	runDemo()
}

func showHelp() {
	fmt.Printf("%s - Демонстрация Viper с переменными окружения\n\n", appName)

	fmt.Println("USAGE:")
	fmt.Printf("  %s [flags]\n\n", appName)

	fmt.Println("FLAGS:")
	fmt.Println("  -c, --config string      Path to config file")
	fmt.Println("  -l, --log-level string   Log level (trace, debug, info, warn, error)")
	fmt.Println("  -v, --verbose            Enable verbose logging")
	fmt.Println("      --debug              Enable debug mode")
	fmt.Println("      --show-env           Show environment variables")
	fmt.Println("      --show-config        Show current configuration")
	fmt.Println("  -h, --help               Show help")
	fmt.Println()

	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  Все настройки можно задать через переменные окружения с префиксом VIPER_")
	fmt.Println()
	fmt.Println("  Примеры:")
	fmt.Println("    VIPER_LOG_LEVEL=debug              # Уровень логирования")
	fmt.Println("    VIPER_LOG_FORMAT=json              # Формат логов")
	fmt.Println("    VIPER_LOG_PRETTY=true              # Красивый вывод")
	fmt.Println("    VIPER_SERVER_HOST=0.0.0.0          # Хост сервера")
	fmt.Println("    VIPER_SERVER_PORT=8080              # Порт сервера")
	fmt.Println("    VIPER_DATABASE_HOST=localhost       # Хост БД")
	fmt.Println("    VIPER_DATABASE_PASSWORD=secret      # Пароль БД")
	fmt.Println("    VIPER_APP_ENVIRONMENT=production    # Окружение")
	fmt.Println("    VIPER_APP_DEBUG=false               # Режим отладки")
	fmt.Println()

	fmt.Println("ПРИОРИТЕТ НАСТРОЕК (от высшего к низшему):")
	fmt.Println("  1. Флаги командной строки")
	fmt.Println("  2. Переменные окружения")
	fmt.Println("  3. Конфигурационный файл")
	fmt.Println("  4. Значения по умолчанию")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Printf("  %s                                    # Запуск с настройками по умолчанию\n", appName)
	fmt.Printf("  %s --log-level=debug                  # Переопределить уровень логирования\n", appName)
	fmt.Printf("  %s --show-env                         # Показать переменные окружения\n", appName)
	fmt.Printf("  %s --show-config                      # Показать текущую конфигурацию\n", appName)
	fmt.Printf("  VIPER_LOG_LEVEL=trace %s              # Установить переменную окружения\n", appName)
	fmt.Printf("  %s --config=config.yaml               # Использовать конфигурационный файл\n", appName)
}

func initConfig() error {
	// Устанавливаем значения по умолчанию
	setDefaults()

	// Настраиваем Viper для работы с переменными окружения
	viper.SetEnvPrefix("VIPER")                // Префикс для переменных окружения
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()                       // Автоматически читаем переменные окружения

	// Читаем конфигурационный файл если указан
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		log.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Config file loaded")
	} else {
		// Ищем конфигурационный файл в стандартных местах
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.viper-env-demo")
		viper.AddConfigPath("/etc/viper-env-demo")

		if err := viper.ReadInConfig(); err == nil {
			log.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Config file found and loaded")
		}
	}

	// Переопределяем настройки флагами командной строки
	if *logLevel != "" {
		viper.Set("log.level", *logLevel)
	}
	if *verbose {
		viper.Set("log.level", "debug")
	}
	if *debug {
		viper.Set("app.debug", true)
		viper.Set("log.level", "debug")
	}

	// Загружаем конфигурацию в структуру
	config = &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

func setDefaults() {
	// Настройки логирования
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.caller", false)
	viper.SetDefault("log.timestamp", true)
	viper.SetDefault("log.pretty", false)
	viper.SetDefault("log.no_color", false)

	// Настройки сервера
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.tls.enabled", false)
	viper.SetDefault("server.tls.cert_file", "")
	viper.SetDefault("server.tls.key_file", "")

	// Настройки базы данных
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "viper_demo")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.ssl_mode", "prefer")

	// Настройки приложения
	viper.SetDefault("app.name", appName)
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)
	viper.SetDefault("app.features", []string{"logging", "metrics"})
}

func setupLogging() error {
	// Парсим уровень логирования
	level, err := zerolog.ParseLevel(config.Log.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)

	// Настраиваем вывод
	var output *os.File
	switch config.Log.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		file, err := os.OpenFile(config.Log.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Настраиваем формат
	var logger zerolog.Logger
	if config.Log.Format == "json" {
		logger = zerolog.New(output)
	} else {
		writer := zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "15:04:05",
			NoColor:    config.Log.NoColor,
		}

		if config.Log.Pretty && !config.Log.NoColor {
			writer.FormatLevel = func(i interface{}) string {
				level := strings.ToUpper(fmt.Sprintf("%s", i))
				switch level {
				case "TRACE":
					return fmt.Sprintf("\033[90m[%s]\033[0m", level)
				case "DEBUG":
					return fmt.Sprintf("\033[36m[%s]\033[0m", level)
				case "INFO":
					return fmt.Sprintf("\033[32m[%s]\033[0m", level)
				case "WARN":
					return fmt.Sprintf("\033[33m[%s]\033[0m", level)
				case "ERROR":
					return fmt.Sprintf("\033[31m[%s]\033[0m", level)
				default:
					return fmt.Sprintf("[%s]", level)
				}
			}
		}

		logger = zerolog.New(writer)
	}

	// Добавляем контекст
	if config.Log.Timestamp {
		logger = logger.With().Timestamp().Logger()
	}

	if config.Log.Caller {
		logger = logger.With().Caller().Logger()
	}

	// Добавляем информацию о приложении
	logger = logger.With().
		Str("app", config.App.Name).
		Str("version", config.App.Version).
		Str("environment", config.App.Environment).
		Logger()

	log.Logger = logger
	return nil
}

func showEnvironmentVariables() {
	fmt.Println("🌍 Переменные окружения с префиксом VIPER_:")
	fmt.Println()

	envVars := []string{
		"VIPER_LOG_LEVEL",
		"VIPER_LOG_FORMAT",
		"VIPER_LOG_OUTPUT",
		"VIPER_LOG_CALLER",
		"VIPER_LOG_TIMESTAMP",
		"VIPER_LOG_PRETTY",
		"VIPER_LOG_NO_COLOR",
		"VIPER_SERVER_HOST",
		"VIPER_SERVER_PORT",
		"VIPER_SERVER_READ_TIMEOUT",
		"VIPER_SERVER_WRITE_TIMEOUT",
		"VIPER_SERVER_TLS_ENABLED",
		"VIPER_DATABASE_HOST",
		"VIPER_DATABASE_PORT",
		"VIPER_DATABASE_DATABASE",
		"VIPER_DATABASE_USERNAME",
		"VIPER_DATABASE_PASSWORD",
		"VIPER_DATABASE_MAX_CONNECTIONS",
		"VIPER_APP_NAME",
		"VIPER_APP_VERSION",
		"VIPER_APP_ENVIRONMENT",
		"VIPER_APP_DEBUG",
	}

	found := false
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			fmt.Printf("  ✅ %s=%s\n", envVar, value)
			found = true
		} else {
			fmt.Printf("  ❌ %s (не установлена)\n", envVar)
		}
	}

	if !found {
		fmt.Println("  Переменные окружения не найдены")
	}

	fmt.Println()
	fmt.Println("💡 Примеры установки переменных:")
	fmt.Println("  export VIPER_LOG_LEVEL=debug")
	fmt.Println("  export VIPER_SERVER_PORT=3000")
	fmt.Println("  export VIPER_DATABASE_PASSWORD=mysecretpassword")
	fmt.Println("  export VIPER_APP_ENVIRONMENT=production")
}

func showCurrentConfig() {
	fmt.Println("⚙️  Текущая конфигурация:")
	fmt.Println()

	fmt.Printf("📋 Логирование:\n")
	fmt.Printf("  Уровень:      %s\n", config.Log.Level)
	fmt.Printf("  Формат:       %s\n", config.Log.Format)
	fmt.Printf("  Вывод:        %s\n", config.Log.Output)
	fmt.Printf("  Caller:       %t\n", config.Log.Caller)
	fmt.Printf("  Timestamp:    %t\n", config.Log.Timestamp)
	fmt.Printf("  Pretty:       %t\n", config.Log.Pretty)
	fmt.Printf("  No Color:     %t\n", config.Log.NoColor)
	fmt.Println()

	fmt.Printf("🌐 Сервер:\n")
	fmt.Printf("  Хост:         %s\n", config.Server.Host)
	fmt.Printf("  Порт:         %d\n", config.Server.Port)
	fmt.Printf("  Read Timeout: %v\n", config.Server.ReadTimeout)
	fmt.Printf("  Write Timeout:%v\n", config.Server.WriteTimeout)
	fmt.Printf("  TLS:          %t\n", config.Server.TLS.Enabled)
	fmt.Println()

	fmt.Printf("🗄️  База данных:\n")
	fmt.Printf("  Драйвер:      %s\n", config.Database.Driver)
	fmt.Printf("  Хост:         %s\n", config.Database.Host)
	fmt.Printf("  Порт:         %d\n", config.Database.Port)
	fmt.Printf("  База:         %s\n", config.Database.Database)
	fmt.Printf("  Пользователь: %s\n", config.Database.Username)
	fmt.Printf("  Пароль:       %s\n", maskPassword(config.Database.Password))
	fmt.Printf("  Макс. соед.:  %d\n", config.Database.MaxConns)
	fmt.Printf("  SSL Mode:     %s\n", config.Database.SSLMode)
	fmt.Println()

	fmt.Printf("🚀 Приложение:\n")
	fmt.Printf("  Имя:          %s\n", config.App.Name)
	fmt.Printf("  Версия:       %s\n", config.App.Version)
	fmt.Printf("  Окружение:    %s\n", config.App.Environment)
	fmt.Printf("  Debug:        %t\n", config.App.Debug)
	fmt.Printf("  Функции:      %v\n", config.App.Features)
}

func maskPassword(password string) string {
	if password == "" {
		return "(не установлен)"
	}
	if len(password) <= 3 {
		return "***"
	}
	return password[:2] + strings.Repeat("*", len(password)-2)
}

func runDemo() {
	log.Info().
		Str("config_source", getConfigSource()).
		Msg("🚀 Запуск демонстрации Viper + Environment Variables")

	// Демонстрируем приоритет настроек
	demonstrateConfigPriority()

	// Демонстрируем работу с переменными окружения
	demonstrateEnvironmentVariables()

	// Демонстрируем динамическое обновление конфигурации
	demonstrateDynamicConfig()

	// Демонстрируем использование конфигурации в коде
	demonstrateConfigUsage()

	log.Info().Msg("✅ Демонстрация завершена")
}

func getConfigSource() string {
	if viper.ConfigFileUsed() != "" {
		return "config_file"
	}
	return "defaults_and_env"
}

func demonstrateConfigPriority() {
	log.Info().Msg("📋 Демонстрация приоритета настроек:")

	// Показываем источники для разных настроек
	log.Info().
		Str("setting", "log.level").
		Str("value", config.Log.Level).
		Str("source", getSettingSource("log.level")).
		Msg("Источник настройки уровня логирования")

	log.Info().
		Str("setting", "server.port").
		Int("value", config.Server.Port).
		Str("source", getSettingSource("server.port")).
		Msg("Источник настройки порта сервера")

	log.Info().
		Str("setting", "app.environment").
		Str("value", config.App.Environment).
		Str("source", getSettingSource("app.environment")).
		Msg("Источник настройки окружения")
}

func getSettingSource(key string) string {
	if viper.InConfig(key) {
		return "config_file"
	}

	envKey := "VIPER_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if os.Getenv(envKey) != "" {
		return "environment"
	}

	return "default"
}

func demonstrateEnvironmentVariables() {
	log.Info().Msg("🌍 Демонстрация переменных окружения:")

	// Проверяем наличие переменных окружения
	envVars := map[string]string{
		"VIPER_LOG_LEVEL":          config.Log.Level,
		"VIPER_SERVER_HOST":        config.Server.Host,
		"VIPER_DATABASE_PASSWORD":  maskPassword(config.Database.Password),
		"VIPER_APP_ENVIRONMENT":    config.App.Environment,
	}

	for envVar, value := range envVars {
		if os.Getenv(envVar) != "" {
			log.Info().
				Str("env_var", envVar).
				Str("value", value).
				Bool("from_env", true).
				Msg("Переменная окружения установлена")
		} else {
			log.Debug().
				Str("env_var", envVar).
				Str("value", value).
				Bool("from_env", false).
				Msg("Переменная окружения не установлена, используется значение по умолчанию")
		}
	}
}

func demonstrateDynamicConfig() {
	log.Info().Msg("🔄 Демонстрация динамической конфигурации:")

	// Показываем как можно изменять конфигурацию во время выполнения
	originalLevel := viper.GetString("log.level")

	log.Info().
		Str("original_level", originalLevel).
		Msg("Исходный уровень логирования")

	// Временно изменяем уровень
	viper.Set("log.level", "debug")
	newLevel := viper.GetString("log.level")

	log.Info().
		Str("new_level", newLevel).
		Msg("Уровень логирования изменен динамически")

	// Демонстрируем debug лог
	log.Debug().Msg("Это debug сообщение теперь видно!")

	// Возвращаем обратно
	viper.Set("log.level", originalLevel)

	log.Info().
		Str("restored_level", originalLevel).
		Msg("Уровень логирования восстановлен")
}

func demonstrateConfigUsage() {
	log.Info().Msg("💼 Демонстрация использования конфигурации:")

	// Имитируем запуск сервера
	log.Info().
		Str("action", "server_start").
		Str("host", config.Server.Host).
		Int("port", config.Server.Port).
		Bool("tls_enabled", config.Server.TLS.Enabled).
		Msg("Запуск HTTP сервера")

	// Имитируем подключение к базе данных
	log.Info().
		Str("action", "database_connect").
		Str("driver", config.Database.Driver).
		Str("host", config.Database.Host).
		Int("port", config.Database.Port).
		Str("database", config.Database.Database).
		Str("username", config.Database.Username).
		Int("max_connections", config.Database.MaxConns).
		Msg("Подключение к базе данных")

	// Показываем настройки приложения
	if config.App.Debug {
		log.Debug().
			Str("mode", "debug").
			Strs("enabled_features", config.App.Features).
			Msg("Приложение запущено в режиме отладки")
	}

	// Имитируем различное поведение в зависимости от окружения
	switch config.App.Environment {
	case "development":
		log.Info().Msg("Режим разработки: включены дополнительные логи")
	case "staging":
		log.Info().Msg("Staging окружение: включена частичная обфускация данных")
	case "production":
		log.Info().Msg("Production окружение: включены метрики и мониторинг")
	}
}
