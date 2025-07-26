# vk-segments-service
A service that stores user data and segments to which users belong on VK platforms

 Микросервис для управления пользовательскими сегментами. Позволяет:
✅ Создавать, удалять и просматривать сегменты
✅ Добавлять/удалять пользователей в сегменты
✅ Распределять сегменты на % пользователей

Технологии

Язык: Go
База данных: PostgreSQL
Фреймворк: Gin (HTTP-роутинг)

Запуск

1. Поднимаем PostgresSQL
```golang
docker-compose up -d  
```
