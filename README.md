Golang Developer Test Cases
====

Вы решили создать конкурента Amazon S3 и знаете как сделать лучший сервис хранения файлов.

На сервер A по REST присылают файл, его надо разрезать на 6 примерно равных частей и сохранить на серверах хранения Bn (n ≥ 6).

При REST-запросе на сервер A нужно достать куски с серверов Bn склеить их и отдать файл.

# Имеем следующее
1. Один сервер для REST запросов
2. Несколько серверов для хранения кусков файлов

# Ограничения
1. Реализовать сервис и тестовый модуль, который обеспечит проверку его работы
2. Сервера для хранения могут добавляться в систему в любой момент, но не могут удаляться из системы
3. Предусмотреть равномерное заполнение серверов хранения
4. Код разместить в Github

# System schemes
![Put](./doc/New_Amason_S3_put.svg)
![Get](./doc/New_Amason_S3_get.svg)

# Схема базы данных

![Scheme DB](./doc/database_diagram.png)

# Соображения по архитектуре

## Реализация
