# Order Reader

Микросервис для обработки и отображения данных о заказах, получаемых из Kafka и сохраняемых в PostgreSQL

## 🚀 Функциональность

- Получение данных из Kafka топика `orders-topic`
- Сохранение в PostgreSQL с правильной структурой данных
- Кэширование в памяти для быстрого доступа
- REST API для получения данных по `order_uid`
- Веб-интерфейс для просмотра информации о заказах

## Архитектура

- Kafka (topic: orders-topic) — источник входящих сообщений о заказах
- Go Service (HTTP 8081)
  - Консюмер читает сообщения из Kafka
  - Сохраняет данные в PostgreSQL
  - Держит кэш в памяти для быстрых ответов
  - Предоставляет REST API и веб-интерфейс
- PostgreSQL — хранение заказов и связанных сущностей (delivery, payment, items)
- Web UI — простая фронтенд-страница, обращается к Go Service по HTTP

*Порядок потоков данных*

1. Kafka → Go Service (читатель топика orders-topic)  
2. Go Service → PostgreSQL (сохранение/обновление записей)  
3. Go Service ⇄ Внутренний кэш (быстрый доступ к недавно запрошенным order_uid)  
4. Клиент (браузер или API клиент) → Go Service (HTTP) → возвращается информация из кэша или БД

## 📦 Зависимости

- Docker & Docker Compose  
- Go 1.21+  
- PostgreSQL 15  
- Kafka

## 🛠 Установка и запуск

### 1 Клонирование и настройка

git clone <your-repo-url>  
cd readermicroservice

### 2 Запуск через Docker Compose

docker-compose up -d

Сервисы будут доступны по адресам  
- API сервис http://localhost:8081  
- Kafka UI http://localhost:8080  
- PostgreSQL localhost:5432

## 📡 API Endpoints

### GET /order/{order_uid}

Получение информации о заказе  
curl http://localhost:8081/order/b563feb7b2b84b6test

Пример ответа  
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}

## 🎯 Использование веб интерфейса

1 Откройте index.html в браузере  
2 Введите order_uid в поле ввода  
3 Нажмите "Get Order"  
4 Информация о заказе отобразится ниже

## 📊 Отправка тестовых данных

### Через Producer в контейнере

# Зайдите в контейнер  
docker exec -it reader-app /bin/sh

# Запустите producer  
cd /app  
go run producer.go

### Через curl

curl -X POST http://localhost:8081/test-order

## ⚙️ Конфигурация

Переменные окружения

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| DB_HOST | db | Хост PostgreSQL |
| DB_PORT | 5432 | Порт PostgreSQL |
| DB_USER | myuser | Пользователь БД |
| DB_PASSWORD | mypassword | Пароль БД |
| DB_NAME | mydatabase | Имя базы данных |
| KAFKA_BROKERS | kafka1:29092 | Адреса Kafka брокеров |

## 🗄 Структура базы данных

Сервис автоматически создает таблицы  
- orders — основная информация о заказах  
- delivery — данные доставки  
- payments — информация об оплате  
- items — товары в заказе

## 🔧 Разработка

### Локальный запуск без Docker

# Установите зависимости  
go mod tidy

# Запустите сервис  
go run cmd/main.go
