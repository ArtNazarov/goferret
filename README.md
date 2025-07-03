# goferret

[Logo](https://dl.dropbox.com/scl/fi/jfhjfas6zuwsj460h7j9p/goferret.png?rlkey=ulydm3q3dt0rygv8wjp09xgfy&st=ng7zs85b)

**goferret** — это генератор статических сайтов, написанный на Go, который позволяет создавать HTML-страницы на основе шаблонов и атрибутов, определённых в отдельных файлах. Проект предназначен для быстрой генерации простых сайтов с использованием минималистичной структуры данных и шаблонов.

## Структура проекта

```
./
├── blocks
│   ├── footer.tpl
│   └── header.tpl
├── templates/
│   ├── blog.tpl
│   └── page.tpl
├── content/
│   ├── about/
│   │   ├── title.val
│   │   ├── content.val
│   │   ├── category.val
│   │   └── template.setting
│   └── contact/
│       ├── title.val
│       ├── email.val
│       ├── phone.val
│       ├── category.val
│       └── template.setting
└── build/           # Директория для сгенерированных HTML-файлов (создаётся автоматически)
```
- **blocks/** — содержит шаблоны глобальных блоков разметки в формате `.tpl`
Блоки становятся общедоступными атрибутами `{header}`, `{footer}` и т.д.
- **templates/** — содержит шаблоны страниц в формате `.tpl`. Каждый шаблон использует переменные в фигурных скобках, например `{title}` или `{content}`.
- **content/** — содержит поддиректории для каждой страницы сайта. В каждой поддиректории размещаются файлы с атрибутами (`*.val`) и файл `template.setting` с именем используемого шаблона. Категория страницы задается в файле `category.val`
- **build/** — автоматически создаётся для вывода сгенерированных HTML-файлов.

## Пример содержимого

**templates/page.tpl**
```
<html>
<head><title>{title}</title></head>
<body>
  <h1>{title}</h1>
  <div>{content}</div>
</body>
</html>
```

**content/about/title.val**
```
О сайте
```

**content/about/content.val**
```
Это пример страницы "О сайте", сгенерированной с помощью goferret.
```

**content/about/template.setting**
```
page
```

**content/about/category.val**
```
main
```

## Компиляция

Для сборки исполняемого файла выполните:

```
go build -o goferret goferret.go
```

В результате появится бинарный файл `./goferret`.

## Использование

1. Убедитесь, что у вас есть директории `templates/` и `content/` с соответствующими файлами.
2. Запустите генератор:

```
./goferret
```

3. Сгенерированные HTML-файлы появятся в директории `build/`.

## Требования
- Go 1.16 или новее
- Linux, macOS или Windows

## Пример вывода

```
Сгенерировано: build/contact.html
Сгенерировано: build/index.html
Сгенерировано: build/main.json
Сгенерировано: build/main.html
Генерация сайта завершена!
```

## Автор

**Артем Назаров**, Оренбург, 2025
Email: programmist.nazarov@gmail.com

Если у вас есть вопросы или предложения, пишите на email автора.
