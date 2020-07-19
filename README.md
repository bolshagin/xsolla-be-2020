Тестовое задание в Xsolla BE School 2020
========================================
Разработан сервис с JSON API на стеке Golang, MySQL.

Используемые библиотеки:
- *BurntSushi/toml* - конфигурирование сервиса
- *gorilla/mux* - роутинг запросов
- *sirupsen/logrus* - логирование
- *stretchr/testify* - тестирование
- *dgrijalva/jwt-go* - генерация JWT-токенов для авторизации
- *google/uuid* - генерация UUID при создании платежной сессии
- *go-sql-driver/mysql* - драйвер для соединения с MySQL базой

**Небольшое примечание по тестированию**. Тестирование хранилища выполняется на тестовом хранилище *apipayment_test*, 
конфиг которого прописан в самом коде. Т.е. для того, чтобы корректно провести все тесты, 
необходимо в БД создать схему *apipayment_test* и таблицу *Sessions*.

Установка 
---------
1. Клонировать репозиторий проекта
   ```sh
   $ git clone https://github.com/bolshagin/xsolla-be-2020.git
   ```
2. Создать таблицу Sessions в базе данных
    ```sql
    CREATE TABLE `Sessions` (
        `SessionID` INT NOT NULL AUTO_INCREMENT,
        `SessionToken` NVARCHAR(4000) NOT NULL,
        `Amount` FLOAT NOT NULL,
        `Purpose` NVARCHAR(4000) NULL,
        `CreatedAt` DATETIME NOT NULL,
        `ClosedAt` DATETIME NOT NULL DEFAULT '1000-01-01 00:00:00',
        PRIMARY KEY (`SessionID`)
    );
    ```
3. Сконфигурировать .toml-конфиг в папке ./configs
   ```toml
   bind_addr = ":8080"
   log_level = "debug"
   
   [store]
   dbname = "apipayment_dev"
   user = "dev"
   password = "12345"
   ```
4. С помощью makefile построить проект
   ```sh
   $ make 
   ```
   или выполнить команду
   ```sh
   $ go build -v ./cmd/apiserver 
   ```
5. Запустить получившийся бинарник
   ```
   apiserver.exe
   ```
   или
   ```
   $ ./apiserver
   ```
6. Для запусков тестов в отдельном окне терминала (после запуска сервера) выполнить команду
   ```sh
   $ make test
   ```

Описание API
------------

### Создание платежной сессии
**/session**

`POST /session` - создает платежную сессию с переданными параметрами суммы платежа и назначания (длина ограничена 210 символами вместе с пробелами). 
Успешный ответ на запрос возвращает json со следующими полями:
* *session_token* (токен платежной сессии)
* *amount* (сумма платежа)
* *purpose* (назначение платежа) 
* *created_at* (дата создание платежной сессии)
* *closed_at* (дата закрытия платежной сессии)

Пример запроса:
```
curl --location --request POST 'http://localhost:8080/session' \
--header 'Content-Type: application/json' \
--data-raw '{
    "amount": 1000,
    "purpose": "услуги ЖКХ"
}
```
Ответ:
```json
{
    "session_token": "905dcda8-1c63-486c-bbd1-c7123e9c3e81",
    "amount": 1000,
    "purpose": "услуги ЖКХ",
    "created_at": "2020-07-19T07:28:14.1422484Z",
    "closed_at": "0001-01-01T00:00:00Z"
}
```
##### Коды ответов
* `201 Created` - платежная сессия создана
* `400 Bad request` - ошибка в формировании запроса или количество символов > 210
* `422 Unprocessable Entity` - ошибка возникшая при создании сессии в базе данных

### Обработка платежной сессии
**/pay**

`POST /pay` - обрабатывает платежную сессию с переданным токеном и параметрами (номер карты, CVC/CVV и дата). 
Время платежной сессии ограничено 15 минутами. При валидных параметрах, платежная сессия считается закрытой.
Валидация следующая: 
* номер карты проверяется по алгоритму Луна 
* в CVV/CVC поле можно передавать только числа (0-9) общей длиной 3 символа
* дата задается в следующем формате "M/YY".


