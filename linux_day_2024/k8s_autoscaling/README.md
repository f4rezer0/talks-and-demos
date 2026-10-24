# Scalabilità  e Resilienza in Kubernetes: come le app Cloud-Native si auto adattano al carico e rinascono dalle ceneri

## Premessa

Questa Demo è volta a dimostrare la funzionalità dell'`horizontal pod autoscaler` e della `failure recovery` dei cluster Kubernetes (K8s).

## Cosa faremo

Creeremo un cluster Kubernetes con `apache` http server e lo porremo sotto stress per testare le funzionalità di adattamento di Kubernetes al traffico rete ed eventuali crash dell'applicazione.

## Creazione del cluster

Di seguito le alternative: creiamo un cluster kubernetes in locale usando k3d (k3s in docker) oppure su un vero cloud provider (civo.com)

### Creazione del cluster in locale (docker) con k3d

Installiamo `k3d`:
```bash
~  $ curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```
Creiamo il cluster:
```bash
~  $ k3d cluster create linuxday-k3d --agents 2 --port '80:80@loadbalancer' --port '443:443@loadbalancer'
```
Esportiamo il kubeconfig ed usiamolo:
```bash
~  $ k3d kubeconfig get linuxday-k3d > ~/.kube/config_linuxday-k3d
~  $ export KUBECONFIG=$HOME/.kube/config_linuxday-k3d
~  $ k get nodes
NAME                     STATUS   ROLES                  AGE    VERSION
k3d-linuxday-k3d-server-0   Ready    control-plane,master   107s   v1.27.4+k3s1
k3d-linuxday-k3d-agent-0    Ready    <none>                 103s   v1.27.4+k3s1
k3d-linuxday-k3d-agent-1    Ready    <none>                 103s   v1.27.4+k3s1
```
Ogni nodo è creato come un container docker:
```bash
$ docker ps | grep k3d
aeca25a461e5   ghcr.io/k3d-io/k3d-tools:5.6.0      "/app/k3d-tools noop"    12 minutes ago   Up 12 minutes                                                                       k3d-lday1-k3d-tools
7d0eeae38447   ghcr.io/k3d-io/k3d-proxy:5.6.0      "/bin/sh -c nginx-pr…"   12 minutes ago   Up 12 minutes   0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:58420->6443/tcp   k3d-lday1-k3d-serverlb
6d921f65e059   rancher/k3s:v1.27.4-k3s1            "/bin/k3d-entrypoint…"   12 minutes ago   Up 12 minutes                                                                       k3d-lday1-k3d-agent-1
aead2f2e57a1   rancher/k3s:v1.27.4-k3s1            "/bin/k3d-entrypoint…"   12 minutes ago   Up 12 minutes                                                                       k3d-lday1-k3d-agent-0
1e7c713453b6   rancher/k3s:v1.27.4-k3s1            "/bin/k3d-entrypoint…"   12 minutes ago   Up 12 minutes
```

