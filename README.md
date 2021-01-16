# netcon-cli

## contestの初期化

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
