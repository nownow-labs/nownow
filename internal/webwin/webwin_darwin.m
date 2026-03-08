#import <Cocoa/Cocoa.h>
#include <stdint.h>

// Forward declaration (defined below, called by windowShouldClose:)
void webwin_hideWindow(uintptr_t wndPtr);

@interface WebWinDelegate : NSObject <NSWindowDelegate>
@end

@implementation WebWinDelegate

- (BOOL)windowShouldClose:(id)sender {
    webwin_hideWindow((uintptr_t)sender);
    return NO;
}

@end

static WebWinDelegate *webWinDelegate = nil;

void webwin_setWindowDelegate(uintptr_t wndPtr) {
    NSWindow *w = (__bridge NSWindow *)(void *)wndPtr;
    webWinDelegate = [[WebWinDelegate alloc] init];
    [w setDelegate:webWinDelegate];
}

void webwin_hideWindow(uintptr_t wndPtr) {
    NSWindow *w = (__bridge NSWindow *)(void *)wndPtr;
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    [w orderOut:nil];
}

void webwin_setAppIcon(const void *data, int length) {
    NSData *imgData = [NSData dataWithBytes:data length:length];
    NSImage *icon = [[NSImage alloc] initWithData:imgData];
    if (icon) {
        [NSApp setApplicationIconImage:icon];
    }
}

void webwin_showWindow(uintptr_t wndPtr) {
    NSWindow *w = (__bridge NSWindow *)(void *)wndPtr;
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
        [NSApp unhide:nil];
        [NSApp activateIgnoringOtherApps:YES];
        [w makeKeyAndOrderFront:nil];
    });
}
