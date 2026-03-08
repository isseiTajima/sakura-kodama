#import <Cocoa/Cocoa.h>
#import <stdbool.h>

// Go側の関数を定義
extern void goOnTraySettingsClicked();
extern void goOnTrayQuitClicked();

@interface NativeTrayHandler : NSObject
- (void)onSettings:(id)sender;
- (void)onQuit:(id)sender;
@end

@implementation NativeTrayHandler
- (void)onSettings:(id)sender {
    goOnTraySettingsClicked();
}
- (void)onQuit:(id)sender {
    goOnTrayQuitClicked();
}
@end

static NSStatusItem *statusItem;
static NativeTrayHandler *trayHandler;

void SetupNativeTray(const char* iconPath) {
    // パスをコピーして保持
    NSString *path = iconPath ? [NSString stringWithUTF8String:iconPath] : nil;
    
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem == nil) {
            statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
            [statusItem retain];
            trayHandler = [[NativeTrayHandler alloc] init];
            [trayHandler retain];
        }
        
        NSImage *image = nil;
        if (path) {
            image = [[NSImage alloc] initWithContentsOfFile:path];
        }
        
        if (image) {
            [image setSize:NSMakeSize(18, 18)];
            [image setTemplate:YES];
            statusItem.button.image = image;
        } else {
            statusItem.button.title = @"🌸";
        }
        
        NSMenu *menu = [[NSMenu alloc] init];
        [menu addItemWithTitle:@"設定を開く" action:@selector(onSettings:) keyEquivalent:@","];
        [menu itemArray].lastObject.target = trayHandler;
        
        [menu addItem:[NSMenuItem separatorItem]];
        
        [menu addItemWithTitle:@"終了" action:@selector(onQuit:) keyEquivalent:@"q"];
        [menu itemArray].lastObject.target = trayHandler;
        
        statusItem.menu = menu;
    });
}

void SetWindowClickThroughNative(bool ignore) {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *window in [NSApp windows]) {
            // 全てのウィンドウに対してマウスイベントを無視（または許可）する設定を強制
            [window setIgnoresMouseEvents:ignore];
            
            // 透過を維持するための念押しの設定
            if (ignore) {
                [window setHasShadow:NO];
            } else {
                [window setHasShadow:YES];
            }
        }
    });
}
