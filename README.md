# Task Service

REST API сервис для управления задачами медицинской системы.

## Основной функционал

Сервис поддерживает:

- создание задачи
- получение списка задач
- получение задачи по ID
- обновление задачи
- удаление задачи
- создание задач с периодичностью (recurrence)

---

## Архитектура

- `domain` — модель задачи и базовые типы
- `usecase` — бизнес-логика и валидация
- `repository` — работа с PostgreSQL
- `transport/http` — HTTP API (handlers, DTO)

Новая функциональность встроена в существующую архитектуру без её изменения.

---

## Логика создания задач

### Обычная задача

Если передан `scheduled_for`:

→ создается **одна задача** с указанной датой

Если `scheduled_for` не передан:

→ используется текущая дата (UTC, без времени)

---

### Задача с recurrence

Если передан `recurrence`:

→ система:
1. валидирует входные данные
2. вычисляет список дат
3. создает **несколько обычных задач** (по одной на каждую дату)

⚠️ В БД не хранится "серия" задач — только обычные задачи

---

## Формат recurrence

```json
{
  "type": "daily | monthly | specific_dates | odd_days | even_days",
  "every_n_days": 2,
  "day_of_month": 15,
  "start_date": "2026-04-01",
  "end_date": "2026-04-30",
  "dates": ["2026-04-02", "2026-04-10"]
}
```

---

## Примеры тестовых запросов к сервису

- Обычная задача POST http://localhost:8080/api/v1/tasks 
``` json
{
  "title": "Single task",
  "description": "Manual test",
  "status": "new",
  "scheduled_for": "2026-04-10"
}
```
- Список задач GET http://localhost:8080/api/v1/tasks 
- Получить задачу по id GET http://localhost:8080/api/v1/tasks/1
- Обновление задачи GET http://localhost:8080/api/v1/tasks/1
``` json
{
  "title": "Updated task",
  "description": "Updated description",
  "status": "in_progress",
  "scheduled_for": "2026-04-11"
}
```
- Удаление DELETE http://localhost:8080/api/v1/tasks/1
- Поыторение каждые н дней POST http://localhost:8080/api/v1/tasks создаст 4 задачи 01, 03, 05, 07
```json
{
  "title": "Daily check",
  "description": "Ward A",
  "status": "new",
  "recurrence": {
    "type": "daily",
    "every_n_days": 2,
    "start_date": "2026-04-01",
    "end_date": "2026-04-07"
  }
}
```
- Повторение раз в месяц POST http://localhost:8080/api/v1/tasks создаст 3 задачи, 15 апреля, мая, июня
``` json
{
  "title": "Monthly report",
  "description": "Pharmacy",
  "status": "new",
  "recurrence": {
    "type": "monthly",
    "day_of_month": 15,
    "start_date": "2026-04-01",
    "end_date": "2026-06-30"
  }
}
```
- Несколько задач на конкрентые даты POST http://localhost:8080/api/v1/tasks создаст 2 задачи на 10 и 12 апреля, одна дата повторяется специально, для проверки валидации
``` json 
{
  "title": "ECG follow-up",
  "description": "Room 12",
  "status": "new",
  "recurrence": {
    "type": "specific_dates",
    "dates": [
      "2026-04-10",
      "2026-04-10",
      "2026-04-12"
    ]
  }
}
```
- Создать повторяющиеся задачи на каждый нечетный день месяца с начального дня до конечного POST http://localhost:8080/api/v1/tasks создаст 3 задачи 01, 03, 05
``` json 
{
  "title": "Odd days task",
  "description": "Only odd days",
  "status": "new",
  "recurrence": {
    "type": "odd_days",
    "start_date": "2026-04-01",
    "end_date": "2026-04-05"
  }
}
``` 
- Создать повторяющиеся задачи на каждый четный день месяца с начального дня до конечного POST http://localhost:8080/api/v1/tasks создаст 3 задачи 02, 04, 06
``` json
{
  "title": "Even days task",
  "description": "Only even days",
  "status": "new",
  "recurrence": {
    "type": "even_days",
    "start_date": "2026-04-01",
    "end_date": "2026-04-06"
  }
}
```


