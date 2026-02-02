package fingerprint

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"strings"
	"sync"
	"time"
)

// Plugin tarayıcı eklentisi bilgisi
type Plugin struct {
	Name        string
	Filename    string
	Description string
}

// AdvancedFingerprint gelişmiş tarayıcı fingerprint'i
type AdvancedFingerprint struct {
	UserAgent           string
	Platform            string
	Language            string
	Languages           []string
	HardwareConcurrency int
	DeviceMemory        float64
	MaxTouchPoints      int
	ScreenWidth         int
	ScreenHeight        int
	ScreenColorDepth    int
	ScreenPixelRatio    float64
	AvailWidth          int
	AvailHeight         int
	WebGLVendor         string
	WebGLRenderer       string
	CanvasHash          string
	AudioHash           string
	Fonts               []string
	Plugins             []Plugin
	TimezoneOffset      int
	Timezone            string
	DoNotTrack          string
	CookieEnabled       bool
	OnLine              bool
	OSVersion           string
}

var advRng = mrand.New(mrand.NewSource(time.Now().UnixNano()))
var advMu sync.Mutex

func intn(max int) int {
	advMu.Lock()
	defer advMu.Unlock()
	return advRng.Intn(max)
}

func generateHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateCanvasHash() string {
	return generateHex(16)
}

func generateAudioHash() string {
	return generateHex(16)
}

// Platform'a göre font listesi
func generateRealisticFonts(platform string) []string {
	base := []string{"Arial", "Helvetica", "Times New Roman", "Courier New", "Verdana", "Georgia"}
	switch {
	case strings.Contains(platform, "Win"):
		return append(base, "Segoe UI", "Calibri", "Tahoma", "Trebuchet MS", "Microsoft Sans Serif")
	case strings.Contains(platform, "Mac"):
		return append(base, "Helvetica Neue", "San Francisco", "Lucida Grande", "Menlo", "Monaco")
	case strings.Contains(platform, "Linux"):
		return append(base, "Ubuntu", "Liberation Sans", "DejaVu Sans", "FreeSans")
	default:
		return base
	}
}

// Platform'a göre WebGL bilgisi
func generateWebGLInfo(platform string) (vendor, renderer string) {
	vendors := []string{"Google Inc. (NVIDIA)", "Google Inc. (Intel)", "Google Inc. (AMD)", "Google Inc. (Microsoft)", "Apple Inc."}
	renderers := []string{
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1660 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Microsoft, Microsoft Basic Render Driver Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Apple, Apple M1, OpenGL 4.1)",
		"ANGLE (AMD, AMD Radeon RX 580 Series Direct3D11 vs_5_0 ps_5_0)",
	}
	i := intn(len(vendors))
	return vendors[i], renderers[i]
}

// Platform'a göre tutarlı donanım değerleri
func getPlatformConstraints(platform string) (hwMin, hwMax int, memMin, memMax float64, touch int) {
	switch {
	case strings.Contains(platform, "Win"):
		return 4, 16, 8, 32, 0
	case strings.Contains(platform, "Mac"):
		return 4, 10, 8, 64, 0
	case strings.Contains(platform, "Linux"):
		return 2, 16, 4, 32, 0
	case strings.Contains(platform, "Android"):
		return 4, 8, 4, 12, 5 + intn(6)
	case strings.Contains(platform, "iPhone") || strings.Contains(platform, "iPad"):
		return 6, 8, 4, 6, 5
	default:
		return 4, 16, 8, 32, 0
	}
}

