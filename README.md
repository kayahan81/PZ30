---
# Практическое задание 30

## ЭФМО-02-25 

## Алиев Каяхан Командар оглы
---
# Тема работы
Реализация очереди задач (producer–consumer): retries, DLQ, идемпотентность

## Цели занятия
Построить рабочую очередь задач, которая устойчиво обрабатывает ошибки: временные ошибки ретраятся, “плохие” сообщения уходят в DLQ, а обработчик устойчив к дублям (идемпотентен).


## Коды статуса:
-	200 OK — успешный ответ
-	201 Created — ресурс создан
-	204 No Content — успешно, без тела
-	400 Bad Request — неверные данные
-	404 Not Found — ресурс не найден
-	422 Unprocessable Entity — некорректные данные по смыслу
-	500 Internal Server Error — внутренняя ошибка

# Примечания по конфигурации и требования

Для запуска требуется:

Go: версия 1.25.1

<img width="841" height="232" alt="Установка Git и Go" src="https://github.com/user-attachments/assets/8e01d831-5a7f-4376-8348-9052b240aec9" />


# Команды запуска/сборки
## 1) Клонировать данный репозиторий в удобную для вас папку:
```Powershell
git clone https://github.com/kayahan81/pz30
```
## 2) Перейти в папку pz19:
```Powershell
cd pz30
```
## 3) Загрузка зависимостей:
```Powershell
go mod tidy
```
## 4) Команда запуска
В первом окне
```Powershell
go run 
```

# Проверка работоспособности
## Успешная задача
```Powershell
curl -X POST http://localhost:8082/v1/jobs/process-task `
  -H "Authorization: Bearer demo-token" `
  -H "Content-Type: application/json" `
  -d '{\"task_id\":\"t_123\"}'
```
## Задача с ошибками
```Powershell
curl -X POST http://localhost:8082/v1/jobs/process-task `
  -H "Authorization: Bearer demo-token" `
  -H "Content-Type: application/json" `
  -d '{\"task_id\":\"fail\"}'
```
