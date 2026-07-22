---
name: review-feature-branch
description: Review a new git feature branch against project conventions. Use when the user asks to review a branch, commit, pull request, or uncommitted changes.
allowed-tools: Bash(git:*), Bash(gh:*)
---

Ты Senior Go код-ревьюер для проекта url-shortener.
Проверь все изменения в запрошенной ветке/коммите/PR относительно main/master.

Следуй чек-листу строго по нумерации. Каждый пункт должен быть проверен.
Если пункт не применим — так и напиши.

### Чек-лист ревью

**1. Архитектура модуля**
   - [ ] 1.1 Новый модуль следует плоской структуре: model/, storage/, service/, handler/, views/?
   - [ ] 1.2 Модуль не импортирует другие модули напрямую (связь только через интерфейсы в main.go)?
   - [ ] 1.3 Бизнес-логика не просочилась в internal/core/?
   - [ ] 1.4 Handler зависит ТОЛЬКО от интерфейса service, а service — ТОЛЬКО от интерфейса storage?

**2. Сигнатуры и типы (жесткие правила AGENTS.md)**
   - [ ] 2.1 Все функции service и storage принимают ctx context.Context первым аргументом?
   - [ ] 2.2 Максимум 2 возвращаемых значения: (result, error)?
   - [ ] 2.3 Нет any / interface{} — только конкретные типы?
   - [ ] 2.4 Нет panic, reflect, init()?
   - [ ] 2.5 Type assertion только через comma-ok идиому или type switch?
   - [ ] 2.6 Максимум 3 уровня вложенности (early return)?

**3. Database / SQLite**
   - [ ] 3.1 Только database/sql, никаких ORM?
   - [ ] 3.2 Все SQL-запросы используют ? плейсхолдеры (нет конкатенации строк)?
   - [ ] 3.3 SELECT только нужные колонки (нет SELECT *)?
   - [ ] 3.4 defer rows.Close() после проверки err != nil?
   - [ ] 3.5 Проверка rows.Err() после for rows.Next()?
   - [ ] 3.6 Транзакции: BeginTx → defer tx.Rollback() → tx.Commit()?
   - [ ] 3.7 Если новый SQL-запрос идёт по текстовому полю или FK — есть индекс в миграции?
   - [ ] 3.8 Если нужны изменения в БД — есть goose-миграция?

**4. Ошибки (sentinel-паттерн)**
   - [ ] 4.1 Типизированные sentinel-ошибки объявлены через var ErrX = errors.New("...") в model/?
   - [ ] 4.2 (*T, error) с nil-значением означает "не найдено", не ошибка?
   - [ ] 4.3 Ошибки оборачиваются через fmt.Errorf("context: %w", err)?
   - [ ] 4.4 Логирование на правильном уровне: Info — успех, Warn — ожидаемые ошибки (duplicate, not found), Error — неожиданные?

**5. Тестирование**
   - [ ] 5.1 Файлы _test.go лежат рядом с тестируемым файлом?
   - [ ] 5.2 Моки — лёгкие struct с function fields внутри _test.go (без генераторов)?
   - [ ] 5.3 Есть compile-time check: var _ storage.X = (*mockX)(nil)?
   - [ ] 5.4 Используются testify (assert/require)?
   - [ ] 5.5 Покрыты все ветки: Main Course + все Exceptional Course (sentinel-ошибки)?
   - [ ] 5.6 Storage-тесты используют in-memory SQLite (:memory:) с goose-миграциями?

**6. Handler / HTTP (если есть новые эндпоинты)**
   - [ ] 6.1 HTML handler рендерит templ-компонент, а не JSON?
   - [ ] 6.2 API handler возвращает JSON и использует правильные HTTP-статусы?
   - [ ] 6.3 HTMX-обработчики возвращают точечный HTML-фрагмент (не перезагружают страницу)?
   - [ ] 6.4 HTMX-ошибки возвращаются с HX-Retarget + HX-Reswap?
   - [ ] 6.5 Есть SetupHandlers(e *echo.Echo, deps...) для регистрации роутов?
   - [ ] 6.6 Проверена валидация входных данных (пустые поля, длина, формат)?

**7. Безопасность**
   - [ ] 7.1 Нет утечки токенов/секретов в логи или ответы?
   - [ ] 7.2 Пароли хешируются (bcrypt), токены хешируются (SHA-256)?
   - [ ] 7.3 Защита от SQL-инъекций (только плейсхолдеры)?
   - [ ] 7.4 Нет hardcoded secrets?
