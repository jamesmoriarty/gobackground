package main

// > on codeproject that points out an undocumented window message that spawns a
// > * window behind the desktop icons. this is supposedly used to display the
// > * animation when switching backgrounds
//
// https://github.com/Francesco149/weebp/blob/master/src/weebp.c#L137

// > uint WM_SPAWN_WORKER = 0x052C;
// > SendMessage(Progman, WM_SPAWN_WORKER, (IntPtr) 0x0000000D, (IntPtr) 0);
// > SendMessage(Progman, WM_SPAWN_WORKER, (IntPtr) 0x0000000D, (IntPtr) 1);
//
// https://www.codeproject.com/Articles/856020/Draw-Behind-Desktop-Icons-in-Windows-plus?msg=5478543#xx5478543xx

/// <summary>
/// Special hack from https://www.codeproject.com/Articles/856020/Draw-behind-Desktop-Icons-in-Windows
/// Send 0x052C to Progman. This message directs Progman to spawn a
/// WorkerW behind the desktop icons. If it is already there, nothing
/// happens.
/// </summary>
// public static void ShowAlwaysOnDesktopBehindIcons(IntPtr hwnd)
// {
//     var progmanHandle = FindWindowEx(IntPtr.Zero, IntPtr.Zero, "Progman", null);
//     SendMessage(progmanHandle, 0x052C, 0x0000000D, 0);
//     SendMessage(progmanHandle, 0x052C, 0x0000000D, 1);

//     var workerWHandle = IntPtr.Zero;
//     EnumWindows(new EnumWindowsProc((topHandle, topParamHandle) =>
//     {
//         IntPtr shellHandle = FindWindowEx(topHandle, IntPtr.Zero, "SHELLDLL_DefView", null);
//         if (shellHandle != IntPtr.Zero)
//         {
//             workerWHandle = FindWindowEx(IntPtr.Zero, topHandle, "WorkerW", null);
//         }
//         return true;
//     }), IntPtr.Zero);
//     workerWHandle = workerWHandle == IntPtr.Zero ? progmanHandle : workerWHandle;
//     SetParent(hwnd, workerWHandle);
// }
//
// https://github.com/AlexanderPro/AwesomeWallpaper/blob/master/AwesomeWallpaper/Utils/WindowUtils.cs#L16

import (
	"errors"
	"fmt"
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
	URL = "https://source.unsplash.com/collection/220388/1920x1080"

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

	log.WithFields(log.Fields{"path": dir}).Info("Created directory")

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	path := dir + "\\Pictures\\" + fmt.Sprintf("%d", time.Now().UnixNano()) + ".jpg"

	log.WithFields(log.Fields{"url": URL}).Info("Fetching")

	resp, err := http.Get(URL)

	if err != nil {
		log.Fatal("Error: ", err)
		return "", err
	}

	defer resp.Body.Close()

	log.WithFields(log.Fields{"path": path}).Info("Writing file")

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
	pvParam := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path)))
	uiParam := uintptr(0)
	uiAction := uintptr(SPI_SETDESKWALLPAPER)
	fWinIni := uintptr(SPIF_SENDWININICHANGE)
	user32 := syscall.NewLazyDLL("user32.dll")
	systemParametersInfo := user32.NewProc("SystemParametersInfoW") // https://docs.microsoft.com/en-gb/windows/win32/api/winuser/nf-winuser-systemparametersinfow

	log.WithFields(log.Fields{"uiAction": toHex(uiAction), "uiParam": toHex(0), "pvParam": toHex(pvParam), "fWinIni": toHex(fWinIni)}).Info("SystemParametersInfoW")

	ret, _, _ := systemParametersInfo.Call(uiAction, uiParam, pvParam, fWinIni)

	if ret == 0 {
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
