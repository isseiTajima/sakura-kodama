//go:build darwin
package main

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdbool.h>
#include <stdlib.h>

void SetupNativeTray(const char* iconPath);
void SetWindowClickThroughNative(bool ignore);
*/
import "C"
import (
	"os"
	"path/filepath"
	"unsafe"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var globalAppInstance *App

// export は C から呼び出せるようにするために必要
//export goOnTraySettingsClicked
func goOnTraySettingsClicked() {
	if globalAppInstance != nil && globalAppInstance.ctx != nil {
		go func() {
			runtime.EventsEmit(globalAppInstance.ctx, "open-settings")
			runtime.WindowShow(globalAppInstance.ctx)
		}()
	}
}

//export goOnTrayQuitClicked
func goOnTrayQuitClicked() {
	if globalAppInstance != nil && globalAppInstance.ctx != nil {
		go func() {
			runtime.Quit(globalAppInstance.ctx)
		}()
	}
}

func (a *App) setClickThroughNative(enabled bool) {
	C.SetWindowClickThroughNative(C.bool(enabled))
}

func (a *App) setupNativeTray() {
	globalAppInstance = a
	
	// アイコン画像の絶対パスを取得
	cwd, _ := os.Getwd()
	iconPath := filepath.Join(cwd, "frontend/src/assets/chara.png")
	
	cPath := C.CString(iconPath)
	defer C.free(unsafe.Pointer(cPath))
	
	C.SetupNativeTray(cPath)
}
