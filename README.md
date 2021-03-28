# Практическое задание по курсу Goland

## Написание тестов 
#### подробное условие доступно в файле taks.md

## Запуск
```
go test -cover
```

## Построение покрытия
#### Для построения покрытия ваш код должен находиться внутри GOPATH

```
go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html
```