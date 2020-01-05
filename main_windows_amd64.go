package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"

	win "github.com/lxn/win"
	log "github.com/sirupsen/logrus"
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

func scale() float64 {
	log.WithFields(log.Fields{"dll": "winuser"}).Info("FindWindow")

	handle := win.FindWindow(syscall.StringToUTF16Ptr("Progman"), syscall.StringToUTF16Ptr("Program Manager"))
	dpi := win.GetDpiForWindow(handle)
	scale := (float64(dpi) / float64(96))

	log.WithFields(log.Fields{"dll": "winuser", "dpi": dpi, "scale": scale}).Info("GetDpiForWindow")

	return scale
}

func width() int {
	log.WithFields(log.Fields{"dll": "winuser"}).Info("GetSystemMetrics")

	return int(float64(win.GetSystemMetrics(win.SM_CXSCREEN)) * scale())
}

func height() int {
	log.WithFields(log.Fields{"dll": "winuser"}).Info("GetSystemMetrics")

	return int(float64(win.GetSystemMetrics(win.SM_CYSCREEN)) * scale())
}

func getPathFromResp(resp *http.Response) string {
	url, err := url.Parse(resp.Request.URL.String())

	if err != nil {
		panic("Unexpected url")
	}

	filename := filepath.Base(url.Path)

	switch resp.Header.Get("Content-Type") {
	case "image/jpeg":
		return fmt.Sprintf("%s\\%s.jpg", url.Hostname(), filename)
	default:
		panic("Unexpected content type")
	}
}

func getFilePathFromURL(url string) (string, error) {
	dir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)

	log.WithFields(log.Fields{"url": url}).Info("Fetching")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	path := fmt.Sprintf("%s\\Downloads\\%s", dir, getPathFromResp(resp))

	log.WithFields(log.Fields{"path": path}).Info("Writing")

	os.MkdirAll(filepath.Dir(path), os.ModePerm)

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

func getURL() string {
	return fmt.Sprintf(urlTemplate, width(), height())
}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	path, err := getFilePathFromURL(getURL())

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
