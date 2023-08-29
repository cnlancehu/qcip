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
	notifyerrhandle := func() {
		if err != nil {
			err = nil
			erroroccurred = true
		}
	}
	notify = func(title string, message string, succeed bool) {
		var tmpFile *os.File
		tmpFile, err = os.CreateTemp("", "qcip-*.ico")
		notifyerrhandle()
		if succeed {
			tmpFile.Write(successicon)
		} else {
			tmpFile.Write(failedicon)
		}
		err = tmpFile.Close()
		notifyerrhandle()
		iconpath := tmpFile.Name()
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
		notifyerrhandle()
		if erroroccurred {
			errhandle("Error occurred when sending notification cards")
		}
	}
}
