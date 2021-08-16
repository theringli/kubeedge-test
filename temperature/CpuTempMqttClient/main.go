package main

import (
	"bytes"
	"flag"
	"log"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/subpathdev/CpuTempMqttClient/kubeClient"
)

var mqttURL, deviceID, user, password string
var simulate bool

func init() {
	//	flag := flag.NewFlagSet("Usage", flag.ExitOnError)
	flag.StringVar(&mqttURL, "mqttURL", "tcp://127.0.0.1:1883", "URL to the MQTT Broker")
	flag.StringVar(&deviceID, "deviceID", "cpu-sensor-tag01", "The unique ID of this device (created in the cloud)")
	flag.StringVar(&user, "user", "", "User to connect to the MQTT broker")
	flag.StringVar(&password, "password", "", "Password for the MQTT Broker")
	flag.BoolVar(&simulate, "simulate", true, "if you use this flag the input data will be simulated by random numbers and no sensor will be requested")
}

func main() {
	flag.Parse()
	kubeClient.Init(mqttURL, deviceID, user, password)
	for {
		var message string
		var out bytes.Buffer
		var core0 = false

		if simulate {
			message = strconv.Itoa(rand.Intn(200))
		} else {
			cmd := exec.Command("/usr/bin/sensors", "-Au")
			cmd.Stdout = &out
			cmd.Stderr = &out
			err := cmd.Start()
			if err != nil {
				log.Println("Error in command execution. Error: ", err)
			}
			err = cmd.Wait()
			if err != nil {
				log.Println("Error by waiting on command execution. Error: ", err)
			}
			str := strings.Split(out.String(), "\n")

			for _, element := range str {
				if core0 {
					if strings.Contains(element, "temp2_input") {
						val := strings.Split(element, ": ")
						message = val[1]
						core0 = false
					}
				}

				if strings.Contains(element, "Core 0:") {
					core0 = true
				}

			}
		}

		kubeClient.Update(message)
		time.Sleep(10 * time.Second)
	}
}
