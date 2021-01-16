# netcon-cli

## contestの初期化

スコアサーバーで問題を開いたときにURLに書かれているUUIDがProblemIDになる
https://dev.netcon.janog.gr.jp/problems/968fd81d-d511-4ab5-84db-13fc756f3d7c#answers

mapping.example.yaml
```yaml
- problem_id: 89bc780e-7a54-4015-8327-125564a7da50
  machine_image_name: image-aki
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: d14ccfff-6410-4aea-a31d-d323f8050214
  machine_image_name: image-kit
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: 6b0b1605-9021-4848-a4ab-246f22ffcb61
  machine_image_name: image-kny
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: 8b8082f6-d7df-4555-ae25-feb07f857987
  machine_image_name: image-nas
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: 968fd81d-d511-4ab5-84db-13fc756f3d7c
  machine_image_name: image-pea
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: 227803fb-2fe1-4b89-a805-79e7679bf030
  machine_image_name: image-sc0
  project: networkcontest
  zone: asia-northeast1-b
- problem_id: 561d9876-7568-4096-b164-126cba6e4eb7
  machine_image_name: image-sc1
  project: networkcontest
  zone: asia-northeast1-b
```

```bash
netcon contest init --vmms-credential ${CREDENTIAL} --mapping-file-path ./mapping.yaml --count 1
```

## score serve

```bash
netcon scoreserver instance get --name image-sc0-xxxxx

netcon scoreserver instance list
```

## vm-management-server

```bash
netcon vmms instance create --credential ${CREDENTIAL} --problem-id 564c4898-c55c-460f-ad0a-eab5a539514f --machine-image-name image-sc0
netcon vmms instance delete --credential ${CREDENTIAL} --instance-name image-sc0-rxfe9
```

## tips

すべての問題を削除したい場合

!!!ワンライナーで書くと事故の元なので必ず2行に分けましょう!!!

```bash
PROBLEMS=$(netcon scoreserver instance list | jq -r '.[].name')
echo $PROBLEMS | xargs -n1 ./netcon vmms instance delete --credential ${CREDENTIAL} --instance-name
```
