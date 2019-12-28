package main

import (
	"errors"
	"fmt"
	win "github.com/lxn/win"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const (
	URL = "https://source.unsplash.com/%dx%d/?backgrounds,desktop,computer"

	// fWinIni
	SPIF_UPDATEINIFILE    = 0x0
	SPIF_SENDCHANGE       = 0x1
	SPIF_SENDWININICHANGE = 0x2

	// uiAction
	SPI_SETDESKWALLPAPER = 0x0014
)

func toHex(ptr uintptr) string {
	s := fmt.Sprintf("%d", ptr)
	n, _ := strconv.Atoi(s)
	h := fmt.Sprintf("0x%x", n)
	return h
}

func getRandomDesktopWallpaperPath() (string, error) {
	dir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	path := fmt.Sprintf("%s\\Pictures\\%d.jpg", dir, time.Now().UnixNano())

	width := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	height := int(win.GetSystemMetrics(win.SM_CYSCREEN))

	log.WithFields(log.Fields{"width": width, "height": height}).Info("GetSystemMetrics")

	url := fmt.Sprintf(URL, width, height)

	log.WithFields(log.Fields{"url": url}).Info("Fetching")

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	defer resp.Body.Close()

	log.WithFields(log.Fields{"path": path}).Info("Writing")

	out, err := os.Create(path)

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	return path, nil
}

func setDesktopWallpaper(path string) error {
	pvParam := unsafe.Pointer(syscall.StringToUTF16Ptr(path))
	uiParam := uint32(0)
	uiAction := uint32(SPI_SETDESKWALLPAPER)
	fWinIni := uint32(SPIF_SENDWININICHANGE)

	log.WithFields(log.Fields{"uiAction": uiAction, "uiParam": 0, "pvParam": toHex(uintptr(pvParam)), "fWinIni": fWinIni}).Info("SystemParametersInfoW")

	ret := win.SystemParametersInfo(uiAction, uiParam, pvParam, fWinIni)

	if ret != true {
		return errors.New("Failed setting desktopwallpaper")
	}

	return nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	path, err := getRandomDesktopWallpaperPath()

	if err != nil {
		os.Exit(2)
	}

	err = setDesktopWallpaper(path)

	if err != nil {
		os.Exit(2)
	}
}
