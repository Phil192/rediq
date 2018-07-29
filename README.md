# Простая имплементация похожего на Redis кеша в памяти
```
-- Key-value хранилище строк, списков, словарей
-- TTL на каждый ключ
-- Операторы
    --GET
    --SET
    --REMOVE
    --KEYS
-- Дополнительные операции (получить значение по индексу из списка,
    получить значение по ключу из словаря)
-- Golang API Client к кешу
-- API tcp(telnet)/REST API
```
Предоставить несколько тестов, документацию по API,
документацию по развертыванию, несколько кейсов
и примеров использования кеша вызовов http/tcp api.

## Дополнительно:
```
-- Сохранение на диск
-- Масштабирование
-- Авторизация
-- Нагрузочные тесты
```

Кэш сервер создается с следующими параметрами коммандной строки:
```
logLevel     int                - уровень детализации логирования
socket       string             - сокет, который слушает сервер кэша
log          string             - путь к файлу для лога
dump         string             - путь к файлу дампа кэша
stdout       bool               - выводить логи сервера только в терминал
gcCap        int                - емкость канала, по которым данные передаются сборщику мусора
shards       int                - максимальное число шардов, по умолчанию 256
items        int                - максимальное число элементов в шарде, по умолчанию 2048
```
## Golang API
Хранилище хранит объекты типа Value, содержащие поля:
```
body     interface{}
ttl      int
type     InputType int
```
API поддерживает следующие методы:
```
Set(string, interface{}, int) (error)
Get(string) (*Value, error)
Remove(key string) (error)
Keys() ([]string)
GetBy(string, interface{}) (interface{}, error)
```
Кэш создается методом NewCache, принимающий параметры:
```
ShardsNum      int               - максимальное число шардов, по умолчанию 256
ItemsPerShard  int               - максимальное число элементов в шарде, по умолчанию 2048
DumpPath       int               - путь к файлу дампа кэша
GCCap          int               - емкость канала, по которым данные передаются сборщику мусора
```
Запускается кэш методом Run(), который читает и сохраняет данные дампа
и запускает обратный отсчет TTL, а затем удаляет просроченные элементы.
В случае получения сигнала (например SIGINT), кэш сбрасывает данные в дамп.
Кэш агностичен по отношению к App и его API можно использовать независимо.

## REST HTTP API:
Перед запуском сервера нужно создать App с помощью метода
NewApp(), который принимает следующие параметры:
```
Cache          ptr                - интерфейс кэша
LogFile        int                - путь к файлу для лога
SetSocket      string             - сокет, который слушает App
```
В качестве роутера используется gin-gonic (по причине radix tree).
Методом App.RouteAPI() создается необходимый роутинг и данный метод
принимает параметром необходимый gin.Engine
Для запуска HTTP сервера нужно выполнить метод App.ListenAndServe().

REST API принимает и возвращает данные в формате JSON (Set
возвращает служебную информацию - url сохраненного объекта)
Все URL начинаются с /api/v1
```

| Хэндлер  | Метод  | Url                  | Body                               | Пример успешного ответа          | Пример ошибки                                                    |
|----------|--------|----------------------|------------------------------------|----------------------------------|------------------------------------------------------------------|
| Keys     | GET    | /keys/:key           | --                                 | ["test","tist","tost"]           | --                                                               |
| Get      | GET    | /get/:key            | --                                 | [{"body":"123","ttl":2,"type":0}]| {"error": "not found in cache"}                                  |
| GetBy    | GET    | /getby/?key=&index=  | --                                 | ["ok"]                           | {"error": "cant get item at index"}                              |
| Remove   | DELETE | /remove/:key         | --                                 | "OK"                             | --                                                               |
| Set      | POST   | /set                 | {"key":"123","value":"3","ttl":0}  | [0.0.0.0:8081/api/v1/get/123]    | {"error":"invalid character 'a' looking for beginning of value"} |

```
REST HTTP интерактивный клиент реализует интерфейс Cache.
Интерактивная оболочка реализована с помощью ishell.
Доступны команды:
```
set    <key> <value> <ttl>
get    <key>
getby  <key> <index>
remove <key>
keys   <mask>
```

## Развертывание
```
go get -u github.com/phil192/rediq (или git clone git@github.com:Phil192/rediq.git)
cd $GOPATH/github.com/phil192/rediq

для сервера:
docker build -t <srv-name> . -f dep/Dockerfile.srv
для клиента:
docker build -t <cli-name> . -f dep/Dockerfile.cli

docker run -p 8888:8081 <cli or srv name>
```

## Масштабирование:
Реализовано посредством деления мапы кэша на шарды, к которым "прикрепляется" ответственный RWMutex.
Новые шарды создаются по маскам ключей. Пустые шарды удаляются.


## Юнит и интеграционное тестирование:

Кэш имеет файл storage/cache_test.go, в котором реализованы юнит-тесты.
Сервер кэша имеет файл rest/listener_test.go, в котором реализованы интеграционные тесты.
Все тесты успешны.
Для запуска используются команды "make test-cache" и "make test-rest" соответственно.

## Гонка данных
go build -race успешен.

## Нагрузочное тестирование:
Выполнено для самого "тяжелого" метода (за исключением метода Keys, который apriori "тяжелый").
Тестирование выполнено посредством Apache Benchmark.
```
ab -p sample -t 60 -n 200 -c 200 -v 2 http://0.0.0.0:8081/api/v1/set

Concurrency Level:      200
Time taken for tests:   0.208 seconds
Complete requests:      200
Failed requests:        0
Total transferred:      28800 bytes
Total body sent:        34000
HTML transferred:       5400 bytes
Requests per second:    959.62 [#/sec] (mean)
Time per request:       208.416 [ms] (mean)
Time per request:       1.042 [ms] (mean, across all concurrent requests)
Transfer rate:          134.95 [Kbytes/sec] received
                        159.31 kb/s sent
                        294.26 kb/s total

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   0.7      2       3
Processing:     1   29  23.0     28     204
Waiting:        1   28  22.7     28     204
Total:          4   31  22.7     31     205

Percentage of the requests served within a certain time (ms)
  50%     31
  66%     32
  75%     33
  80%     34
  90%     49
  95%     59
  98%     61
  99%    205
 100%    205 (longest request)
```