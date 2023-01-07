package main

import (
	"notification/config"
	"notification/notification"
)

func main() {
	notification.InitNotification()
	config.Run()

}
