package main

import (
	"github.com/YasinSaee/Push-Notification/config"
	"github.com/YasinSaee/Push-Notification/notification"
)

func main() {
	notification.InitNotification()
	config.Run()

}
