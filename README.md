# test-backend-engory03

[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/uvnTmvcw)

Это репозиторий для выполнения тестового задания на стажировку в Авито.

Инструкция - в [INSTRUCTIONS.md](INSTRUCTIONS.md).

Подробное условие задания - в [TASK.md](TASK.md).

Описание API — в [api.yaml](api.yaml).

Результаты нагрузочного тестирования - в [loadtest/RESULTS.md](loadtest/RESULTS.md).

Конфигурация линтера - в [.golangci.yaml](.golangci.yaml).

## Инструкция

Запустить сервис в Docker (работает без .env файла):

```sh
# linux
make up

# windows
.\tasks.ps1 up

# или просто
docker compose up -d --build
```

> Swagger API доступен по адресу: http://localhost:8080/docs/index.html
>
> Для защищённых эндпоинтов нажмите кнопку Authorize. В открывшемся окне введите `Bearer <токен>`.

Остановить сервис, запущенный в Docker:

```sh
# linux
make down

# windows
.\tasks.ps1 down

# или просто
docker compose down

# если нужно очистить БД (volume)
docker compose down -v
```

Заполнить БД тестовыми данными (seed) в Docker:

```sh
# linux
make seed

# windows
.\tasks.ps1 seed
```

Прогнать unit-тесты:

```sh
# linux
make test

# windows
.\tasks.ps1 test
```

Прогнать unit-тесты cо статистикой покрытия:

```sh
# linux
make test-coverage

# windows
.\tasks.ps1 test-coverage
```

Прогнать E2E-тесты с testcontainers для тестовой БД PostgreSQL:

```sh
# linux
make test-e2e

# windows
.\tasks.ps1 test-e2e
```

Провести нагрузочное тестирование (load testing):

```sh
# linux
make loadtest

# windows
.\tasks.ps1 loadtest
```

> Перед нагрузочным тестированием сначала необходимо запустить команды:
>
> ```sh
> # linux
> make up
> make seed
>
> # windows
> .\tasks.ps1 up
> .\tasks.ps1 seed
> ```

Прогнать линтер:

```sh
# linux
make lint

# windows
.\tasks.ps1 lint
```

## Conference Service (мок) и сбои при создании брони

Для опционального поля `createConferenceLink: true` в `POST /bookings/create` используется **мок** внешнего сервиса (`internal/bookings/conference`), без реального HTTP.

**Принятые решения при ошибках:**

1. **Мок недоступен / вернул ошибку** — бронь **не создаётся**. Клиент получает **500** с кодом `INTERNAL_ERROR` (в Swagger для этого сценария предусмотрен общий ответ 500). Ссылка на конференцию в БД не появляется, состояние данных согласовано.
2. **Мок успешен, но запись брони в БД не удалась** (редкий сбой инфраструктуры) — в реальной системе могла бы остаться «висячая» конференция у провайдера; в моке побочных эффектов нет.
3. **Гонка двух запросов на один слот** — проверка «нет активной брони» выполняется **внутри транзакции** перед `INSERT`; второй запрос получает **409** `SLOT_ALREADY_BOOKED` с сообщением `slot is already booked`, как в `api.yaml`.

**Имитация сбоя замоканного сервиса:** в `.env` задайте `CONFERENCE_MOCK_FAIL=1` или `true` — любой вызов создания ссылки будет завершаться ошибкой (удобно для ручных проверок).
