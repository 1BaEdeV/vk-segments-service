# vk-segments-service
A service that stores user data and segments to which users belong on VK platforms

__Позволяет:__  
✅ Создавать, удалять и просматривать сегменты  
✅ Добавлять/удалять пользователей в сегменты  
✅ Распределять сегменты на % пользователей

__Технологии__  
Язык: _Go_  
База данных: _PostgreSQL_  
Фреймворк: _Gin_ (HTTP-роутинг)

Запуск

__1. Поднять бд PostgresSQL:__
   ```bash
   docker-compose up -d  
   ```
__2. Запустить сервер:__
   ```bash
   go run cmd/app/main.go  
   ```

__API Endpoints__

- __POST__ /segments – создать сегмент

- __GET__ /segments – список всех сегментов

- __POST__ /users/:id/segments – добавить пользователя в сегмент

- __GET__ /users/:id/segments – сегменты пользователя

Пример запроса:
```bash
curl -X POST http://localhost:8080/segments/MAIL_GPT/distribute \
  -H "Content-Type: application/json" \
  -d '{"percent":30}'
```bash
