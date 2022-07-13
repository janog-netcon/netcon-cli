# README.md

## 実行方法

### 手元からgcloudを使う場合

```sh
./scripts/coordinate.py --project example_project --zones asia-northeast1-b asia-northeast2-a asia-east1-b
```

### dockerコンテナを使う場合

```sh
docker run -i -t --rm -v creds.json:/creds.json -e GOOGLE_APPLICATION_CREDENTIALS=/creds.json coordinate --dry_run true --auth_type sa --project example_project --zones asia-northeast1-b asia-northeast2-a asia-east1-b --loop true --loop_interval 30 --vmdb_url http://host.docker.internal:8905
```
