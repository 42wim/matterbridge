# Contributing

## Настройки

`vksdk` написан на [Go](https://golang.org/).

Требования:

- [Go 1.18+](https://golang.org/doc/install)
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [global .gitignore](https://help.github.com/en/articles/ignoring-files#create-a-global-gitignore)

Сделайте fork и клонируйте `vksdk` куда угодно:

```sh
git clone git@github.com:<your name>/vksdk.git
```

Создайте новую ветку

```sh
git checkout -b <name_of_your_new_branch>
```

## Тестирование изменений

Для начала проверьте ваш код с помощью
[golangci-lint](https://github.com/golangci/golangci-lint)

```sh
golangci-lint run
```

Затем можно запускать тесты

```sh
# SERVICE_TOKEN=""
# GROUP_TOKEN=""
# CLIENT_SECRET=""
# USER_TOKEN=""
# WIDGET_TOKEN=""
# MARUSIA_TOKEN=""
# CLIENT_ID="123456"
# GROUP_ID="123456"
# ACCOUNT_ID="123456"
go test ./...
```

Задавать токены не обязательно - тесты с их использованием будут пропущены.
**Не** рекомендуется задавать свой `USER_TOKEN`, так как тесты делают много
страшных вещей.

Настройки для VSCode `.vscode/setting.json`

```json
{
  "go.testEnvVars": {
    "SERVICE_TOKEN": "",
    "WIDGET_TOKEN": "",
    "MARUSIA_TOKEN": "",
    "GROUP_TOKEN": "",
    "CLIENT_SECRET": "",
    "USER_TOKEN": "",
    "CLIENT_ID": "123456",
    "GROUP_ID": "123456",
    "ACCOUNT_ID": "123456"
  }
}
```

## Создание коммита

Сообщения коммитов должны быть хорошо отформатированы, и чтобы сделать их
«стандартизированным», мы используем
[Conventional Commits](https://www.conventionalcommits.org/ru).

```sh
git add .
git commit
```

## Отправьте pull request

Отправьте изменения в ваш репозиторий

```sh
git push origin <name_of_your_new_branch>
```

Затем откройте [pull request](https://github.com/SevereCloud/vksdk/pulls)
с веткой master
