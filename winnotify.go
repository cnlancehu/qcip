//go:build windows

package main

import (
	"gopkg.in/toast.v1"
)

func init() {
	// 系统为 windows 时，启用 winnotify 并显示对应的帮助信息

	notifyHelpMsg = "\n  -n, --winnotify	Send notifacation cards, only available on Windows"
	notify = func(title string, message string) {
		notification := toast.Notification{
			AppID:   "QCIP",
			Title:   title,
			Message: message,
		}
		err := notification.Push()
		if err != nil {
			errhandle("Error while sending notification: " + err.Error())
		}
	}
}
