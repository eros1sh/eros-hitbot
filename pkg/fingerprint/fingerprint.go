package fingerprint

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var platforms = []string{"Win32", "MacIntel", "Linux x86_64"}
var langs = []string{"tr-TR"}
var timezones = []string{"Europe/Istanbul", "Europe/London", "America/New_York", "Europe/Paris"}
var vendors = []string{"Google Inc.", "Apple Computer, Inc."}
var renderers = []string{
	"ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0)",
	"ANGLE (Microsoft, Microsoft Basic Render Driver Direct3D11 vs_5_0 ps_5_0)",
	"ANGLE (Apple, Apple M1, OpenGL 4.1)",
	"ANGLE (NVIDIA, NVIDIA GeForce GTX 1060 Direct3D11 vs_5_0 ps_5_0)",
}

type FP struct {
	Platform     string
	Language     string
	Languages    string
	ScreenW      int
	ScreenH      int
	AvailW       int
	AvailH       int
	InnerW       int
	InnerH       int
	OuterW       int
	OuterH       int
	DevicePixel  float64
	HardwareConc int
	DeviceMem    int64
	ColorDepth   int
	Timezone     string
	Vendor       string
	Renderer     string
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var mu sync.Mutex

func Random() FP {
	mu.Lock()
	defer mu.Unlock()

	sw := 1366 + rng.Intn(600)
	sh := 768 + rng.Intn(400)
	availW := sw - 10
	availH := sh - 80
	innerW := availW - 22
	innerH := availH - 100
	outerW := innerW + 22
	outerH := innerH + 100

	return FP{
		Platform:     platforms[rng.Intn(len(platforms))],
		Language:     langs[rng.Intn(len(langs))],
		Languages:    langs[rng.Intn(len(langs))] + ", " + langs[rng.Intn(len(langs))] + ", en",
		ScreenW:      sw,
		ScreenH:      sh,
		AvailW:       availW,
		AvailH:       availH,
		InnerW:       innerW,
		InnerH:       innerH,
		OuterW:       outerW,
		OuterH:       outerH,
		DevicePixel:  1.0 + float64(rng.Intn(2)),
		HardwareConc: []int{2, 4, 8, 12, 16}[rng.Intn(5)],
		DeviceMem:    []int64{4, 8, 16, 32}[rng.Intn(4)],
		ColorDepth:   24,
		Timezone:     timezones[rng.Intn(len(timezones))],
		Vendor:       vendors[rng.Intn(len(vendors))],
		Renderer:     renderers[rng.Intn(len(renderers))],
	}
}

// InjectScript webdriver gizleme ve navigator/screen override - Page.addScriptToEvaluateOnNewDocument için
func InjectScript(fp FP) string {
	parts := strings.Split(fp.Languages, ", ")
	for i, p := range parts {
		parts[i] = "'" + strings.TrimSpace(p) + "'"
	}
	langsArr := strings.Join(parts, ", ")
	return fmt.Sprintf(
		`(function(){try{Object.defineProperty(navigator,'webdriver',{get:()=>undefined,configurable:true});Object.defineProperty(navigator,'platform',{get:()=>'%s',configurable:true});Object.defineProperty(navigator,'language',{get:()=>'%s',configurable:true});Object.defineProperty(navigator,'languages',{get:()=>[%s],configurable:true});Object.defineProperty(navigator,'hardwareConcurrency',{get:()=>%d,configurable:true});Object.defineProperty(navigator,'deviceMemory',{get:()=>%d,configurable:true});Object.defineProperty(navigator,'vendor',{get:()=>'%s',configurable:true});}catch(e){}})();`,
		fp.Platform, fp.Language, langsArr, fp.HardwareConc, fp.DeviceMem, fp.Vendor)
}
