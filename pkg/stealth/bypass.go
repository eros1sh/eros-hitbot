package stealth

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
)

// StealthConfig anti-detection ayarları
type StealthConfig struct {
	UserAgent           string
	Platform            string
	Vendor              string
	WebGLVendor         string
	WebGLRenderer       string
	Languages           []string
	Plugins             []Plugin
	MimeTypes           []MimeType
	ScreenWidth         int
	ScreenHeight        int
	AvailWidth          int
	AvailHeight         int
	ColorDepth          int
	PixelDepth          int
	HardwareConcurrency int
	DeviceMemory        int
}

// Plugin tarayıcı eklentisi
type Plugin struct {
	Name        string
	Description string
	Filename    string
}

// MimeType MIME tipi
type MimeType struct {
	Type        string
	Description string
	Suffixes    string
}

// InjectStealthScripts tüm anti-detection scriptlerini enjekte eder
func InjectStealthScripts(ctx context.Context, config StealthConfig) error {
	scripts := []string{
		getWebdriverScript(),
		getChromeRuntimeScript(),
		getPermissionsScript(),
		getPluginsScript(config.Plugins),
		getLanguagesScript(config.Languages),
		getWebGLScript(config.WebGLVendor, config.WebGLRenderer),
		getNavigatorScript(config),
		getIframeContentWindowScript(),
		getMediaDevicesScript(),
		getOutDimensionScript(config),
		getVendorScript(config.Vendor),
		getCodecsScript(),
		getBatteryScript(),
		getConnectionScript(),
	}

	for _, script := range scripts {
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("stealth script injection failed: %w", err)
		}
	}

	return nil
}

// GetOnNewDocumentScript sayfa yüklenmeden önce çalışacak birleşik script
func GetOnNewDocumentScript(config StealthConfig) string {
	return strings.Join([]string{
		getWebdriverScript(),
		getChromeRuntimeScript(),
		getPermissionsScript(),
		getPluginsScript(config.Plugins),
		getLanguagesScript(config.Languages),
		getWebGLScript(config.WebGLVendor, config.WebGLRenderer),
		getNavigatorScript(config),
		getIframeContentWindowScript(),
		getMediaDevicesScript(),
		getOutDimensionScript(config),
		getVendorScript(config.Vendor),
		getCodecsScript(),
		getBatteryScript(),
		getConnectionScript(),
	}, "\n")
}

func getWebdriverScript() string {
	return `
(function(){
	try{
		Object.defineProperty(navigator,'webdriver',{get:()=>undefined,configurable:true});
		delete Object.getPrototypeOf(navigator).webdriver;
		delete window.webdriver;
		var cdc=Object.getOwnPropertyNames(window).filter(function(p){return /^cdc_.*?_/.test(p);});
		cdc.forEach(function(p){delete window[p];});
	}catch(e){}
})();`
}

func getChromeRuntimeScript() string {
	return `
(function(){
	if(!window.chrome){window.chrome={};}
	if(!window.chrome.runtime){window.chrome.runtime={};}
	window.chrome.runtime.sendMessage=function(){};
	window.chrome.runtime.connect=function(){
		return{onMessage:{addListener:function(){},removeListener:function(){}},postMessage:function(){},disconnect:function(){}};
	};
	window.chrome.loadTimes=function(){
		var t=performance.timing;
		return{commitLoadTime:t.domContentLoadedEventStart/1000,connectionInfo:'http/1.1',finishDocumentLoadTime:t.domContentLoadedEventEnd/1000,finishLoadTime:t.loadEventEnd/1000,firstPaintAfterLoadTime:0,firstPaintTime:t.domLoading/1000,navigationType:'Other',npnNegotiatedProtocol:'http/1.1',requestTime:t.navigationStart/1000,startLoadTime:t.navigationStart/1000,wasAlternateProtocolAvailable:false,wasFetchedViaSpdy:false,wasNpnNegotiated:false};
	};
	window.chrome.csi=function(){
		var t=performance.timing;
		return{onloadT:t.domContentLoadedEventEnd,pageT:t.loadEventEnd-t.navigationStart,startE:t.navigationStart,tran:15};
	};
	window.chrome.app={isInstalled:false,InstallState:{DISABLED:'disabled',INSTALLED:'installed',NOT_INSTALLED:'not_installed'},RunningState:{CANNOT_RUN:'cannot_run',READY_TO_RUN:'ready_to_run',RUNNING:'running'}};
})();`
}

