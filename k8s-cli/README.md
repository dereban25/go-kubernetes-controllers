# k8s-cli - Simple Kubernetes CLI Tool

Простой инструмент командной строки для работы с Kubernetes кластерами, построенный с использованием Cobra CLI.

## 🎯 Возможности

- **Управление контекстами**: переключение между Kubernetes контекстами
- **Просмотр ресурсов**: список подов, деплойментов, сервисов и namespace'ов
- **Создание ресурсов**: применение YAML файлов для создания ресурсов
- **Гибкий вывод**: поддержка форматов table, JSON и YAML

## 🚀 Быстрый старт

### Установка

```bash
# Инициализация проекта (уже выполнено)
go mod init k8s-cli

# Установка зависимостей
make deps

# Сборка приложения
make build

# Установка в систему (опционально)
make install
```

### Использование

```bash
# Показать справку
./bin/k8s-cli --help

# Список контекстов
./bin/k8s-cli context list

# Переключить контекст
./bin/k8s-cli context set my-cluster

# Список подов
./bin/k8s-cli list pods

# Список подов в определенном namespace
./bin/k8s-cli list pods -n kube-system

# Создать ресурс из YAML
./bin/k8s-cli apply file examples/pod.yaml
```

## 📋 Команды

### Управление контекстами
- `k8s-cli context list` - показать все контексты
- `k8s-cli context current` - показать текущий контекст
- `k8s-cli context set <name>` - переключить контекст

### Просмотр ресурсов
- `k8s-cli list pods` - список подов
- `k8s-cli list deployments` - список деплойментов
- `k8s-cli list services` - список сервисов
- `k8s-cli list namespaces` - список namespace'ов

### Создание ресурсов
- `k8s-cli apply file <filename>` - создать ресурсы из YAML файла

## 🛠 Разработка

```bash
# Установить зависимости
make deps

# Запустить тесты
make test

# Форматировать код
make fmt

# Запустить демонстрацию
make demo
```

## 📁 Структура проекта

```
k8s-cli/
├── cmd/                    # CLI команды
├── internal/               # Внутренние пакеты
│   ├── k8s/               # Kubernetes клиент
│   └── utils/             # Утилиты
├── examples/              # Примеры YAML файлов
├── main.go               # Точка входа
├── go.mod                # Go модуль
└── Makefile              # Автоматизация сборки
```

## 🔧 Требования

- Go 1.21 или выше
- kubectl настроенный для доступа к кластеру
- Доступ к Kubernetes кластеру

## 📖 Примеры

```bash
# Просмотр всех контекстов
$ k8s-cli context list
Доступные контексты:
  minikube
* kind-cluster (текущий)
  production

# Список подов в JSON формате
$ k8s-cli list pods -o json

# Создание пода из примера
$ k8s-cli apply file examples/pod.yaml
✅ Ресурсы успешно созданы из файла: examples/pod.yaml
```