// GenerateAdvancedFingerprint benzersiz ve tutarlı fingerprint üretir
func GenerateAdvancedFingerprint() *AdvancedFingerprint {
	platforms := []string{"Win32", "MacIntel", "Linux x86_64"}
	platform := platforms[intn(len(platforms))]

	hwMin, hwMax, memMin, memMax, touch := getPlatformConstraints(platform)
	hwChoices := []int{2, 4, 6, 8, 12, 16}
	memChoices := []float64{2, 4, 8, 16, 32, 64}

	var hw int
	for _, c := range hwChoices {
		if c >= hwMin && c <= hwMax {
			hw = c
			break
		}
		hw = c
	}
	if hw < hwMin {
		hw = hwMin
	}
	if hw > hwMax {
		hw = hwMax
	}

	var mem float64
	for _, m := range memChoices {
		if m >= memMin && m <= memMax {
			mem = m
			break
		}
		mem = m
	}
	if mem < memMin {
		mem = memMin
	}

	sw := 1366 + intn(600)
	sh := 768 + intn(400)
	availW := sw - 10
	availH := sh - 80

	colorDepths := []int{24, 30, 32}
	pixelRatios := []float64{1.0, 1.5, 2.0, 3.0}

	langs := []string{"tr-TR", "en-US", "en-GB", "de-DE"}
	lang := langs[intn(len(langs))]
	langList := []string{lang, "tr", "en"}

	timezones := []struct {
		tz     string
		offset int
	}{
		{"Europe/Istanbul", -180},
		{"Europe/London", 0},
		{"America/New_York", 300},
		{"Europe/Paris", -60},
	}
	tzSel := timezones[intn(len(timezones))]

	vendor, renderer := generateWebGLInfo(platform)

	doNotTrack := []string{"1", "unspecified", "null"}
	dnt := doNotTrack[intn(len(doNotTrack))]

	osVersions := map[string][]string{
		"Win32":        {"Windows NT 10.0", "Windows NT 6.3"},
		"MacIntel":     {"Mac OS X 10_15_7", "Mac OS X 10_14_6"},
		"Linux x86_64": {"Linux x86_64"},
	}
	osVer := "Windows NT 10.0"
	if v, ok := osVersions[platform]; ok && len(v) > 0 {
		osVer = v[intn(len(v))]
	}

	return &AdvancedFingerprint{
		UserAgent:           "Mozilla/5.0 (compatible; generated)",
		Platform:            platform,
		Language:            lang,
		Languages:           langList,
		HardwareConcurrency: hw,
		DeviceMemory:        mem,
		MaxTouchPoints:      touch,
		ScreenWidth:         sw,
		ScreenHeight:        sh,
		ScreenColorDepth:    colorDepths[intn(len(colorDepths))],
		ScreenPixelRatio:    pixelRatios[intn(len(pixelRatios))],
		AvailWidth:          availW,
		AvailHeight:         availH,
		WebGLVendor:         vendor,
		WebGLRenderer:       renderer,
		CanvasHash:          generateCanvasHash(),
		AudioHash:           generateAudioHash(),
		Fonts:               generateRealisticFonts(platform),
		Plugins:             []Plugin{},
		TimezoneOffset:      tzSel.offset,
		Timezone:            tzSel.tz,
		DoNotTrack:          dnt,
		CookieEnabled:       true,
		OnLine:              true,
		OSVersion:           osVer,
	}
}

// ToChromedpScript Chromedp'de çalıştırılacak JS kodu üretir
func (f *AdvancedFingerprint) ToChromedpScript() string {
	langsJS := make([]string, len(f.Languages))
	for i, l := range f.Languages {
		langsJS[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(l, "'", "\\'"))
	}
	langsStr := strings.Join(langsJS, ",")

	return fmt.Sprintf(`(function(){
try{
Object.defineProperty(navigator,'webdriver',{get:()=>undefined,configurable:true});
Object.defineProperty(navigator,'platform',{get:()=>'%s',configurable:true});
Object.defineProperty(navigator,'language',{get:()=>'%s',configurable:true});
Object.defineProperty(navigator,'languages',{get:()=>[%s],configurable:true});
Object.defineProperty(navigator,'hardwareConcurrency',{get:()=>%d,configurable:true});
Object.defineProperty(navigator,'deviceMemory',{get:()=>%g,configurable:true});
Object.defineProperty(navigator,'maxTouchPoints',{get:()=>%d,configurable:true});
Object.defineProperty(navigator,'vendor',{get:()=>'%s',configurable:true});
Object.defineProperty(navigator,'cookieEnabled',{get:()=>%t,configurable:true});
Object.defineProperty(navigator,'onLine',{get:()=>%t,configurable:true});
Object.defineProperty(navigator,'doNotTrack',{get:()=>%s,configurable:true});
if(window.screen){
Object.defineProperty(screen,'width',{get:()=>%d,configurable:true});
Object.defineProperty(screen,'height',{get:()=>%d,configurable:true});
Object.defineProperty(screen,'availWidth',{get:()=>%d,configurable:true});
Object.defineProperty(screen,'availHeight',{get:()=>%d,configurable:true});
Object.defineProperty(screen,'colorDepth',{get:()=>%d,configurable:true});
}
}catch(e){}
})();`,
		escapeJS(f.Platform),
		escapeJS(f.Language),
		langsStr,
		f.HardwareConcurrency,
		f.DeviceMemory,
		f.MaxTouchPoints,
		escapeJS(f.WebGLVendor),
		f.CookieEnabled,
		f.OnLine,
		doNotTrackJS(f.DoNotTrack),
		f.ScreenWidth, f.ScreenHeight, f.AvailWidth, f.AvailHeight, f.ScreenColorDepth,
	)
}

func escapeJS(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "'", "\\'")
}

func doNotTrackJS(s string) string {
	if s == "null" {
		return "null"
	}
	return "'" + escapeJS(s) + "'"
}

// IsConsistent iç tutarlılık kontrolü
func (f *AdvancedFingerprint) IsConsistent() bool {
	hwMin, hwMax, memMin, memMax, touch := getPlatformConstraints(f.Platform)
	if f.HardwareConcurrency < hwMin || f.HardwareConcurrency > hwMax {
		return false
	}
	if f.DeviceMemory < memMin || f.DeviceMemory > memMax {
		return false
	}
	if f.MaxTouchPoints != touch {
		return false
	}
	if f.ScreenWidth < f.AvailWidth || f.ScreenHeight < f.AvailHeight {
		return false
	}
	return true
}
