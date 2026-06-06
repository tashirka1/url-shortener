# Bash: шпаргалка

## Навигация

```bash
pwd                  # текущая директория
cd <dir>             # перейти в директорию
cd ..                # на уровень выше
cd ~                 # домой
ls -lah              # список файлов (все, подробно, читаемые размеры)
```

## Работа с файлами

```bash
mkdir -p <dir>       # создать директорию (с родителями)
touch <file>         # создать пустой файл или обновить время
cp <src> <dst>       # копировать
mv <src> <dst>       # переместить/переименовать
rm -rf <path>        # удалить рекурсивно (осторожно!)
cat <file>           # вывести содержимое
less <file>          # просмотр с прокруткой
head -n 5 <file>     # первые 5 строк
tail -f <file>       # хвост файла + следить за изменениями
nano <file>          # редактировать файл
chmod +x <file>      # сделать исполняемым
```

## Диски и память

```bash
df -h                # свободное место на разделах
du -sh <dir>         # размер директории суммарно
```

## Процессы

```bash
htop                 # мониторинг процессов (интерактивно)
ps aux               # список всех процессов
kill <pid>           # завершить процесс
kill -9 <pid>        # принудительно
journalctl -u <service> -f  # логи systemd (хвост)
```

## Установка пакетов (Debian/Ubuntu)

```bash
sudo apt update          # обновить список пакетов
sudo apt install <pkg>   # установить пакет
```

## Сеть

```bash
curl <url>           # HTTP-запрос
curl -X POST -d '{}' <url>  # POST с телом
wget <url>           # скачать файл
ping <host>          # проверить доступность
dig <domain>         # DNS-запрос
```

## SSH

```bash
ssh user@host        # подключиться по SSH
ssh-copy-id user@host   # скопировать свой ключ на сервер
scp file user@host:path  # копировать файл на сервер
scp user@host:path/file .  # скачать файл с сервера
```

## Поиск

```bash
find <dir> -name "*.go"   # найти файлы по имени
grep -r "text" .          # искать текст в файлах
which <cmd>               # путь к исполняемому файлу
```

## Git

```bash
git status           # состояние рабочей директории
git add <file>       # добавить в индекс
git commit -m "msg"  # создать коммит
git pull             # стянуть изменения
git push             # отправить изменения
git reset <file>     # убрать файл из индекса
git reset --hard     # откатить все локальные изменения (осторожно)
git tag <name>       # создать тег
git tag -d <name>    # удалить тег локально
git push --tags      # отправить теги в remote
git log --oneline --graph --all  # история коммитов (граф)
git diff             # незакоммиченные изменения
git branch           # список веток
git checkout -b <branch>  # создать и переключиться на ветку
git merge <branch>   # слить ветку в текущую
git stash            # спрятать изменения
git stash pop        # достать из стеша
git rebase <branch>  # перебазировать на ветку
```

## Архивы

```bash
tar -czf archive.tar.gz <dir>   # упаковать
tar -xzf archive.tar.gz         # распаковать
zip -r archive.zip <dir>        # упаковать в zip
unzip archive.zip               # распаковать zip
```

## Docker

```bash
docker ps                   # запущенные контейнеры
docker ps -a                # все контейнеры
docker image ls             # список образов
docker system prune -a      # отчистить всё (образы, контейнеры, кеш)
docker logs -f <container>  # логи контейнера (хвост)
docker exec -it <container> <cmd>  # выполнить команду внутри
docker compose up -d        # запустить compose-проект
docker compose down         # остановить compose-проект
```
