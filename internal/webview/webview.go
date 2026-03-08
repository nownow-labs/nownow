//go:build windows || darwin

// Package webview provides a Go wrapper around the webview C library.
// Vendored from github.com/ollama/ollama (MIT License).
package webview

/*
#cgo CXXFLAGS: -DWEBVIEW_STATIC

#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11
#cgo darwin LDFLAGS: -framework WebKit -ldl

#cgo windows CXXFLAGS: -DWEBVIEW_EDGE -std=c++14
#cgo windows LDFLAGS: -static -ladvapi32 -lole32 -lshell32 -lshlwapi -luser32 -lversion

#include "webview.h"

#include <stdlib.h>
#include <stdint.h>

void CgoWebViewDispatch(webview_t w, uintptr_t arg);
*/
import "C"

import (
	"runtime"
	"sync"
	"unsafe"
)

func init() {
	runtime.LockOSThread()
}

// Hint configures window sizing.
type Hint int

const (
	HintNone  Hint = C.WEBVIEW_HINT_NONE
	HintFixed Hint = C.WEBVIEW_HINT_FIXED
	HintMin   Hint = C.WEBVIEW_HINT_MIN
	HintMax   Hint = C.WEBVIEW_HINT_MAX
)

// WebView is the interface for a webview window.
type WebView interface {
	Run()
	Terminate()
	Dispatch(f func())
	Destroy()
	Window() unsafe.Pointer
	SetTitle(title string)
	SetSize(w int, h int, hint Hint)
	Navigate(url string)
	SetHtml(html string)
	Init(js string)
	Eval(js string)
}

type webview struct {
	w C.webview_t
}

var (
	m        sync.Mutex
	index    uintptr
	dispatch = map[uintptr]func(){}
)

// New creates a new webview instance. Returns nil if creation fails.
func New(debug bool) WebView {
	w := &webview{}
	var d C.int
	if debug {
		d = 1
	}
	w.w = C.webview_create(d, nil)
	if w.w == nil {
		return nil
	}
	return w
}

func (w *webview) Destroy()               { C.webview_destroy(w.w) }
func (w *webview) Run()                    { C.webview_run(w.w) }
func (w *webview) Terminate()              { C.webview_terminate(w.w) }
func (w *webview) Window() unsafe.Pointer  { return C.webview_get_window(w.w) }

func (w *webview) Navigate(url string) {
	s := C.CString(url)
	defer C.free(unsafe.Pointer(s))
	C.webview_navigate(w.w, s)
}

func (w *webview) SetHtml(html string) {
	s := C.CString(html)
	defer C.free(unsafe.Pointer(s))
	C.webview_set_html(w.w, s)
}

func (w *webview) SetTitle(title string) {
	s := C.CString(title)
	defer C.free(unsafe.Pointer(s))
	C.webview_set_title(w.w, s)
}

func (w *webview) SetSize(width int, height int, hint Hint) {
	C.webview_set_size(w.w, C.int(width), C.int(height), C.webview_hint_t(hint))
}

func (w *webview) Init(js string) {
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	C.webview_init(w.w, s)
}

func (w *webview) Eval(js string) {
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	C.webview_eval(w.w, s)
}

func (w *webview) Dispatch(f func()) {
	m.Lock()
	for ; dispatch[index] != nil; index++ {
	}
	dispatch[index] = f
	m.Unlock()
	C.CgoWebViewDispatch(w.w, C.uintptr_t(index))
}

//export _webviewDispatchGoCallback
func _webviewDispatchGoCallback(index unsafe.Pointer) {
	m.Lock()
	f := dispatch[uintptr(index)]
	delete(dispatch, uintptr(index))
	m.Unlock()
	f()
}
