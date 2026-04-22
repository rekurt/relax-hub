# Деплой Relax Hub в Managed Kubernetes VK Cloud

Цель: поднять `https://relax-hub.ru` в уже созданном Kubernetes-кластере. Манифесты лежат в `deploy/k8s/vkcloud`.

## Архитектура

- `relax-hub-api`: один pod Go API на `:8080`.
- `relax-hub-frontend`: один pod Vite SPA в nginx.
- `relax-hub-postgres`: один StatefulSet pod PostgreSQL/PostGIS с PVC.
- `relax-hub-redis`: один StatefulSet pod Redis с PVC.
- `relax-hub-minio`: один StatefulSet pod MinIO с PVC для S3-compatible storage.
- `Ingress`: `relax-hub.ru` маршрутизирует `/api`, `/widget.js`, `/widget.css`, `/sitemap.xml`, `/calendar/*`, `/prerender/*`, `/health`, `/ready`, `/version` в API; `/relax-hub-media/*` в MinIO; остальное во frontend.

Это single-instance схема. Она дешевле и проще, но без HA: при падении node или повреждении PVC сервис будет недоступен до восстановления. Для production обязательно настройте backup PostgreSQL и MinIO.

## Предварительные условия

1. `kubectl` смотрит в нужный VK Cloud Kubernetes-кластер.
2. В кластере есть default `StorageClass` для PVC.
3. Есть container registry с двумя образами: backend и frontend.
4. Есть публичный ingress/gateway controller и TLS. Манифесты используют стандартный `Ingress` с `ingressClassName: nginx`; если класс другой, поменяйте `deploy/k8s/vkcloud/ingress.yaml` и `deploy/k8s/vkcloud/cert-manager-cluster-issuer.yaml`.
5. На 23 апреля 2026 community `kubernetes/ingress-nginx` уже retired и не получает security fixes. Для нового production используйте контроллер, который поддерживается VK Cloud или вашей платформенной командой, либо поддерживаемую альтернативу с совместимым `IngressClass`.

## Сборка и публикация образов

Замените registry/tag на свои:

```bash
export REGISTRY=registry.example.com/relax-hub
export TAG=$(git rev-parse --short HEAD)
export PLATFORM=linux/amd64

docker buildx build --platform "$PLATFORM" -t "$REGISTRY/api:$TAG" --push .
docker buildx build --platform "$PLATFORM" -f frontend/Dockerfile -t "$REGISTRY/frontend:$TAG" --push .
```

Если node pool в VK Cloud создан на ARM, замените `PLATFORM` на `linux/arm64`.

Обновите kustomize-образы:

```bash
cd deploy/k8s/vkcloud
kustomize edit set image relax-hub-api="$REGISTRY/api:$TAG"
kustomize edit set image relax-hub-frontend="$REGISTRY/frontend:$TAG"
cd migrate
kustomize edit set image relax-hub-api="$REGISTRY/api:$TAG"
cd ../../../..
```

Если registry приватный:

```bash
kubectl -n relax-hub create secret docker-registry registry-credentials \
  --docker-server="$REGISTRY_HOST" \
  --docker-username="$REGISTRY_USER" \
  --docker-password="$REGISTRY_PASSWORD"
```

Манифесты уже используют `imagePullSecrets: registry-credentials` для API, frontend и migration job.

## Деплой из GitHub Actions

Workflow `.github/workflows/deploy-vkcloud.yml` делает полный деплой из GitHub:

1. Собирает backend и frontend Docker-образы.
2. Публикует их в registry.
3. Создает/обновляет Kubernetes Secrets.
4. Поднимает PostgreSQL, Redis и MinIO в namespace `relax-hub`.
5. Запускает миграции.
6. Выкатывает API, frontend и Ingress.

Workflow запускается автоматически при push в `master` и вручную через `workflow_dispatch`.

Обязательные GitHub Secrets:

- `KUBE_CONFIG_B64`: kubeconfig для VK Cloud Kubernetes, закодированный в base64.
- `REGISTRY_PASSWORD`: token/password для container registry. Для GHCR нужен PAT с `read:packages` и `write:packages`.
- `POSTGRES_PASSWORD`: пароль PostgreSQL внутри кластера.
- `REDIS_PASSWORD`: пароль Redis внутри кластера.
- `MINIO_ROOT_PASSWORD`: root password MinIO.
- `BANI_JWT_SECRET`: JWT secret минимум 32 символа, лучше `openssl rand -hex 32`.

Обязательные GitHub Variables или Secrets:

- `REGISTRY_USERNAME`: пользователь registry. Если не задан, workflow использует `${{ github.actor }}`.

Опциональные GitHub Variables:

- `REGISTRY_SERVER`: default `ghcr.io`.
- `API_IMAGE`: default `ghcr.io/<owner>/relax-hub-api`.
- `FRONTEND_IMAGE`: default `ghcr.io/<owner>/relax-hub-frontend`.
- `DEPLOY_PLATFORM`: default `linux/amd64`.
- `POSTGRES_DB`: default `relax_hub`.
- `POSTGRES_USER`: default `relax_hub`.
- `MINIO_ROOT_USER`: default `relaxhub`.

Опциональные GitHub Secrets для внешних интеграций:

- `BANI_PAYMENT_YOOKASSA_SHOP_ID`
- `BANI_PAYMENT_YOOKASSA_SECRET_KEY`
- `BANI_PAYMENT_YOOKASSA_PAYOUT_AGENT_ID`
- `BANI_PAYMENT_YOOKASSA_PAYOUT_SECRET_KEY`
- `BANI_PAYMENT_BEPAID_SHOP_ID`
- `BANI_PAYMENT_BEPAID_SECRET_KEY`
- `BANI_OAUTH_VK_CLIENT_ID`
- `BANI_OAUTH_VK_CLIENT_SECRET`
- `BANI_OAUTH_YANDEX_CLIENT_ID`
- `BANI_OAUTH_YANDEX_CLIENT_SECRET`
- `BANI_OAUTH_GOOGLE_CLIENT_ID`
- `BANI_OAUTH_GOOGLE_CLIENT_SECRET`
- `BANI_TELEGRAM_BOT_TOKEN`
- `BANI_EMAIL_HOST`
- `BANI_EMAIL_USERNAME`
- `BANI_EMAIL_PASSWORD`
- `BANI_EMAIL_FROM`
- `BANI_SMS_API_KEY`
- `BANI_WEBPUSH_VAPID_PUBLIC_KEY`
- `BANI_WEBPUSH_VAPID_PRIVATE_KEY`
- `BANI_WEBPUSH_VAPID_CONTACT`
- `BANI_GEO_ISOCHRONE_API_KEY`
- `BANI_GEO_YANDEX_SEARCH_API_KEY`
- `BANI_FISCAL_ATOL_LOGIN`
- `BANI_FISCAL_ATOL_PASSWORD`
- `BANI_FISCAL_ATOL_GROUP_CODE`

`KUBE_CONFIG_B64` должен работать на чистом GitHub runner. Если локальный kubeconfig использует exec-plugin или локальный CLI-токен, создайте отдельный kubeconfig с service account token или статическим токеном VK Cloud.

macOS:

```bash
base64 -i ~/.kube/config | pbcopy
```

Linux:

```bash
base64 -w 0 ~/.kube/config
```

Если используете `production` environment в GitHub, добавьте Secrets/Variables в этот environment. Иначе можно добавить их на уровне repository.

## Секреты

Создайте namespace и Secret:

```bash
kubectl apply -f deploy/k8s/vkcloud/namespace.yaml
cp deploy/k8s/vkcloud/secret.example.yaml /tmp/relax-hub-secrets.yaml
```

Заполните `/tmp/relax-hub-secrets.yaml` реальными значениями и примените:

```bash
kubectl apply -f /tmp/relax-hub-secrets.yaml
```

Важные поля:

- `POSTGRES_PASSWORD` и пароль внутри `BANI_DATABASE_DSN` должны совпадать.
- Если в пароле PostgreSQL есть спецсимволы, URL-encode значение в `BANI_DATABASE_DSN`.
- `REDIS_PASSWORD` и `BANI_REDIS_PASSWORD` должны совпадать.
- `MINIO_ROOT_USER` должен совпадать с `BANI_STORAGE_ACCESS_KEY`.
- `MINIO_ROOT_PASSWORD` должен совпадать с `BANI_STORAGE_SECRET_KEY`.
- `BANI_JWT_SECRET`: минимум 32 символа, лучше `openssl rand -hex 32`.

`BANI_ENVIRONMENT` в `configmap.yaml` выставлен в `staging` намеренно: single-instance PostgreSQL и MinIO внутри кластера работают по plain TCP/HTTP. Если поставить `production`, текущая валидация приложения потребует TLS для PostgreSQL и object storage.

## TLS

Если в кластере уже есть `cert-manager`, примените ClusterIssuer:

```bash
kubectl apply -f deploy/k8s/vkcloud/cert-manager-cluster-issuer.yaml
```

Если TLS выпускается другим способом, удалите annotation `cert-manager.io/cluster-issuer` из `ingress.yaml` и укажите существующий `tls.secretName`.

## Применение

Сначала поднимите конфиг и stateful-инфраструктуру:

```bash
kubectl -n relax-hub apply -f deploy/k8s/vkcloud/configmap.yaml
kubectl -n relax-hub apply -f deploy/k8s/vkcloud/postgres.yaml
kubectl -n relax-hub apply -f deploy/k8s/vkcloud/redis.yaml
kubectl -n relax-hub apply -f deploy/k8s/vkcloud/minio.yaml

kubectl -n relax-hub rollout status statefulset/relax-hub-postgres
kubectl -n relax-hub rollout status statefulset/relax-hub-redis
kubectl -n relax-hub rollout status statefulset/relax-hub-minio
```

Затем запустите миграции:

```bash
kubectl -n relax-hub delete job relax-hub-migrate --ignore-not-found
kubectl apply -k deploy/k8s/vkcloud/migrate
kubectl -n relax-hub wait --for=condition=complete job/relax-hub-migrate --timeout=300s
```

После успешных миграций выкатывайте приложение:

```bash
kubectl apply -k deploy/k8s/vkcloud
kubectl -n relax-hub rollout status deploy/relax-hub-api
kubectl -n relax-hub rollout status deploy/relax-hub-frontend
```

## DNS

Получите внешний адрес Ingress:

```bash
kubectl -n relax-hub get ingress relax-hub
kubectl get svc -A | grep -i loadbalancer
```

В DNS-зоне домена создайте или обновите запись:

- `A relax-hub.ru -> EXTERNAL-IP` для IPv4.
- `AAAA relax-hub.ru -> EXTERNAL-IPv6`, если балансировщик выдал IPv6.

После обновления DNS проверьте:

```bash
dig +short relax-hub.ru
curl -I https://relax-hub.ru
curl https://relax-hub.ru/health
curl https://relax-hub.ru/api/v1/cities
```

Сертификат Let's Encrypt появится только после того, как `relax-hub.ru` указывает на внешний адрес ingress controller.

## Backups

Минимально настройте регулярные backups:

```bash
kubectl -n relax-hub exec statefulset/relax-hub-postgres -- \
  pg_dump -U relax_hub -d relax_hub --format=custom --file=/tmp/relax_hub.dump
```

Для MinIO нужен отдельный backup PVC или репликация bucket-а наружу. Без этого загруженные пользователями файлы живут только на PVC `relax-hub-minio`.

## Быстрая диагностика

```bash
kubectl -n relax-hub get pods,svc,statefulset,pvc,ingress,certificate
kubectl -n relax-hub logs deploy/relax-hub-api --tail=100
kubectl -n relax-hub logs statefulset/relax-hub-postgres --tail=100
kubectl -n relax-hub logs statefulset/relax-hub-redis --tail=100
kubectl -n relax-hub logs statefulset/relax-hub-minio --tail=100
kubectl -n relax-hub describe ingress relax-hub
kubectl -n relax-hub describe certificate relax-hub-ru-tls
```

Если API pod не готов, первым делом смотрите `/ready`: он проверяет PostgreSQL и Redis.