Пример запроса:
```
curl --location --request POST 'http://localhost:8080/pay' \
--header 'Content-Type: application/json' \
--data-raw '{
    "session_token": "905dcda8-1c63-486c-bbd1-c7123e9c3e81",
    "card_number": "4111 1111 1111 1111",
    "code": "123",
    "date": "1/23"
}'
```
Ответ:
```json
{
    "payment": "successful"
}
```
##### Коды ответов
* `200 OK` - платежная сессия выполнена
* `400 Bad request` - ошибка в формировании запроса, либо ошибки связанные с неправильным форматом параметров платежа
* `500 Internal Server Error` - не найдена платежная сессия в БД, либо ошибки связанные с БД

### Получение JWT-токена
**/get-token**

`GET /get-token` - эндпойнт для получения JWT-токена. Длительность токена 1 час.
Пример запроса:
```
curl --location --request GET 'http://localhost:8080/get-token'
```
Ответ:
```json
{
    "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTUxNDU0MzYsInVzZXIiOiJEZWZhdWx0IFVzZXIifQ.WTCQjcc-Z6yyNe7Y0T2uMyUTfdLwtUN_ZIq5L8Llw7E"
}
```
##### Коды ответов
* `201 Created` - JWT-токен успешно создан
* `500 Internal Server Error` - ошибки связанные с генерацией JWT-токена

### Получение созданных сессий за определенный период
**/stat**

`GET /stat` - получение созданных платежных сессий за указанный период. 
Эндпойнт закрыт авторизацией по JWT-токену. Корректный формат даты YYYY-MM-DD.

Возвращает:
* список платежных сессий
    * *amount* (сумма платежа)
    * *purpose* (назначение платежа) 
    * *created_at* (дата создание платежной сессии)
    * *closed_at* (дата закрытия платежной сессии)

Пример запроса:
```
curl --location --request GET 'http://localhost:8080/stat' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTUxNDU0MzYsInVzZXIiOiJEZWZhdWx0IFVzZXIifQ.WTCQjcc-Z6yyNe7Y0T2uMyUTfdLwtUN_ZIq5L8Llw7E' \
--header 'Content-Type: application/json' \
--data-raw '{
    "date_begin": "2020-07-18",
    "date_end": "2020-07-20"
}'
```
Ответ:
```json
[
    {
        "amount": 1000,
        "purpose": "for testing",
        "created_at": "2020-07-19T06:56:53Z",
        "closed_at": "2020-07-19T06:57:03Z"
    },
    {
        "amount": 1000,
        "purpose": "for testing",
        "created_at": "2020-07-19T05:46:30Z",
        "closed_at": "2020-07-19T05:46:40Z"
    },
    {
        "amount": 1000,
        "purpose": "for testing",
        "created_at": "2020-07-18T15:21:33Z",
        "closed_at": "2020-07-18T15:21:44Z"
    },
    {
        "amount": 100,
        "purpose": "test",
        "created_at": "2020-07-18T15:11:33Z",
        "closed_at": "2020-07-18T15:11:33Z"
    },
    {
        "amount": 100,
        "purpose": "test",
        "created_at": "2020-07-18T14:46:39Z",
        "closed_at": "2020-07-18T14:46:39Z"
    },
    {
        "amount": 1000,
        "purpose": "for testing",
        "created_at": "2020-07-18T14:44:54Z",
        "closed_at": "2020-07-18T14:45:49Z"
    },
    {
        "amount": 1000,
        "purpose": "for testing",
        "created_at": "2020-07-18T14:41:28Z",
        "closed_at": "2020-07-18T14:44:30Z"
    },
    {
        "amount": 5000,
        "purpose": "for testing",
        "created_at": "2020-07-18T14:26:26Z",
        "closed_at": "2020-07-18T14:35:08Z"
    }
]
```
##### Коды ответов
* `200 OK` - данные успешно переданы
* `400 Bad request` - ошибка в формировании запроса, либо ошибки связанные с неправильным форматом даты
* `401 Unautorized` - ошибка при авторизации по переданному JWT-токену
* `500 Internal Server Error` - ошибки связанная с работой БД, либо записей за указанный период не найдено