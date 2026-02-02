package mobile

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// Emulator mobil cihaz emülatörü
type Emulator struct {
	Device DeviceProfile
}

// NewEmulator yeni emülatör
func NewEmulator(device DeviceProfile) *Emulator {
	return &Emulator{Device: device}
}

// ApplyDeviceEmulation cihaz ayarlarını uygular
func (e *Emulator) ApplyDeviceEmulation(ctx context.Context) error {
	if err := emulation.SetDeviceMetricsOverride(
		int64(e.Device.ScreenWidth),
		int64(e.Device.ScreenHeight),
		e.Device.PixelRatio,
		e.Device.Mobile,
	).Do(ctx); err != nil {
		return err
	}
	if e.Device.TouchEnabled {
		if err := emulation.SetTouchEmulationEnabled(true).
			WithMaxTouchPoints(int64(e.Device.MaxTouchPoints)).
			Do(ctx); err != nil {
			return err
		}
	}
	return e.overrideNavigator(ctx)
}

func (e *Emulator) overrideNavigator(ctx context.Context) error {
	platform := e.getPlatformString()
	angle := "0"
	if e.Device.Orientation == "landscape" {
		angle = "90"
	}
	script := fmt.Sprintf(`(function(){
		try{
			Object.defineProperty(navigator,'platform',{get:()=>'%s',configurable:true});
			Object.defineProperty(navigator,'maxTouchPoints',{get:()=>%d,configurable:true});
			if(screen.orientation){
				Object.defineProperty(screen.orientation,'type',{get:()=>'%s-primary',configurable:true});
				Object.defineProperty(screen.orientation,'angle',{get:()=>%s,configurable:true});
			}
		}catch(e){}
	})();`,
		platform, e.Device.MaxTouchPoints, e.Device.Orientation, angle,
	)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

func (e *Emulator) getPlatformString() string {
	if e.Device.Platform == "iOS" {
		return "iPhone"
	}
	return "Linux armv8l"
}

// SimulateTouchEvent dokunma eventi
func (e *Emulator) SimulateTouchEvent(ctx context.Context, x, y int, eventType string) error {
	script := fmt.Sprintf(`(function(){
		var el = document.elementFromPoint(%d,%d);
		if(!el) return;
		var touch = new Touch({
			identifier: Date.now(), target: el, clientX: %d, clientY: %d,
			screenX: %d, screenY: %d, pageX: %d, pageY: %d
		});
		var ev = new TouchEvent('%s',{
			touches: [touch], targetTouches: [touch], changedTouches: [touch],
			bubbles: true, cancelable: true
		});
		el.dispatchEvent(ev);
	})();`, x, y, x, y, x, y, x, y, eventType)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

// SwipeGesture swipe hareketi
func (e *Emulator) SwipeGesture(ctx context.Context, startX, startY, endX, endY int, durationMs int) error {
	_ = e.SimulateTouchEvent(ctx, startX, startY, "touchstart")
	steps := durationMs / 10
	if steps < 2 {
		steps = 2
	}
	for i := 1; i < steps; i++ {
		progress := float64(i) / float64(steps)
		x := int(float64(startX) + progress*float64(endX-startX))
		y := int(float64(startY) + progress*float64(endY-startY))
		_ = e.SimulateTouchEvent(ctx, x, y, "touchmove")
	}
	return e.SimulateTouchEvent(ctx, endX, endY, "touchend")
}
