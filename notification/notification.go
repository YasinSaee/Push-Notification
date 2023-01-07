package notification

import (
	"notification/config"
	"notification/notif"
)

func InitNotification() {
	notif.Register(config.EC)
}
