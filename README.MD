## chat-server
## В репозитории представлен мой pet-проект для backend-разработки на Go
**Используемые технологии:** 

```
- net/http
- chi, chi/middleware
- gorilla/websocket
- postgres
- redis
```

**Запуск**
```bash
git clone https://github.com/P3rCh1/chat-server.git
cd chat-server
export CONFIG_PATH=./configs/config_example.yaml ##необходимо изменить файл под систему
go run cmd/chat-server/main.go
```

**API Endpoints**

- Аутентификация

1) POST	/register  
Регистрация нового пользователя  
Имя и email должны быть уникальными  
Пример:
```
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d  '{
        "username": "user",
        "email": "user@example.com",
        "password": "password"
    }'
```

2) PUT 	/login  
Получение JWT токена  
Пример:
```
curl -X PUT http://localhost:8080/login \
-H "Content-Type: application/json" \
-d  '{
        "email": "user@example.com",
        "password": "password"
    }'
```

- Пользователи  
1) GET	/profile  
Посмотреть профиль  
Пример:
```
curl -X GET http://localhost:8080/profile \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

2) PUT	/change-name  
Изменить имя пользователя  
Пример:  
```
curl -X PUT http://localhost:8080/change-name \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
        "username": "newname1234"
    }'
```

- Комнаты  
1) POST	/create-room  
Создать новую комнату  
Название должно быть уникальным  
Пример:
```
curl -X POST http://localhost:8080/create-room \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
    "name": "GeneralChat",
    "is_private": false
    }'
```

2) PUT	    /join  
Присоединиться к комнате  
Если is_private=false: пользователи могут присоединиться к комнате  
Пример:
```
curl -X PUT http://localhost:8080/join \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
    "room_id": 1
    }'
```

3) PUT	    /invite  
Пригласить пользователя  
Создатель может добавлять пользователей в комнаты  
Пример:
```
curl -X PUT http://localhost:8080/invite \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
    "room_id": 1,
    "user_id": 1
    }'
```

- Websocket  
1) /ws  
Подключение websocket  
Пример:
```
wscat -c "ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

- Сообщения отправляются следующим образом:
```
{"type":"your_type", "content":"your_content"}
```
- Типы сообщений:
1) "join" присоединиться к комнате  
Можно присоединиться только если до этого получил доступ  
Если пользователь уже состоит в комнате - покинет ее и успешно присоединится к новой
2) "leave" покинуть комнату  
3) "message" написать сообщение в комнату

- Все сообщения хранятся в базе данных и в дальнейшем возможно добавить к ним доступ  
- При присоединени, покидании комнаты, всем ее участникам рассылается сообщение от лица system_username (настраивается в конфиге)  
- На все ошибочные запросы отправляет ответ error_username (настраивается в конфиге)