func getPermissionsScript() string {
	return `
(function(){
	var oq=navigator.permissions&&navigator.permissions.query;
	if(oq){
		navigator.permissions.query=function(p){
			return p.name==='notifications'?Promise.resolve({state:Notification.permission}):oq.call(navigator.permissions,p);
		};
	}
})();`
}

func getPluginsScript(plugins []Plugin) string {
	if len(plugins) == 0 {
		plugins = getDefaultPlugins()
	}
	var parts []string
	for _, p := range plugins {
		parts = append(parts, fmt.Sprintf(`{name:'%s',description:'%s',filename:'%s',length:1,item:function(){return this[0];},namedItem:function(){return this[0];}}`,
			escapeJS(p.Name), escapeJS(p.Description), escapeJS(p.Filename)))
	}
	pluginsJS := "[" + strings.Join(parts, ",") + "]"
	return fmt.Sprintf(`
(function(){
	Object.defineProperty(navigator,'plugins',{get:function(){var p=%s;p.refresh=function(){};return p;},configurable:true});
	Object.defineProperty(navigator,'mimeTypes',{get:function(){return[{type:'application/pdf',description:'Portable Document Format',suffixes:'pdf'},{type:'application/x-google-chrome-pdf',description:'Portable Document Format',suffixes:'pdf'}];},configurable:true});
})();`, pluginsJS)
}

func getLanguagesScript(languages []string) string {
	if len(languages) == 0 {
		languages = []string{"en-US", "en"}
	}
	var parts []string
	for _, l := range languages {
		parts = append(parts, "'"+escapeJS(l)+"'")
	}
	return fmt.Sprintf(`(function(){Object.defineProperty(navigator,'languages',{get:function(){return [%s];},configurable:true});})();`, strings.Join(parts, ","))
}

func getWebGLScript(vendor, renderer string) string {
	if vendor == "" {
		vendor = "Google Inc. (Intel)"
	}
	if renderer == "" {
		renderer = "ANGLE (Intel, Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11)"
	}
	vendor = escapeJS(vendor)
	renderer = escapeJS(renderer)
	return fmt.Sprintf(`
(function(){
	var gp=WebGLRenderingContext.prototype.getParameter;
	WebGLRenderingContext.prototype.getParameter=function(p){
		if(p===37445)return '%s';
		if(p===37446)return '%s';
		return gp.apply(this,arguments);
	};
	if(window.WebGL2RenderingContext){
		var gp2=WebGL2RenderingContext.prototype.getParameter;
		WebGL2RenderingContext.prototype.getParameter=function(p){
			if(p===37445)return '%s';
			if(p===37446)return '%s';
			return gp2.apply(this,arguments);
		};
	}
})();`, vendor, renderer, vendor, renderer)
}

func getNavigatorScript(config StealthConfig) string {
	ua := escapeJS(config.UserAgent)
	plat := escapeJS(config.Platform)
	if plat == "" {
		plat = "Win32"
	}
	hw := config.HardwareConcurrency
	if hw <= 0 {
		hw = 8
	}
	dm := config.DeviceMemory
	if dm <= 0 {
		dm = 8
	}
	return fmt.Sprintf(`
(function(){
	Object.defineProperty(navigator,'platform',{get:function(){return '%s';},configurable:true});
	Object.defineProperty(navigator,'hardwareConcurrency',{get:function(){return %d;},configurable:true});
	Object.defineProperty(navigator,'deviceMemory',{get:function(){return %d;},configurable:true});
	Object.defineProperty(navigator,'maxTouchPoints',{get:function(){return 0;},configurable:true});
	Object.defineProperty(navigator,'userAgent',{get:function(){return '%s';},configurable:true});
	Object.defineProperty(navigator,'appVersion',{get:function(){return '%s';},configurable:true});
	Object.defineProperty(navigator,'doNotTrack',{get:function(){return null;},configurable:true});
	Object.defineProperty(navigator,'product',{get:function(){return 'Gecko';},configurable:true});
	Object.defineProperty(navigator,'productSub',{get:function(){return '20030107';},configurable:true});
})();`, plat, hw, dm, ua, ua)
}

func getIframeContentWindowScript() string {
	return `(function(){try{Object.defineProperty(HTMLIFrameElement.prototype,'contentWindow',{get:function(){return window;}});}catch(e){}})();`
}

