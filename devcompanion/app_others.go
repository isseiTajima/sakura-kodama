//go:build !darwin
package main

func (a *App) setClickThroughNative(enabled bool) {}
func (a *App) setupNativeTray() {}
