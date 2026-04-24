# Postgres backup — nightly pg_dump to MinIO

`CronJob relax-hub-postgres-backup` fires daily at **02:00 Europe/Moscow**, dumps the entire database, gzips it, and uploads to `local/relax-hub-backups/postgres/` inside the in-cluster MinIO. Files older than 30 days are pruned by the same job.

The job is applied automatically by the main kustomize overlay — once `deploy/k8s/vkcloud/kustomization.yaml` lists `backup/cronjob.yaml` under `resources`, each workflow run keeps it synced.

## Current limits (intentional)

- **Same failure domain**: MinIO runs on the same cluster / same Ceph volumes as Postgres. A node loss kills both. This protects against **logical** corruption only.
- **No continuous WAL archiving**: RPO is up to 24 h. Consider `pgBackRest` + external S3 when we hit real traffic.

## Trigger manually

```bash
export KUBECONFIG=.deploy-local/kubeconfig.yaml
kubectl -n relax-hub create job \
  --from=cronjob/relax-hub-postgres-backup \
  manual-backup-$(date +%s)
kubectl -n relax-hub logs -f job/manual-backup-* --tail=-1
kubectl -n relax-hub exec sts/relax-hub-minio -- mc ls local/relax-hub-backups/postgres/
```

## Restore procedure

**DESTRUCTIVE** — overwrites the database. Do this only on a fresh pod or after confirming with ops.

```bash
export KUBECONFIG=.deploy-local/kubeconfig.yaml

# 1. Pick the snapshot
kubectl -n relax-hub exec sts/relax-hub-minio -- mc ls local/relax-hub-backups/postgres/

# 2. Pull from MinIO → through local machine → into Postgres pod
FILE=relax_hub_20260424T020000Z.sql.gz    # <— replace
kubectl -n relax-hub exec sts/relax-hub-minio -- \
  mc cp "local/relax-hub-backups/postgres/${FILE}" "/tmp/${FILE}"
kubectl -n relax-hub cp "relax-hub-minio-0:/tmp/${FILE}" "/tmp/${FILE}"
kubectl -n relax-hub cp "/tmp/${FILE}" "relax-hub-postgres-0:/tmp/${FILE}"

# 3. Stop the API + any other writer before restoring (optional but safer)
kubectl -n relax-hub scale deploy/relax-hub-api --replicas=0

# 4. Restore
kubectl -n relax-hub exec -it sts/relax-hub-postgres -- bash -c "
  gunzip -c /tmp/${FILE} | psql -U relax_hub -d relax_hub
"

# 5. Resume writes
kubectl -n relax-hub scale deploy/relax-hub-api --replicas=1
```

## Monitor

```bash
# Most recent CronJob runs
kubectl -n relax-hub get jobs -l app.kubernetes.io/component=backup --sort-by=.metadata.creationTimestamp | tail -5

# Latest dump size sanity-check
kubectl -n relax-hub exec sts/relax-hub-minio -- \
  mc ls local/relax-hub-backups/postgres/ | sort | tail -3
```

## Disable temporarily

```bash
kubectl -n relax-hub patch cronjob relax-hub-postgres-backup \
  -p '{"spec":{"suspend":true}}'
```

Reverse with `suspend:false`.
