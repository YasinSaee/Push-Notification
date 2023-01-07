package notification

import (
	"github.com/YasinSaee/Push-Notification/config"
	"github.com/YasinSaee/Push-Notification/notif"
)

func InitNotification() {
	notif.Register(config.EC)
}
