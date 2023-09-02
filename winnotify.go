//go:build windows

package main

import (
	_ "embed"
	"os"

	"gopkg.in/toast.v1"
)

var (
	//go:embed static/success.ico
	successicon []byte
	//go:embed static/failed.ico
	failedicon []byte
	err        error = nil
)

func init() {
	// 系统为 windows 时，启用 winnotify 并显示对应的帮助信息
	var erroroccurred bool = false
	notifyHelpMsg = "\n  -n, --winnotify	Send notifacation cards, only available on Windows"
	notifyerrcheck := func() {
		if err != nil {
			err = nil
			erroroccurred = true
		}
	}
	notify = func(title string, message string, succeed bool) {
		var iconFile *os.File
		iconFile, err = os.CreateTemp("", "qcip-*.ico")
		notifyerrcheck()
		if succeed {
			iconFile.Write(successicon)
		} else {
			iconFile.Write(failedicon)
		}
		err = iconFile.Close()
		notifyerrcheck()
		iconpath := iconFile.Name()
		var notification toast.Notification
		if !erroroccurred {
			notification = toast.Notification{
				AppID:   "QCIP",
				Title:   title,
				Message: message,
				Icon:    iconpath,
			}
		} else {
			notification = toast.Notification{
				AppID:   "QCIP",
				Title:   title,
				Message: message,
			}
		}
		err = notification.Push()
		notifyerrcheck()
		if erroroccurred {
			errhandle("Error occurred when sending notification cards")
		}
	}
}
