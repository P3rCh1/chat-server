# chat-server
## В репозитории представлен мой pet-проект — backend для мессенджера на языке Go
### Используемые технологии:

```
- net/http
- chi, chi/middleware
- gorilla/websocket,
- gRPC,
- Kafka
- PostgreSQL
- Redis
- Docker
```

### Запуск
```bash
git clone https://github.com/P3rCh1/chat-server.git
cd chat-server
docker compose up --build -d
```
- Для запуска необходимо объявить следующие переменные окружения:  
```
# PostgreSQL
POSTGRES_USER=chat                      # Пользователь БД
POSTGRES_PASSWORD=your_strong_pass_1    # Пароль
POSTGRES_DB=chatdb                      # Название БД

# Redis
REDIS_PASSWORD=your_strong_pass_2

# Пути к конфигураиям (по умолчанию находятся в корневых дирректориях микросервисов)
GATEWAY_CONFIG_PATH=./config.yaml
SESSION_CONFIG_PATH=./config.yaml
USER_CONFIG_PATH=./config.yaml
ROOMS_CONFIG_PATH=./config.yaml
MESSAGE_CONFIG_PATH=./config.yaml

# JWT
JWT_SECRET=your_jwt_secret               # Секретный ключ для токенов
```
- Так же перед запуском можно настроить config.yaml для каждого сервиса
  
### API Endpoints

- Аутентификация

1) POST	/register  
Регистрация нового пользователя  
Имя и email должны быть уникальными  
Пример:
```
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d  '{
        "Username": "user",
        "Email": "user@example.com",
        "Password": "password"
    }'
```

2) PUT 	/login  
Получение JWT токена  
Пример:
```
curl -X PUT http://localhost:8080/login \
-H "Content-Type: application/json" \
-d  '{
        "Email": "user@example.com",
        "Password": "password"
    }'
```

- Пользователи  

1) GET	/profile  
Посмотреть свой профиль  
Пример:
```
curl -X GET http://localhost:8080/profile \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

2) GET	/profile/{UID}  
Посмотреть профиль по UID 
Пример:
```
curl -X GET http://localhost:8080/profile/1
```

3) PUT	/change-name  
Изменить имя пользователя  
Пример:  
```
curl -X PUT http://localhost:8080/change-name \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
        "NewName": "newname1234"
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
        "Name": "GeneralChat",
        "IsPrivate": false
    }'
```

2) PUT /join  
Присоединиться к комнате  
Если is_private=false: пользователи могут присоединиться к комнате  
Пример:
```
curl -X PUT http://localhost:8080/join \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
        "RoomID": 1
    }'
```

3) PUT /invite  
Пригласить пользователя  
Создатель может добавлять пользователей в комнаты  
Пример:
```
curl -X PUT http://localhost:8080/invite \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-H "Content-Type: application/json" \
-d  '{
        "RoomID": 1,
        "UID": 1
    }'
```

4) GET /rooms  
Получить слайс ID комнат пользователя  
Пример:
```
curl -X GET http://localhost:8080/rooms \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

4) GET /room/{roomID}  
Получить информацию о комнате  
Пример:
```
curl -X GET http://localhost:8080/room/1
```  

- Messages  
1) GET /messages/{roomID}  
Получить слайс сообщений, отправленных в комнате, отсортированных по времени добавления в обратном порядке  
Eсли LastID = 0, отправляются сообщения начаиная с последнего 
Лимит отправки - 100 сообщений
Пример:
```
curl -X GET http://localhost:8080/messages/1 \
-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
-d '{"LastID":100}'
```

- Websocket  
1) /ws  
Подключиться по websocket  
Пример:
```
wscat -c "ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```  
- Войти в комнату
Если RoomID = 0 - выходишь из комнаты
```
{"Type":"enter", "RoomID":1}
```  
- Написать сообщение
```
{"Type":"message","Text":"my message"}
```
  
### Архитектура проекта
- За получение запросов и удержание вебсокет соединения отвечает "gateway-service", 
с другими сервисами он общается с помощью gRPC  
- "session-service" отвечает за валидацию и создание jwt токенов  
- "user-service" отвечает за запросы на создание изменение и получение профилей пользователей  
- "rooms-service" аналогично отвечает за запросы, связанных с комнатами. Кроме того отправляет сообщение об успешном добавлении пользователя в нужную комнату  
- "message-service" принимает запросы на отправку сообщения для их сохранения в базу и дальнейшей отправки в комнаты с помощью Kafka  
<br>

- Кроме того, настроено кэширование в Redis для профилей пользователей, списка их комнат, профилей комнат
