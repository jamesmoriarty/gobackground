package main

import (
	"errors"
	"fmt"
	win "github.com/lxn/win"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	// Unspashed Search
	urlTemplate = "https://source.unsplash.com/%dx%d/?backgrounds,desktop,computer"

	// Registry Wallpaper
	wallpaperDir = "Control Panel\\Desktop"
	wallpaperKey = "Wallpaper"

	// Registry WallpaperStyle
	wallpaperStyleDir   = "Control Panel\\Desktop"
	wallpaperStyleKey   = "WallpaperStyle"
	wallpaperStyleValue = "10"

	// SystemParametersInfo fWinIni
	spifUpdateIniFile    = uint32(0x1)
	spifSendWinIniChange = uint32(0x2)

	// SystemParametersInfo uiAction
	spiSetDesktopWallpaper = uint32(0x0014)
)

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

func setRegistryValue(dir string, key string, value string) error {
	var handle win.HKEY

	log.WithFields(log.Fields{"dll": "advapi32"}).Info("RegOpenKeyEx")

	ret := win.RegOpenKeyEx(win.HKEY_CURRENT_USER, syscall.StringToUTF16Ptr(dir), 0, syscall.KEY_WRITE, &handle)

	if ret != 0 {
		return fmt.Errorf("Failed Opening Registry Key: %s", errstr(ret))
	}

	log.WithFields(log.Fields{"dll": "advapi32"}).Info("RegSetValueEx")

	ret = win.RegSetValueEx(handle, syscall.StringToUTF16Ptr(key), 0, win.REG_SZ, (*byte)(unsafe.Pointer(syscall.StringToUTF16Ptr(value))), uint32(len(value)*2))

	if ret != 0 {
		return fmt.Errorf("Failed Setting Registry Key Value: %s", errstr(ret))
	}

	return nil
}

func width() int {
	log.WithFields(log.Fields{"dll": "winuser"}).Info("GetSystemMetrics")

	return int(win.GetSystemMetrics(win.SM_CXSCREEN))
}

func height() int {
	log.WithFields(log.Fields{"dll": "winuser"}).Info("GetSystemMetrics")

	return int(win.GetSystemMetrics(win.SM_CYSCREEN))
}

func getRandomDesktopWallpaperPath(url string) (string, error) {
	dir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s\\Pictures\\%d.jpg", dir, time.Now().UnixNano())

	log.WithFields(log.Fields{"url": url}).Info("Fetching")

	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	log.WithFields(log.Fields{"path": path}).Info("Writing")

	out, err := os.Create(path)

	if err != nil {
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return "", err
	}

	return path, nil
}

func setDesktopWallpaper(path string) error {
	pvParam := unsafe.Pointer(syscall.StringToUTF16Ptr(path))
	uiParam := uint32(0)
	uiAction := spiSetDesktopWallpaper
	fWinIni := spifSendWinIniChange

	log.WithFields(log.Fields{"dll": "winuser"}).Info("SystemParametersInfoW")

	ret := win.SystemParametersInfo(uiAction, uiParam, pvParam, fWinIni)

	if ret != true {
		return errors.New("Failed setting Desktop Wallpaper")
	}

	return nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	url := fmt.Sprintf(urlTemplate, width(), height())

	path, err := getRandomDesktopWallpaperPath(url)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	err = setDesktopWallpaper(path)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	err = setRegistryValue(wallpaperDir, wallpaperKey, path)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	err = setRegistryValue(wallpaperStyleDir, wallpaperStyleKey, wallpaperStyleValue)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
