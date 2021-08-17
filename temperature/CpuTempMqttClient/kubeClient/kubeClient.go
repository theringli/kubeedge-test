// Package kubeClient provide the functionality to initialize a MQTT connection and update the device twin
// in a feature version this package should also provide to handle incoming message via a callback function
package kubeClient

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Token interface {
	Wait() bool
	WaitTimeout(time.Duration) bool
	Error() error
}

//DeviceStateUpdate is the structure used in updating the device state
type DeviceStateUpdate struct {
	State string `json:"state,omitempty"`
}

//BaseMessage the base struct of event message
type BaseMessage struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

//TwinValue the struct of twin value
type TwinValue struct {
	Value    *string        `json:"value,omitempty"`
	Metadata *ValueMetadata `json:"metadata,omitempty"`
}

//ValueMetadata the meta of value
type ValueMetadata struct {
	Timestamp int64 `json:"timestamp,omitempty"`
}

//TypeMetadata the meta of value type
type TypeMetadata struct {
	Type string `json:"type,omitempty"`
}

//TwinVersion twin version
type TwinVersion struct {
	CloudVersion int64 `json:"cloud"`
	EdgeVersion  int64 `json:"edge"`
}

// MsgTwin the structe of device twin
type MsgTwin struct {
	Actual          *TwinValue    `json:"actual,omitempty"`
	Expected        *TwinValue    `json:"expected,omitempty"`
	Optional        *bool         `json:"optional,omitempty"`
	Metadata        *TypeMetadata `json:"metadata,omitempty"`
	ExpectedVersion *TwinValue    `json:"expected_version,omitempty"`
	ActualVersion   *TwinVersion  `json:"actual_version,omitempty"`
}

// DeviceTwinUdpdate the struct of device twin update
type DeviceTwinUpdate struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

type DeviceTwinUpdateDelta struct {
	BaseMessage
	Twin  map[string]*MsgTwin `json:"twin"`
	Delta map[string]string   `json:"delta"`
}

var (
	Prefix                = "$hw/events/device/"
	StateUpdateSuffix     = "/state/update"
	TwinUpdateSuffix      = "/twin/update"
	TwinCloudUpdateSuffix = "/twin/cloud_updated"
	TwinGetResultSuffix   = "/twin/get/result"
	TwinGetSuffix         = "/twin/get"
)

var token_client Token
var clientOpts *MQTT.ClientOptions
var client MQTT.Client
var deviceID string
var eventID int
var cpu_id string

// mqttConfig creates the mqtt client config
func mqttConfig(server, clientID, user, password string) *MQTT.ClientOptions {
	options := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientID).SetCleanSession(true)
	if user != "" {
		options.SetUsername(user)
		if password != "" {
			options.SetPassword(password)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	options.SetTLSConfig(tlsConfig)
	return options
}

// changeSensorStatus function is used to change the state of this sensor
func changeSensorStatus(state string) {
	log.Println("Changing the state of the device to ", state)
	var sensorStateUpdate DeviceStateUpdate
	sensorStateUpdate.State = state
	messageBody, err := json.Marshal(sensorStateUpdate)
	if err != nil {
		log.Panicln(err)
	}
	statusUpdate := Prefix + deviceID + StateUpdateSuffix
	token_client = client.Publish(statusUpdate, 0, false, messageBody)
	if token_client.Wait() && token_client.Error() != nil {
		log.Panicln("client.publish() Error in sensor state update is: ", token_client.Error())
	}
}

// changeTwinValue sends the updated twin value to the edge through the MQTT broker
func changeTwinValue(updateMessage DeviceTwinUpdate) {
	messageBody, err := json.Marshal(updateMessage)
	if err != nil {
		log.Println("Error: ", err)
	}
	topic := Prefix + deviceID + TwinUpdateSuffix
	token_client = client.Publish(topic, 0, false, messageBody)
	if token_client.Wait() && token_client.Error() != nil {
		log.Println("client.publish() Error in device twin update is: ", token_client.Error())
	}
}

// syncToCloud function syncs the updated device twin information to the cloud
func syncToCloud(message DeviceTwinUpdate) {
	topic := Prefix + deviceID + TwinCloudUpdateSuffix
	messageBody, err := json.Marshal(message)
	if err != nil {
		log.Println("syncToCoud marshal error is: ", err)
	}
	token_client = client.Publish(topic, 0, false, messageBody)
	if token_client.Wait() && token_client.Error() != nil {
		log.Println("client.publish() erro in device twin update to cloud is: ", token_client.Error())
	}
}

// createActualUpdateMessage function is used to create the device twin update message
func createActualUpdateMessage(actualValue string) DeviceTwinUpdate {
	var message DeviceTwinUpdate
	actualMap := map[string]*MsgTwin{
		"CPU_Temperatur": {Actual: &TwinValue{Value: &actualValue}, Metadata: &TypeMetadata{Type: "int"}},
		"cpu_id":         {Actual: &TwinValue{Value: &cpu_id}, Metadata: &TypeMetadata{Type: "int"}},
	}
	message.Twin = actualMap
	message.Timestamp = time.Now().Unix()
	message.EventID = strconv.Itoa(eventID)
	eventID++
	return message
}

// Update this function is used to update values on the edge and in the cloud
func Update(value string) {
	log.Println("Syncing to edge")
	updateMessage := createActualUpdateMessage(value)
	changeTwinValue(updateMessage)
	time.Sleep(2 * time.Second)
	log.Println("Syncing to cloud")
	syncToCloud(updateMessage)
}

func handleMessage(client MQTT.Client, origMessage MQTT.Message) {
	var deltaUpdate DeviceTwinUpdateDelta
	err := json.Unmarshal(origMessage.Payload(), &deltaUpdate)
	if err != nil {
		log.Printf("can not unmarshal receive message error is: %v", err)
		return
	}

	if deltaUpdate.Delta["cpu_id"] != "" {
		cpu_id = deltaUpdate.Delta["cpu_id"]
	}
	log.Printf("current delta is: %v", deltaUpdate.Delta)
}

func processSubscription(client MQTT.Client) {
	topic := Prefix + deviceID + "/twin/update/delta"
	token := client.Subscribe(topic, 0, handleMessage)
	if token.Error() != nil {
		log.Printf("Error in process subscription: %v", token.Error())
	}
}

// Init initialize the MQTT connection and set the used ipAddress and deviceID
// in a feature version this function also register the callback method to handle incoming messages
// ipAddress and deviceID has to be set! If you don't want to add an user or an password in the
// MQTT Connection set user and password to nil
func Init(ipAddress, id, user, password string) {
	deviceID = id
	eventID = 0
	cpu_id = "0"
	clientOpts = mqttConfig(ipAddress, deviceID, user, password)
	client = MQTT.NewClient(clientOpts)
	if token_client = client.Connect(); token_client.Wait() && token_client.Error() != nil {
		log.Println("client.Connect() Error is: ", token_client.Error())
	}
	go processSubscription(client)
	changeSensorStatus("online")
}
