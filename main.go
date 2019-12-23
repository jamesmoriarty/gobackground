package main

import (
	"io"
	"io/ioutil"
	"os"
	"net/http"
	"syscall"
	"unsafe"
	log "github.com/sirupsen/logrus"
)

var (
	url                  = "https://source.unsplash.com/collection/220388/1920x1080"
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW") // https://docs.microsoft.com/en-gb/windows/win32/api/winuser/nf-winuser-systemparametersinfow
	uiAction             = 0x0014 // SPI_SETDESKWALLPAPER 0x0014 - Note  When the SPI_SETDESKWALLPAPER flag is used, SystemParametersInfo returns TRUE unless there is an error (like when the specified file doesn't exist).
)

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	dir, err := ioutil.TempDir("", "gobackground")

	log.WithFields(log.Fields{"path": dir}).Info("Created directory")

	if err != nil {
		log.Fatal("Error: ", err)
        os.Exit(2)
	}

	path := dir + "\\background.jpg"

	log.WithFields(log.Fields{"url": url}).Info("Fetching")

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal("Error: ", err)
        os.Exit(2)
	}

	defer resp.Body.Close()

	log.WithFields(log.Fields{"path": path}).Info("Writing file")

	out, err := os.Create(path)

    if err != nil {
		log.Fatal("Error: ", err)
        os.Exit(2)
	}
	
    defer out.Close()

	_, err = io.Copy(out, resp.Body)
	
	if err != nil {
		log.Fatal("Error: ", err)
        os.Exit(2)
	}

	pvParam := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path)))

	log.WithFields(log.Fields{"uiAction": uiAction, "uiParam": 0, "pvParam": pvParam}).Info("SystemParametersInfoW")


	systemParametersInfo.Call(uintptr(uiAction), 0, pvParam, 2)
}