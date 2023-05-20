# Базовые настройки ОС, firewall, пользователей
Раздел cosmos.yml: Host base configuration && hardening.

Содержит в себе следующие роли:
- users_management - взял роль со старой работы, т.к. она покрывает больщую часть задач: создаёт пользователя, прописывает ему ssh ключ, добавляет в sudoers
- sshd - коммьюнити роль, позволяет быстро и безболезненно настроить sshd, отключил доступ по паролю, т.к. на предыдужем этапе добавил пользователя с правами sudo
- lynis - вспомогательная мини-роль, устанавливающая пакет lynis, чтобы не делать вообще никаких ручных изменений на удалённом сервере
- hardening - роль, где по очереди удовлетворяются требования lynis

## Подробнее про hardeinng:
- iptables - настройки firewall
- packages - апгрейд всех пакетов до свежих версий, установка дополнительных утилит, которые рекомендовал lynis
- system_config - настройки sysctl, прав доступа к директориямм и файлам, прочие требования lynis

Не стал перевешивать sshd на другой порт, т.к. не вижу особого смысла в этом (сканеры всё равно найдут sshd на другом порту).

# Сборка, установка и настройка клиента (gaiad)
Раздел cosmos.yml: Build, install and configure gaiad

При выполнении этого этапа опирался на [официальную документацию](https://hub.cosmos.network/main/hub-tutorials/join-testnet.html)

Собрал приложение из исходников в docker, чтобы не засорять локальную и удалённую ОС и получать на выходе только исполняемый файл, который потом будет копироваться с локальной маишны на удалённую.  
Т.к. у меня Mac M1, переменная DOCKER_DEFAULT_PLATFORM=linux/amd64 оказалась обязательной при сборке.

После копирования собранного пакета в /usr/sbin происходит проверка, было ли приложение инициализировано ранее и первичная инициализация, если этого не было сделано.

Добавил переменные для изменения app.toml, чтобы активировать API, думал экспортить статус оттуда, но изучать API оказалось долго.  
Плюсом такого решения была бы возможность запускать экспортер удалённо, например, в сети мониторинга, а минусом, конечно же, безопасность, но эндпоинт API можно жёстко закрыть файрволлом и этот вопрос будет решён.

# Экспортер
Раздел cosmos.yml: Build, and run gaiad-exporter

Исполняемый файл собирается в Docker на локальной машине и приносится на хост с помощью ansible, запуск через systemd unit

Добавил:
- тип данных GaiaStatus, содержащий все поля из JSON, кторый возвращается при вызове `gaiad status` - задел на будущее, если ещё какие-то метрики или лейблы понадобятся.  
- флаг с портом, на котором будет запускаться экспортер
- флаг metrics_path, если вдруг понадобится перевесить с дефолтного `/metrics`

Можно доработать:
- добавить таймаут на выполнение `gaiad status`, передавать через флаг
- добавить логирование
- добавить валидацию JSON, если при вызове `gaiad status` возникнет ошибка, например на число открытых файлов или если gaiad убьёт OOM
- добавить в метрику лейбл с юзером, инициализировавшим gaiad

# Запуск playbook
```bash
ansible-playbook -bCDi ansible/inventories/cosmos.hosts ansible/cosmos.yml -l cosmos_test -t gaia
```