func getMediaDevicesScript() string {
	return `
(function(){
	if(navigator.mediaDevices&&navigator.mediaDevices.enumerateDevices){
		navigator.mediaDevices.enumerateDevices=function(){
			return Promise.resolve([{deviceId:'default',kind:'audioinput',label:'Default - Microphone Array',groupId:'a9c2e7e5e5f4f5e5'},{deviceId:'default',kind:'audiooutput',label:'Default - Speakers',groupId:'d9c2e7e5e5f4f5e5'},{deviceId:'default',kind:'videoinput',label:'HD Webcam',groupId:'c9c2e7e5e5f4f5e5'}]);
		};
	}
})();`
}

func getOutDimensionScript(config StealthConfig) string {
	sw, sh := config.ScreenWidth, config.ScreenHeight
	aw, ah := config.AvailWidth, config.AvailHeight
	cd, pd := config.ColorDepth, config.PixelDepth
	if sw <= 0 {
		sw, sh = 1920, 1080
	}
	if aw <= 0 {
		aw, ah = sw, sh-40
	}
	if cd <= 0 {
		cd, pd = 24, 24
	}
	return fmt.Sprintf(`
(function(){
	Object.defineProperty(screen,'availWidth',{get:function(){return %d;},configurable:true});
	Object.defineProperty(screen,'availHeight',{get:function(){return %d;},configurable:true});
	Object.defineProperty(screen,'width',{get:function(){return %d;},configurable:true});
	Object.defineProperty(screen,'height',{get:function(){return %d;},configurable:true});
	Object.defineProperty(screen,'colorDepth',{get:function(){return %d;},configurable:true});
	Object.defineProperty(screen,'pixelDepth',{get:function(){return %d;},configurable:true});
	Object.defineProperty(window,'outerWidth',{get:function(){return %d;},configurable:true});
	Object.defineProperty(window,'outerHeight',{get:function(){return %d;},configurable:true});
	Object.defineProperty(window,'innerWidth',{get:function(){return %d;},configurable:true});
	Object.defineProperty(window,'innerHeight',{get:function(){return %d;},configurable:true});
})();`, aw, ah, sw, sh, cd, pd, sw, sh, sw-17, sh-150)
}

func getVendorScript(vendor string) string {
	if vendor == "" {
		vendor = "Google Inc."
	}
	return fmt.Sprintf(`(function(){Object.defineProperty(navigator,'vendor',{get:function(){return '%s';},configurable:true});})();`, escapeJS(vendor))
}

func getCodecsScript() string {
	return `
(function(){
	var cpt=HTMLMediaElement.prototype.canPlayType;
	HTMLMediaElement.prototype.canPlayType=function(t){
		if(t==='video/mp4; codecs="avc1.42E01E"')return 'probably';
		if(t==='video/webm; codecs="vp8, vorbis"')return 'probably';
		return cpt.apply(this,arguments);
	};
})();`
}

func getBatteryScript() string {
	return `
(function(){
	if(navigator.getBattery){
		navigator.getBattery=function(){
			return Promise.resolve({charging:true,chargingTime:0,dischargingTime:Infinity,level:1,addEventListener:function(){},removeEventListener:function(){},dispatchEvent:function(){return true;}});
		};
	}
})();`
}

func getConnectionScript() string {
	return `
(function(){
	Object.defineProperty(navigator,'connection',{get:function(){return{effectiveType:'4g',rtt:50,downlink:10,saveData:false,addEventListener:function(){},removeEventListener:function(){},dispatchEvent:function(){return true;}};},configurable:true});
})();`
}

func getDefaultPlugins() []Plugin {
	return []Plugin{
		{Name: "Chrome PDF Plugin", Description: "Portable Document Format", Filename: "internal-pdf-viewer"},
		{Name: "Chrome PDF Viewer", Description: "Portable Document Format", Filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai"},
		{Name: "Native Client", Description: "", Filename: "internal-nacl-plugin"},
	}
}

func escapeJS(s string) string {
	return strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "").Replace(s)
}

// GetDefaultStealthConfig varsayılan stealth konfigürasyonu
func GetDefaultStealthConfig() StealthConfig {
	return StealthConfig{
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Platform:      "Win32",
		Vendor:        "Google Inc.",
		WebGLVendor:   "Google Inc. (Intel)",
		WebGLRenderer: "ANGLE (Intel, Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		Languages:     []string{"en-US", "en"},
		Plugins:       getDefaultPlugins(),
		ScreenWidth:   1920,
		ScreenHeight:  1080,
		AvailWidth:    1920,
		AvailHeight:   1040,
		ColorDepth:    24,
		PixelDepth:    24,
	}
}
