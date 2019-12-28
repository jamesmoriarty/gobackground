# Notes

```
> on codeproject that points out an undocumented window message that spawns a
> * window behind the desktop icons. this is supposedly used to display the
> * animation when switching backgrounds
```

https://github.com/Francesco149/weebp/blob/master/src/weebp.c#L137

```
> uint WM_SPAWN_WORKER = 0x052C;
> SendMessage(Progman, WM_SPAWN_WORKER, (IntPtr) 0x0000000D, (IntPtr) 0);
> SendMessage(Progman, WM_SPAWN_WORKER, (IntPtr) 0x0000000D, (IntPtr) 1);
```

https://www.codeproject.com/Articles/856020/Draw-Behind-Desktop-Icons-in-Windows-plus?msg=5478543#xx5478543xx

```
/ <summary>
/ Special hack from https://www.codeproject.com/Articles/856020/Draw-behind-Desktop-Icons-in-Windows
/ Send 0x052C to Progman. This message directs Progman to spawn a
/ WorkerW behind the desktop icons. If it is already there, nothing
/ happens.
/ </summary>
public static void ShowAlwaysOnDesktopBehindIcons(IntPtr hwnd)
{
    var progmanHandle = FindWindowEx(IntPtr.Zero, IntPtr.Zero, "Progman", null);
    SendMessage(progmanHandle, 0x052C, 0x0000000D, 0);
    SendMessage(progmanHandle, 0x052C, 0x0000000D, 1);

    var workerWHandle = IntPtr.Zero;
    EnumWindows(new EnumWindowsProc((topHandle, topParamHandle) =>
    {
        IntPtr shellHandle = FindWindowEx(topHandle, IntPtr.Zero, "SHELLDLL_DefView", null);
        if (shellHandle != IntPtr.Zero)
        {
            workerWHandle = FindWindowEx(IntPtr.Zero, topHandle, "WorkerW", null);
        }
        return true;
    }), IntPtr.Zero);
    workerWHandle = workerWHandle == IntPtr.Zero ? progmanHandle : workerWHandle;
    SetParent(hwnd, workerWHandle);
}
```

https://github.com/AlexanderPro/AwesomeWallpaper/blob/master/AwesomeWallpaper/Utils/WindowUtils.cs#L16