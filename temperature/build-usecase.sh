#!/bin/bash
case $1 in
    build)
        kubectl apply -f timescale-db.yml
        while [ $(kubectl get pods | grep timescale | awk '{ print $3;}') != "Running" ]
        do
            sleep 5
        done
        POD_NAME=$(kubectl get po -l app=timescale -o jsonpath='{.items[0].metadata.name}')
        kubectl exec -ti ${POD_NAME} -- psql -h 127.0.0.1 -p 5432 -U postgres -c "CREATE ROLE kubeedge WITH LOGIN PASSWORD 'kubeedge'; CREATE ROLE grafana WITH LOGIN PASSWORD 'grafana';"
        kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE grafana WITH OWNER grafana"
        kubectl exec -ti ${POD_NAME} -- psql -U postgres -c "CREATE DATABASE demo WITH OWNER kubeedge"
        kubectl apply -f grafana.yml
        kubectl apply -f device/deviceModel.yaml
        kubectl apply -f device/deviceInstance.yaml
        kubectl apply -f kubeedge-database.yml
        kubectl apply -f cpu-temp-sensor.yaml
        ;;
    
    delete)
        kubectl delete -f timescale-db.yml
        kubectl delete -f grafana.yml
        kubectl delete -f device/deviceModel.yaml
        kubectl delete -f device/deviceInstance.yaml
        kubectl delete -f kubeedge-database.yml
        kubectl delete -f cpu-temp-sensor.yaml
        ;;

    rebuild)
        ./build-usecase.sh delete
        ./build-usecase.sh build
        ;;

    *)
        echo "Use arguments: build, delete, or rebuild"
        ;;
esac