### Creazione del cluster su un vero Cloud Provider (Civo)
Usiamo la CLI (https://github.com/civo/cli):
```bash
~  $ civo kubernetes create linuxday --cluster-type=talos --nodes=2 --region=lon1
The cluster linuxday (67d7e831-579a-4185-abe4-689ac698199f) has been created
```
Recuperiamo il `kubeconfig` e puntiamo al cluster:
```bash
~  $ civo kubernetes config linuxday --region=lon1 > ~/.kube/config_linuxday
~  $ export KUBECONFIG=~/.kube/config_linuxday
```

## Installazione del `metrics-server`

Aggiungiamo un alias per comodità e testiamo la connessione al cluster:
```bash
~  $ alias k='kubectl'
~  $ k get ns
NAME              STATUS   AGE
default           Active   2m26s
kube-node-lease   Active   2m26s
kube-public       Active   2m26s
kube-system       Active   2m26s
```

Installiamo il metrics-server, indispensabile per il funzionamento dell'horizontal pod autoscaler:
```bash
~  $ k apply -f metrics-server.yaml
```

## Installazione dei componenti

Installiamo tutti i componenti:
```bash
~  $ k apply -f php-apache.yaml
deployment.apps/php-apache-deployment created
horizontalpodautoscaler.autoscaling/php-apache-hpa created
service/php-apache-svc created
```

## Test del `horizontal-pod-autoscaler`

Osserviamo (su due terminali diversi) in tempo reale i pod e l'horizontalpodautoscaler:
```bash
# terminale 1
~  $ k get pods --watch
# terminale 2
~  $ k get hpa --watch
```
Poi avviamo il carico scegliendo uno dei seguenti modi:

Utilizzando l'IP del Loadbalancer (se avviato tramite il Cloud Provider):
```bash
~  $ LB_IP=$(k get svc/php-apache-svc -o jsonpath={.status.loadBalancer.ingress[0].ip})
~  $ while sleep 0.01; do curl http://$LB_IP:80>/dev/null; done
```
utilizzando il `kubectl port-forward`:
```bash
~  $ k port-forward svc/php-apache-svc 8888:80
# su un altro terminale
~  $ while sleep 0.01; do curl http://localhost:8888>/dev/null; done 
```

utilizzando un pod nel cluster creato ad-hoc per effettuare le chiamate alla nostra app:
```bash
kubectl run -i --tty load-generator --rm --image=busybox:1.28 --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://php-apache-svc; done"
```

Tornando sui terminali precedenti notiamo che l'horizontal pod autoscaler ha aumentato il numero di repliche fino al massimo (10):
```bash
~  $ k get hpa --watch
NAME             REFERENCE                          TARGETS    MINPODS   MAXPODS   REPLICAS   AGE
php-apache-hpa   Deployment/php-apache-deployment   250%/10%   1         10        1          24m
php-apache-hpa   Deployment/php-apache-deployment   250%/10%   1         10        4          25m
php-apache-hpa   Deployment/php-apache-deployment   117%/10%   1         10        8          25m
php-apache-hpa   Deployment/php-apache-deployment   36%/10%    1         10        10         25m
php-apache-hpa   Deployment/php-apache-deployment   41%/10%    1         10        10         25m
php-apache-hpa   Deployment/php-apache-deployment   21%/10%    1         10        10         26m
php-apache-hpa   Deployment/php-apache-deployment   22%/10%    1         10        10         26m
php-apache-hpa   Deployment/php-apache-deployment   27%/10%    1         10        10         26m
```
ed il numero delle repliche è effettivamente salito a 10:
```bash
~  $ k get pods
NAME                                     READY   STATUS    RESTARTS   AGE
php-apache-deployment-5bdbb8dbf8-c7d92   1/1     Running   0          28m
load-generator                           1/1     Running   0          5m9s
php-apache-deployment-5bdbb8dbf8-mb2nx   1/1     Running   0          3m43s
php-apache-deployment-5bdbb8dbf8-x5f7p   1/1     Running   0          3m28s
php-apache-deployment-5bdbb8dbf8-l4nvq   1/1     Running   0          3m28s
php-apache-deployment-5bdbb8dbf8-bv9nv   1/1     Running   0          3m43s
php-apache-deployment-5bdbb8dbf8-j2mcx   1/1     Running   0          3m43s
php-apache-deployment-5bdbb8dbf8-p5vsz   1/1     Running   0          3m28s
php-apache-deployment-5bdbb8dbf8-8nxpn   1/1     Running   0          3m28s
php-apache-deployment-5bdbb8dbf8-tr7m9   1/1     Running   0          3m13s
php-apache-deployment-5bdbb8dbf8-hdbtr   1/1     Running   0          3m13s
```
Se stoppiamo con Ctrl+C il processo che genera il carico, dopo un certo tempo, i pod vengono nuovamente scalati ad 1 replica:
```bash
$ k get hpa -w
NAME             REFERENCE                          TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
php-apache-hpa   Deployment/php-apache-deployment   3%/50%    1         10        10         38m
php-apache-hpa   Deployment/php-apache-deployment   0%/50%    1         10        10         38m
php-apache-hpa   Deployment/php-apache-deployment   0%/50%    1         10        10         39m
php-apache-hpa   Deployment/php-apache-deployment   0%/50%    1         10        1          39m
```
## Test di recovery

Proviamo ad cancellare i pod:
```bash
~  $ kubectl get pods --no-headers | awk '{print $1}' | xargs -I {} kubectl delete pod {}
```
e vediamo come questi sono ricreati:
```bash
~ $ k get pods
```

