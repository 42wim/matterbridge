# Bots Long Poll API

[![PkgGoDev](https://pkg.go.dev/badge/github.com/SevereCloud/vksdk/v2/longpoll-bot)](https://pkg.go.dev/github.com/SevereCloud/vksdk/v2/longpoll-bot)
[![VK](https://img.shields.io/badge/developers-%234a76a8.svg?logo=VK&logoColor=white)](https://vk.com/dev/bots_longpoll)

## Подключение Bots Long Poll API

Long Poll настраивается автоматически. Вам не требуется заходить в настройки
сообщества.

### Версия API

Данная библиотека поддерживает версию API **5.131**.

### Инициализация

Модуль можно использовать с ключом доступа пользователя, полученным в
Standalone-приложении через Implicit Flow(требуются права доступа: **groups**)
или с ключом доступа сообщества(требуются права доступа: **manage**).

В начале необходимо инициализировать api:

```go
vk := api.NewVK("<TOKEN>")
```

А потом сам longpoll

```go
lp, err := longpoll.NewLongPoll(vk api.VK, groupID int)
// По умолчанию Wait = 25
// lp.Wait = 90
// lp.Ts = "123"
```

### HTTP client

В модуле реализована возможность изменять HTTP клиент - `lp.Client`

Пример прокси

```go
dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
httpTransport := &http.Transport{
	Dial:              dialer.Dial,
	// DisableKeepAlives: true,
}
httpTransport.Dial = dialer.Dial
lp.Client.Transport = httpTransport
```

### Обработчик событий

Для каждого события существует отдельный обработчик, который передает функции
`ctx` и `object`.

Пример для события `message_new`

```go
lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
	...
})
```

Если вы хотите получать полный ответ от Long Poll(например для сохранения `ts`
или специальной обработки `failed`), можно воспользоваться следующим обработчиком.

```go
lp.FullResponse(func(resp object.LongPollBotResponse) {
	...
})
```

Полный список событий Вы найдёте [в документации](https://vk.com/dev/groups_events)

### Контекст

Поля `groupID`, `ts` и `eventID` передаются в `ctx`. Чтобы получить их, можно
воспользоваться следующими функциями:

```go
groupID := events.GroupIDFromContext(ctx)
eventID := events.EventIDFromContext(ctx)
ts := longpoll.TsFromContext(ctx)
```

### Запуск и остановка

```go
// Запуск
if err := lp.Run(); err != nil {
	log.Fatal(err)
}

// Безопасное завершение
// Ждет пока соединение закроется и события обработаются
lp.Shutdown()

// Закрыть соединение
// Требует lp.Client.Transport = &http.Transport{DisableKeepAlives: true}
lp.Client.CloseIdleConnections()
```

## Пример

```go
package main

import (
	"log"

	"github.com/SevereCloud/vksdk/v2/api"

	longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"github.com/SevereCloud/vksdk/v2/events"
)

func main() {
	vk := api.NewVK("<TOKEN>")
	lp, err := longpoll.NewLongPoll(vk, 12345678)
	if err != nil {
		panic(err)
	}

	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		log.Print(obj.Message.Text)
	})

	lp.Run()
}

```
