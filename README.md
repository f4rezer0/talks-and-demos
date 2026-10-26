# Farezero Demo e Talks

Questo repository è un monorepo atto a collezionare (in opportune sottocartelle) il materiale di talk e demo del gruppo Farezero.

## Aggiunta del materiale

Per aggiungere il materiale relativo ad un talk si puo' fare:
```bash
~/talks-and-demos (⌥ anybranch) git fetch -tpf
~/talks-and-demos (⌥ anybranch) git checkout main
~/talks-and-demos (⌥ main) git pull --rebase
~/talks-and-demos (⌥ main) mkdir argomento_del_talk
```
Per i talk del Linux Day usare invece `linux_day_AAAA/argomento_del_talk` (vedi ad esempio [linux_day_2026](linux_day_2026/README.md)).
Dopo aver copiato all'interno tutti i contenuti (slides in pdf, cartelle, codice sorgente, Dockerfile, docker-compose.yaml, .gitignore .dockerignore specifici, ecc.) provvedere a fare push:
```bash
~/talks-and-demos (⌥ main) git add .
~/talks-and-demos (⌥ main) git commit -s -m "add argomento_del_talk"
~/talks-and-demos (⌥ main) git push origin main
```
Essendo le cartelle tutte separate, possiamo fare push direttamente su `main` (trunk-based development). 
Nel caso in cui si voglia contribuire al codice o ai contenuti di proprietà di un'altra persona, invece, è consigliabile utilizzare un altro branch ed aprire una Pull Request verso il branch `main` chiedendo la revisione alla persona interessata.

### Se non fai parte dell'organizzazione

Il repository è pubblico, ma solo i membri dell'organizzazione GitHub `f4rezer0` possono fare push. Se non sei tra loro, usa un fork:

1. Fai il fork di questo repository dal pulsante "Fork" in alto a destra.
2. Aggiungi sul tuo fork la cartella del talk, come descritto sopra.
3. Apri una Pull Request verso il branch `main` di questo repository.

Un membro dell'organizzazione revisiona e integra la Pull Request.

Se non hai dimestichezza con git puoi fare tutto dal browser: entra nella cartella `linux_day_AAAA`, scegli "Add file" e poi "Upload files". GitHub crea il fork e ti propone di aprire la Pull Request. In alternativa mandaci le slide in PDF e le carichiamo noi.

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
├── linux_day_2024
│   └── k8s_autoscaling
│       ├── README.md
│       ├── metrics-server.yaml
│       ├── php-apache.yaml
│       └── slides.pdf
└── linux_day_2026
    ├── README.md
    └── argomento_del_talk
        └── ...
```

I materiali presentati al Linux Day sono raccolti in una cartella per edizione (`linux_day_AAAA`), con una sottocartella per ogni intervento.