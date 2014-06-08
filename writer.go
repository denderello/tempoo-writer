package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {

	c := createRedisConnection(":6379")
	defer c.Close()

	psc := redis.PubSubConn{c}
	psc.Subscribe("tempoo-update")

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("Received temperature update with value \"%s\"\n", v.Data)

			var update UpdateMessage
			json.Unmarshal(v.Data, &update)

			updateConn := createRedisConnection(":6379")
			addTemperature(updateConn, update.Temperature)
			updateConn.Close()
		case redis.Subscription:
			fmt.Println("Subscribed to temperature update channel")
		case error:
			fmt.Printf("There was an error when receiving the temperature update: %v\n", v)
			return
		}
	}

}

func addTemperature(c redis.Conn, temp int) {
	data, _ := json.Marshal(TemperatureRecord{temp, time.Now().Unix()})

	_, err := c.Do("LPUSH", "temperatures", data)

	if err != nil {
		log.Fatal(err)
	}

	_, err = c.Do("LTRIM", "temperatures", 0, 9)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("added temperature", temp)
}

func createRedisConnection(address string) redis.Conn {
	conn, err := redis.Dial("tcp", address)

	if err != nil {
		log.Fatal(err)
	}

	return conn
}

type UpdateMessage struct {
	Temperature int
}

type TemperatureRecord struct {
	Temperature int
	Date        int64
}
