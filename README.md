# Farezero Demo e Talks

Questo repository è un monorepo atto a collezionare (in opportune sottocartelle) il materiale di talk e demo del gruppo Farezero.

## Aggiunta del materiale

Per aggiungere il materiale relativo ad un talk si puo' fare:
```bash
~/talks-and-demos (⌥ anybranch) git fetch -tpf
~/talks-and-demos (⌥ main) git checkout main
~/talks-and-demos (⌥ main) git pull
~/talks-and-demos (⌥ main) mkdir argomento_del_talk
```
Dopo aver copiato all'interno tutti i contenuti (slides in pdf, cartelle, codice sorgente, Dockerfile, docker-compose.yaml, .gitignore .dockerignore specifici, ecc.) provvedere a fare push:
```bash
~/talks-and-demos (⌥ main) git add .
~/talks-and-demos (⌥ main) git commit -s -m "add argomento_del_talk"
~/talks-and-demos (⌥ main) git push origin main
```
## Struttura del repo

In tal modo, la struttura del repo sarà simile a:
```bash
~/talks-and-demos (⌥ main) tree
.
├── LICENSE
├── README.md
├── docker_k8s_demo
│   ├── Dockerfile
│   ├── README.md
│   ├── docker-k8s-demo-slides.pdf
│   ├── go.mod
│   ├── main.go
│   └── serverinfo-deployment.yaml
└── k8s_autoscaling_linuxday_2024
    ├── README.md
    ├── metrics-server.yaml
    ├── php-apache.yaml
    └── slides.pdf
```