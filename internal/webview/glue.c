#include "webview.h"

#include <stdlib.h>
#include <stdint.h>

void _webviewDispatchGoCallback(void *);

static void _webview_dispatch_cb(webview_t w, void *arg) {
    _webviewDispatchGoCallback(arg);
}

void CgoWebViewDispatch(webview_t w, uintptr_t arg) {
    webview_dispatch(w, _webview_dispatch_cb, (void *)arg);
}
