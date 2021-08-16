<h2>Usecase<h2>

https://www.inovex.de/de/blog/kubeedge-use-case-vom-sensor-bis-zur-visualisierung-in-grafana/


1. Run _kubectl apply -f timescale-db.yml_ and follow the steps to create Database:

POD_NAME=$(kubectl get po -l app=timescale -o jsonpath='{.items[0].metadata.name}')

kubectl exec -ti ${POD_NAME} -- psql -h 127.0.0.1 -p 5432 -U postgres -c "CREATE ROLE kubeedge WITH LOGIN PASSWORD 'kubeedge'; CREATE ROLE grafana WITH LOGIN PASSWORD 'grafana';"

kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE grafana WITH OWNER grafana"

kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE demo WITH OWNER kubeedge"

2. Run _kubectl apply -f kubeedge-database.yml_

3. Finaly _kubectl apply -f grafana.yml_

4. Check if all pods are running _kubectl get pods_ and access grafana x.x.x.x:30975 <br>
 _(User: admin Passwort: verryStrongPassword)_

5. Configure Grafana: Add data source -> select PostgresDB<br>

(Konfiguration = Wert)<br>
Name = Timescale-kubeedge<br>
Host = timescale.default.svc.cluster.local<br>
Database = demo<br>
User = kubeedge<br>
Passwort = kubeedge<br>
SSL-Mode = disable<br>

Add dashboard with _Devices.json_ and _Sensors.json_ (left side press plus and import)
