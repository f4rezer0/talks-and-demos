# Scalabilità  e Resilienza in Kubernetes: come le app Cloud-Native si auto adattano al carico e rinascono dalle ceneri

## Premessa

## Cosa faremo

## Creazione del cluster


## Installazione del `metrics-server`

Installiamo il metrics-server, indispensabile per il funzionamento dell'horizontal pod autoscaler:
```bash
~  $ k apply -f metrics-server.yaml
```

## Installazione dei componenti

Installiamo tutti i componenti:
```bash
~  $ k apply -f php-apache.yaml
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

## Test di recovery

Proviamo ad cancellare il pod
```bash
~  $ k delete pods/...
```
... il replicaset

