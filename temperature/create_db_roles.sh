POD_NAME=$(kubectl get po -l app=timescale -o jsonpath='{.items[0].metadata.name}')

kubectl exec -ti ${POD_NAME} -- psql -h 127.0.0.1 -p 5432 -U postgres -c "CREATE ROLE kubeedge WITH LOGIN PASSWORD 'kubeedge'; CREATE ROLE grafana WITH LOGIN PASSWORD 'grafana';"

kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE grafana WITH OWNER grafana"

kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE demo WITH OWNER kubeedge"