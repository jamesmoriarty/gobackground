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
	"unicode/utf16"
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

func errstr(errno int32) string {
	// ask windows for the remaining errors
	var flags uint32 = syscall.FORMAT_MESSAGE_FROM_SYSTEM | syscall.FORMAT_MESSAGE_ARGUMENT_ARRAY | syscall.FORMAT_MESSAGE_IGNORE_INSERTS
	b := make([]uint16, 300)
	n, err := syscall.FormatMessage(flags, 0, uint32(errno), 0, b, nil)
	if err != nil {
		return fmt.Sprintf("error %d (FormatMessage failed with: %v)", errno, err)
	}
	// trim terminating \r and \n
	for ; n > 0 && (b[n-1] == '\n' || b[n-1] == '\r'); n-- {
	}
	return string(utf16.Decode(b[:n]))
}

func setRegString(dir string, key string, value string) error {
	var handle win.HKEY

	log.WithFields(log.Fields{"dll": "advapi32"}).Info("RegOpenKeyEx")

	ret := win.RegOpenKeyEx(win.HKEY_CURRENT_USER, syscall.StringToUTF16Ptr(dir), 0, syscall.KEY_WRITE, &handle)

	if ret != 0 {
		return errors.New(fmt.Sprintf("Failed Opening Registry Key: %s", errstr(ret)))
	}

	log.WithFields(log.Fields{"dll": "advapi32"}).Info("RegSetValueEx")

	ret = win.RegSetValueEx(handle, syscall.StringToUTF16Ptr(key), 0, win.REG_SZ, (*byte)(unsafe.Pointer(syscall.StringToUTF16Ptr(value))), 32)

	if ret != 0 {
		return errors.New(fmt.Sprintf("Failed Setting Registry Key Value: %s", errstr(ret)))
	}

	return nil
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

	log.WithFields(log.Fields{"dll": "user32"}).Info("SystemParametersInfoW")

	ret := win.SystemParametersInfo(uiAction, uiParam, pvParam, fWinIni)

	if ret != true {
		return errors.New("Failed setting Desktop Wallpaper")
	}

	return nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	path, err := getRandomDesktopWallpaperPath()

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	err = setDesktopWallpaper(path)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	err = setRegString("Control Panel\\Desktop", "WallpaperStyle", "10")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
