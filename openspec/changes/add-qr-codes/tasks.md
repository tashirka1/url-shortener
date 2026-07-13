## 1. Dependency

- [ ] 1.1 Add `github.com/skip2/go-qrcode` to `go.mod`

## 2. Data Layer

- [ ] 2.1 Add `GetLinkByCode(ctx, code string, userId int) (model.Link, error)` to `LinkStorage` interface and SQL implementation in `link/storage/link.go` (SELECT без инкремента clicks)
- [ ] 2.2 Add `GetLinkByCode(ctx, code string, userId int) (model.Link, error)` to `LinkService` interface and implementation in `link/service/link.go`

## 3. Handler

- [ ] 3.1 Add `GetQRCode` handler in `link/handler/handler.go`: получает code из URL, вызывает service.GetLinkByCode, генерирует QR через `qrcode.Encode()`, возвращает HTML с `<dialog>` и data URI

## 4. View

- [ ] 4.1 Add templ component `QRCodeModal` — `<dialog>` с `<img>` base64 data URI, текстом короткой ссылки и кнопкой закрытия
- [ ] 4.2 Add кнопку "QR" в строку `Link` и `LastLink` компоненты в `link/view/link.templ`

## 5. Routing

- [ ] 5.1 Add route `GET /link/:code/qr` в link group в `cmd/url-shortener/main.go`

## 6. Verify

- [ ] 6.1 Run `make check` — линтер и тесты должны проходить
- [ ] 6.2 Run `make build-bin` — бинарник должен собираться
