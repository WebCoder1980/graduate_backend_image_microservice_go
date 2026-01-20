# Высокопроизводительное приложение для пакетной обработки изображений

![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white&style=for-the-badge)
![MinIO](https://img.shields.io/badge/MinIO-FF6F00?logo=minio&logoColor=white&style=for-the-badge)
![Kafka](https://img.shields.io/badge/Apache_Kafka-231F20?logo=apachekafka&logoColor=white&style=for-the-badge)
![Keycloak](https://img.shields.io/badge/Keycloak-F54B42?logo=keycloak&logoColor=white&style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=for-the-badge)

Проект представляет собой набор микросервисов для пакетной обработки изображений с использованием Golang, S3 MinIO хранилища, PostgreSQl, Apache Kafka, Keycloak. Сервис позволяет загружать изображения по файлам или URL, обрабатывать их асинхронно и хранить в CDN. Предназначен для использования интернет-магазинами, досками объявлений и другими платформами, где требуется массовая обработка изображений.

Это проект для диплома в моём учебном заведении профессионального образования.

Проект содержит почти весь функционал, на данный момент ведутся работы по оптимизации, "полировке", повышение безопастности, упрощения деплоя (посредством скриптов и Docker).

# Особенности

- Поддержка одиночной и пакетной обработки изображений.
- Асинхронная обработка с прогресс-трекингом через task_id.
- Поддержка следующих параметров обработки:
	- width и height — ресайз с сохранением пропорций или принудительным растяжением.
	- format — выходной формат: jpg, png, webp.
	- quality — качество для lossy-форматов.
- Загрузка как с локальных файлов, так и по URL.
- Хранение изображений в S3-совместимом хранилище (MinIO).
- Веб-интерфейс для загрузки, отслеживания статуса и получения CDN-ссылок.

# Endpoints

## Создать задачу

POST `/api/v1/task`
Body (form-data):

- file - файлы, которые нужно обработать. Если файлов несколько - можно повторять ключ `file`.
- width (optional) - число пикселей ширины в фото после обработки.
- height (optional) - число пикселей высоты в фото после обработки.
- format (optional) - jpg", "png" или "webp" - формат фото после обработки.
- quality (optional) - качество lossy-форматов в фото после обработки. Если фото не поддерживает "качество", то свойство будет проигноровано.

Response: task_id, по которому можно отслеживать статус задачи и собирать результат.

## Отслеживание статуса задачи

GET `/api/v1/task/{task_id}`

Response. Пример Body:

```json
{
    "id": 1,
    "width": null,
    "height": null,
    "format": null,
    "quality": null,
    "common_status_id": 2,
    "images": [
        {
            "id": 1,
            "name": "Зима",
            "format": "jpg",
            "task_id": 1,
            "position": 1,
            "status_id": 2,
            "end_dt": "2026-01-08T23:44:31.009014Z"
        },
        {
            "id": 2,
            "name": "Снимок экрана (9)",
            "format": "png",
            "task_id": 1,
            "position": 2,
            "status_id": 2,
            "end_dt": "2026-01-08T23:44:32.410358Z"
        }
    ],
    "created_dt": "2026-01-08T23:44:25.694492Z"
}
```

## Получить фото после обработки

id фото можно получить в эндпоинте для отслеживания статуса заявки.

GET `/api/v1/image-processor/{image_id}`

Response. Body: фото после обработки. Пример:

<img width="1672" height="1058" alt="image" src="https://github.com/user-attachments/assets/c7345d9e-f66f-43da-b74f-341ca7b9c1b9" />

## Регистрация пользователя

Пользователи на данный момент не используются, кроме как в user microservice.

POST `localhost:37005/api/v1/user/register`

Request. Body - `username` и `password` нового пользователя в формате JSON. Пример:

```json
{
    "username": "maxsmg",
    "password": "qweqwe"
}
```

## Аунтификация. Получение access и refresh token.

POST `localhost:37005/api/v1/user/login`

Request. Body - `username` и `password` пользователя в формате JSON. Пример:

```json
{
    "username": "maxsmg",
    "password": "qweqwe"
}
```

Response. Body - токены и их информация в формате JSON. Пример:

```json
{
    "access_token": "xxxxx",
    "id_token": "xxxxx",
    "expires_in": 60,
    "refresh_expires_in": 1800,
    "refresh_token": "xxxxx",
    "token_type": "Bearer",
    "not-before-policy": 0,
    "session_state": "xxxxxxxx-7087-c594-ac5f-36aedf3fa760",
    "scope": "openid email profile"
}
```

## Обновить токены

POST `/api/v1/user/refresh-token`

Request. Body - единственное поле JSON с `refresh_token`, Пример:

```json
{
    "refresh_token": "xxxxx"
}
```

Response. Body - токены и их информация в формате JSON. Пример:

```json
{
    "access_token": "xxxxx",
    "id_token": "xxxxx",
    "expires_in": 60,
    "refresh_expires_in": 1800,
    "refresh_token": "xxxxx",
    "token_type": "Bearer",
    "not-before-policy": 0,
    "session_state": "xxxxxxxx-7087-c594-ac5f-36aedf3fa760",
    "scope": "openid email profile"
}
```

# Возможные переменные окружения

## Task microservice

```bash
handler_port=37002;kafka_address=localhost:9092;minio_access_key_id=minioadmin;minio_address=localhost:9000;minio_secret_access_key=minioadmin;minio_token=;postgresql_dbname=graduate_task;postgresql_host=localhost;postgresql_password=xxxxx;postgresql_port=37001;postgresql_user=postgres;minio_use_ssl=false
```

## Image processor microservice

```bash
handler_port=37003;kafka_address=localhost:9092;minio_access_key_id=minioadmin;minio_address=localhost:9000;minio_secret_access_key=minioadmin;minio_token=;postgresql_dbname=graduate_image_processor;postgresql_host=localhost;postgresql_password=xxxxx;postgresql_port=37001;postgresql_user=postgres;minio_use_ssl=false
```

## User microservice

```bash
handler_port=37007;keycloak_address=http://localhost:37006;keycloak_admin_password=admin;keycloak_admin_username=admin;keycloak_client_id=testc;keycloak_client_secret=xxxxx;keycloak_realm=master
```

## Gateway API microservice

```bash
handler_port=37005;image_processor_microservice_address=http://localhost:37003;task_microservice_address=http://localhost:37002
```

# Технологический стек

## Backend

- Язык: Go
- Очередь сообщений: Apache Kafka
- База данных: PostgreSQL
- Хранилище объектов: MinIO (S3-совместимое)
- Контейнеризация: Docker

## Frontend

- React.js (или Vue.js)
- Drag-and-drop загрузка файлов
- Поля ввода для URL и параметров обработки
- Панель мониторинга статуса задач
