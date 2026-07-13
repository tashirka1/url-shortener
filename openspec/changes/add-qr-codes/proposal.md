## Why

URL shortener без возможности получить QR-код — половина сервиса. Пользователи создают короткие ссылки для печатных материалов, визиток, рекламных баннеров — везде нужен QR. Сейчас чтобы получить QR, нужно идти в сторонний сервис. Добавляем генерацию прямо в приложение.

## What Changes

- Новая зависимость: `github.com/skip2/go-qrcode` (генерация QR в PNG)
- Новый эндпоинт `GET /link/:code/qr`, авторизованный, отдаёт `image/png`
- Кнопка "QR" в каждой строке таблицы ссылок
- Модальное окно с QR-кодом при нажатии на кнопку
- Никаких изменений в БД — QR генерируется на лету из URL ссылки

## Capabilities

### New Capabilities
- `qr-code`: Генерация QR-кода для короткой ссылки. Получение PNG через `/link/:code/qr`, отображение в модальном окне на странице списка ссылок.

### Modified Capabilities

(no existing specs to modify)

## Impact

- `go.mod` + `go.sum`: новая зависимость `github.com/skip2/go-qrcode`
- `internal/link/handler/handler.go`: новый хендлер `GetQRCode`
- `internal/link/view/link.templ`: кнопка QR + модальное окно
- `cmd/url-shortener/main.go`: новый роут в link group
- Dockerfile: без изменений (Go библиотека, компилируется в бинарник)
- Не затрагивает: auth, rps, core, storage, service, БД
