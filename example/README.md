# Пример выполнения

В данном примере будет описан запуск простого приложения для запроса текущего времени по HTTP запросу. Будет создан стандартный deployment, настроен доступ через loadbalancer, также будет настроен autoscaling.

## Часть 1 Запуск приложения

1) Сборка docker-образа приложения

Собираем docker образ приложения командой:

```bash
docker build . -t itmo-example-app
```

в результате должен появится образ приложения с именем "itmo-example-app", список образов можно посмотреть командой:

```bash
docker image ls
```

2) Загрузка собранного образа

Добавляем образ приложения в minikube командой

```bash
minikube image load itmo-example-app
```

посмотрим добавленный образ командой:

```bash
 minikube image ls | grep itmo-example-app:latest
```

 3) Поднятие приложения в kubernetes
 
Если вы пользуетесь minikube, то доступа к кластеру извне воспользуемся в отдельном терминале командой (нужно ввести пароль от системы):

```bash
minikube tunnel
```

Получим похожий вывод:

```shell
Status:
        machine: minikube
        pid: 2064187
        route: 10.96.0.0/12 -> 192.168.49.2
        minikube: Running
        services: [time-service-deployment]
    errors: 
                minikube: no errors
                router: no errors
                loadbalancer emulator: no errors
```

Создаём deployment командой:

 ```bash
kubectl apply -f kube/auto-scale-deployment.yml
 ```

Создадим балансировщик нагрузки и привяжем в нашему deployment:
```bash
kubectl expose deployment time-service-deployment --type=LoadBalancer --port=8080
```

Посмотрим внешний назначенный адрес (EXTERNAL-IP):
```bash
kubectl get svc
```

```shell
NAME                      TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)          AGE
kubernetes                ClusterIP      10.96.0.1        <none>           443/TCP          48m
time-service-deployment   LoadBalancer   10.110.153.206   10.110.153.206   8080:31031/TCP   43m
```

4) Проверка работоспособности приложения

Нам поднадобиться назначенный внешний адрес который мы смотрели в прошлом пункте, на него отправим обычный GET-запрос через curl:

```bash
curl http://<EXTERNAL-IP у time-service-deployment>:8080/time
```

Получим текущее время:

```shell
Current time is 13:10:45
```

5) Удаление ресурсов

Остановить minikube tunnel можно послав CTRL+C в терминал, где он запущен.

Для удаления сервиса выполним команду:

```
kubectl delete svc time-service-deployment
```

Для удаления deployment выполним команду:

```bash
kubectl delete -f kube/auto-scale-deployment.yml
```

## Часть 2: Автомасштабирование

Для начала поднимем другой deployment с указанными для подов ресурсами:

```bash
kubectl apply -f kube/auto-scale-deployment.yml
```
```bash
kubectl expose deployment time-service-deployment --type=LoadBalancer --port=8080
```

Если вы используете minikube, то нужно включить аддон metrics-server

```bash
minikube addons enable metrics-server
```

Также у minikube могут быть проблемы с работой metrics-server, желательно запустить minikube с флагом --extra-config=kubelet.housekeeping-interval=10s:

```bash
minikube start --extra-config=kubelet.housekeeping-interval=10s
```

Если пользуетесь кластером в vk cloud, то нужно установить metrics seriver:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

Настроим автомасштабирование. Здесь мы указываем имя deployment, который создали ранее и минимальное, максимальное количество подов которые могут быть созданы в ходе автомасштабирования, также указываем метрику и процент её достижения, при котором происходит автомасштабирование.

```bash
kubectl autoscale deployment time-service-deployment --cpu-percent=50 --min=2 --max=5
```

Посмотрим нагрузку на наш deployment:

```bash
kubectl get hpa time-service-deployment
```

Увидим примерно следующее:

```shell
NAME                      REFERENCE                            TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
time-service-deployment   Deployment/time-service-deployment  1%/50%   2         5         2          5m14s
```

Дальше подадим нагрузку на deployment для демонстрации работы автоскейлинга, при помощи образа с предустановленным wget. В него прокидываем shell script который будет посылать в бесконечном цикле запросы на наш сервис:

```bash
kubectl run load-generator --image=busybox -- /bin/sh -c "while true; do wget -q -O- http://<внешний адрес load balancer полученный ранее>:8080/time; done"
```

Спустя какое-то время нагрузка начнёт расти, снова посмотрим нагрузку на наш deployment:

```bash
kubectl get hpa time-service-deployment
```

увидим примерно следующее:

```shell
NAME                      REFERENCE                            TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
time-service-deployment   Deployment/time-service-deployment   83%/50%   2         5         2          5m14s
```

Посмотрим на созданные поды:

```shell
kubectl get pods
```

Изначально в репликасете было 2 пода, были созданы дополнительные:

```shell
NAME                                       READY   STATUS    RESTARTS   AGE
time-service-deployment-65bb69c589-2n69n   1/1     Running   0          27m
time-service-deployment-65bb69c589-ml56p   1/1     Running   0          23m
time-service-deployment-65bb69c589-qklkk   1/1     Running   0          27m
```

Под load-generator можно просто удалить по завершению тестирования:

```bash
kubectl delete pod load-generator
```

После удаления источника нагрузки на наш deployment, нагрузка должна упасть, а дополнительно созданные поды, удалятся.
