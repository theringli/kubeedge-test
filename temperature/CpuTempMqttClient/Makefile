.PHONY: clean CpuTempMqttClient lint yamllint docker

CpuTempMqttClient: main.go kubeClient/kubeClient.go
	go build -o $@ $<

clean:
	go clean
	$(RM) main 
	$(RM) CpuTempMqttClient

lint:
	golangci-lint run ./...
	go vet ./...

YAML=$(wildcard *.yaml)
yamllint: $(YAML)
	yamllint $(YAML)
	
docker:
	$(MAKE) yamllint
	$(MAKE) lint
	docker build -t cpu_temp_mqtt_client -f Dockerfile .
