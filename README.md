# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


Трек «Сервис сбора метрик и алертинга»
  
1. Скомпилируйте ваши сервер и агент в папках cmd/server и cmd/agent командами go build -o server *.go и go build -o agent *.go соответственно.
2. Скачайте бинарный файл с автотестами для вашей ОС — например, metricstest-darwin-arm64 для MacOS на процессоре Apple Silicon.
3. Разместите бинарный файл так, чтобы он был доступен для запуска из командной строки, — пропишите путь в переменную $PATH.
4. Ознакомьтесь с параметрами запуска автотестов в файле .github/workflows/metricstest.yml вашего репозитория. Автотесты для разных инкрементов требуют различных аргументов для запуска.

go build -o cmd/server/server cmd/server/*.go

metricstest_v2 -test.v -test.run=^TestIteration1$ -agent-binary-path=cmd/agent/agent

go build -o cmd/server/server cmd/server/*.go && metricstest_v2 -test.v -test.run=^TestIteration1$ -agent-binary-path=cmd/agent/agent

metricstest_v2 -test.v -test.run=^TestIteration1$ -binary-path=./cmd/server/server

Iter2
go build -o cmd/server/server cmd/server/*.go && go build -o cmd/agent/agent cmd/agent/*.go
metricstest_v2 -test.v -test.run=^TestIteration2A$ -binary-path=./cmd/server/server -agent-binary-path=cmd/agent/agent
metricstest_v2 -test.v -test.run=^TestIteration2B$ -binary-path=./cmd/server/server -agent-binary-path=cmd/agent/agent -source-path=.
