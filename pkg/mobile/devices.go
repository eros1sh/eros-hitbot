package mobile

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

// DeviceType cihaz tipi
type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeMobile  DeviceType = "mobile"
	DeviceTypeTablet  DeviceType = "tablet"
	DeviceTypeMixed   DeviceType = "mixed"
)

// DeviceBrand cihaz markası
type DeviceBrand string

const (
	BrandApple   DeviceBrand = "apple"
	BrandSamsung DeviceBrand = "samsung"
	BrandGoogle  DeviceBrand = "google"
	BrandXiaomi  DeviceBrand = "xiaomi"
	BrandHuawei  DeviceBrand = "huawei"
	BrandOnePlus DeviceBrand = "oneplus"
	BrandWindows DeviceBrand = "windows"
	BrandLinux   DeviceBrand = "linux"
	BrandMac     DeviceBrand = "mac"
)

// DeviceProfile mobil cihaz profili
type DeviceProfile struct {
	Name           string
	UserAgent      string
	ScreenWidth    int
	ScreenHeight   int
	PixelRatio     float64
	Platform       string
	Mobile         bool
	TouchEnabled   bool
	MaxTouchPoints int
	Orientation    string
	Brand          DeviceBrand
	Type           DeviceType
}

// Apple Cihazları
var (
	IPhone13Pro = DeviceProfile{
		Name:           "iPhone 13 Pro",
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
		ScreenWidth:    390,
		ScreenHeight:   844,
		PixelRatio:     3.0,
		Platform:       "iOS",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 5,
		Orientation:    "portrait",
		Brand:          BrandApple,
		Type:           DeviceTypeMobile,
	}
	IPhone14ProMax = DeviceProfile{
		Name:           "iPhone 14 Pro Max",
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		ScreenWidth:    430,
		ScreenHeight:   932,
		PixelRatio:     3.0,
		Platform:       "iOS",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 5,
		Orientation:    "portrait",
		Brand:          BrandApple,
		Type:           DeviceTypeMobile,
	}
	IPhone15 = DeviceProfile{
		Name:           "iPhone 15",
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		ScreenWidth:    393,
		ScreenHeight:   852,
		PixelRatio:     3.0,
		Platform:       "iOS",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 5,
		Orientation:    "portrait",
		Brand:          BrandApple,
		Type:           DeviceTypeMobile,
	}
	IPadPro = DeviceProfile{
		Name:           "iPad Pro 11",
		UserAgent:      "Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
		ScreenWidth:    834,
		ScreenHeight:   1194,
		PixelRatio:     2.0,
		Platform:       "iOS",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 5,
		Orientation:    "portrait",
		Brand:          BrandApple,
		Type:           DeviceTypeTablet,
	}
	IPadAir = DeviceProfile{
		Name:           "iPad Air",
		UserAgent:      "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
		ScreenWidth:    820,
		ScreenHeight:   1180,
		PixelRatio:     2.0,
		Platform:       "iOS",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 5,
		Orientation:    "portrait",
		Brand:          BrandApple,
		Type:           DeviceTypeTablet,
	}
)

// Samsung Cihazları
var (
	SamsungGalaxyS21 = DeviceProfile{
		Name:           "Samsung Galaxy S21",
		UserAgent:      "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36",
		ScreenWidth:    360,
		ScreenHeight:   800,
		PixelRatio:     3.0,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandSamsung,
		Type:           DeviceTypeMobile,
	}
	SamsungGalaxyS23Ultra = DeviceProfile{
		Name:           "Samsung Galaxy S23 Ultra",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    384,
		ScreenHeight:   824,
		PixelRatio:     3.0,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandSamsung,
		Type:           DeviceTypeMobile,
	}
	SamsungGalaxyA54 = DeviceProfile{
		Name:           "Samsung Galaxy A54",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; SM-A546B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    360,
		ScreenHeight:   780,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandSamsung,
		Type:           DeviceTypeMobile,
	}
	SamsungGalaxyTabS8 = DeviceProfile{
		Name:           "Samsung Galaxy Tab S8",
		UserAgent:      "Mozilla/5.0 (Linux; Android 12; SM-X700) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		ScreenWidth:    800,
		ScreenHeight:   1280,
		PixelRatio:     2.0,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandSamsung,
		Type:           DeviceTypeTablet,
	}
)

// Google Cihazları
var (
	GooglePixel6 = DeviceProfile{
		Name:           "Google Pixel 6",
		UserAgent:      "Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36",
		ScreenWidth:    412,
		ScreenHeight:   915,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandGoogle,
		Type:           DeviceTypeMobile,
	}
	GooglePixel7Pro = DeviceProfile{
		Name:           "Google Pixel 7 Pro",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    412,
		ScreenHeight:   892,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandGoogle,
		Type:           DeviceTypeMobile,
	}
	GooglePixel8 = DeviceProfile{
		Name:           "Google Pixel 8",
		UserAgent:      "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    412,
		ScreenHeight:   915,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandGoogle,
		Type:           DeviceTypeMobile,
	}
)

// Xiaomi Cihazları
var (
	XiaomiMi11 = DeviceProfile{
		Name:           "Xiaomi Mi 11",
		UserAgent:      "Mozilla/5.0 (Linux; Android 11; M2011K2G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36",
		ScreenWidth:    393,
		ScreenHeight:   873,
		PixelRatio:     2.75,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandXiaomi,
		Type:           DeviceTypeMobile,
	}
	Xiaomi13Pro = DeviceProfile{
		Name:           "Xiaomi 13 Pro",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; 2210132C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    393,
		ScreenHeight:   873,
		PixelRatio:     2.75,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandXiaomi,
		Type:           DeviceTypeMobile,
	}
	RedmiNote12 = DeviceProfile{
		Name:           "Redmi Note 12",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; 23021RAA2Y) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    393,
		ScreenHeight:   851,
		PixelRatio:     2.75,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandXiaomi,
		Type:           DeviceTypeMobile,
	}
)

// Huawei Cihazları
var (
	HuaweiP40Pro = DeviceProfile{
		Name:           "Huawei P40 Pro",
		UserAgent:      "Mozilla/5.0 (Linux; Android 10; ELS-NX9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36",
		ScreenWidth:    360,
		ScreenHeight:   780,
		PixelRatio:     3.0,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandHuawei,
		Type:           DeviceTypeMobile,
	}
	HuaweiMate50Pro = DeviceProfile{
		Name:           "Huawei Mate 50 Pro",
		UserAgent:      "Mozilla/5.0 (Linux; Android 12; DCO-AL00) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    360,
		ScreenHeight:   780,
		PixelRatio:     3.0,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandHuawei,
		Type:           DeviceTypeMobile,
	}
)

// OnePlus Cihazları
var (
	OnePlus11 = DeviceProfile{
		Name:           "OnePlus 11",
		UserAgent:      "Mozilla/5.0 (Linux; Android 13; CPH2449) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36",
		ScreenWidth:    412,
		ScreenHeight:   915,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandOnePlus,
		Type:           DeviceTypeMobile,
	}
	OnePlus9Pro = DeviceProfile{
		Name:           "OnePlus 9 Pro",
		UserAgent:      "Mozilla/5.0 (Linux; Android 11; LE2123) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.104 Mobile Safari/537.36",
		ScreenWidth:    412,
		ScreenHeight:   906,
		PixelRatio:     2.625,
		Platform:       "Android",
		Mobile:         true,
		TouchEnabled:   true,
		MaxTouchPoints: 10,
		Orientation:    "portrait",
		Brand:          BrandOnePlus,
		Type:           DeviceTypeMobile,
	}
)

// Desktop Cihazları
var (
	WindowsDesktopChrome = DeviceProfile{
		Name:           "Windows Desktop Chrome",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     1.0,
		Platform:       "Win32",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandWindows,
		Type:           DeviceTypeDesktop,
	}
	WindowsDesktopFirefox = DeviceProfile{
		Name:           "Windows Desktop Firefox",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     1.0,
		Platform:       "Win32",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandWindows,
		Type:           DeviceTypeDesktop,
	}
	WindowsDesktopEdge = DeviceProfile{
		Name:           "Windows Desktop Edge",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     1.0,
		Platform:       "Win32",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandWindows,
		Type:           DeviceTypeDesktop,
	}
	Windows11Desktop = DeviceProfile{
		Name:           "Windows 11 Desktop",
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		ScreenWidth:    2560,
		ScreenHeight:   1440,
		PixelRatio:     1.25,
		Platform:       "Win32",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandWindows,
		Type:           DeviceTypeDesktop,
	}
	MacDesktopSafari = DeviceProfile{
		Name:           "Mac Desktop Safari",
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     2.0,
		Platform:       "MacIntel",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandMac,
		Type:           DeviceTypeDesktop,
	}
	MacDesktopChrome = DeviceProfile{
		Name:           "Mac Desktop Chrome",
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     2.0,
		Platform:       "MacIntel",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandMac,
		Type:           DeviceTypeDesktop,
	}
	MacBookPro = DeviceProfile{
		Name:           "MacBook Pro 14",
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		ScreenWidth:    3024,
		ScreenHeight:   1964,
		PixelRatio:     2.0,
		Platform:       "MacIntel",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandMac,
		Type:           DeviceTypeDesktop,
	}
	LinuxDesktopChrome = DeviceProfile{
		Name:           "Linux Desktop Chrome",
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     1.0,
		Platform:       "Linux x86_64",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandLinux,
		Type:           DeviceTypeDesktop,
	}
	LinuxDesktopFirefox = DeviceProfile{
		Name:           "Linux Desktop Firefox",
		UserAgent:      "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
		ScreenWidth:    1920,
		ScreenHeight:   1080,
		PixelRatio:     1.0,
		Platform:       "Linux x86_64",
		Mobile:         false,
		TouchEnabled:   false,
		MaxTouchPoints: 0,
		Orientation:    "landscape",
		Brand:          BrandLinux,
		Type:           DeviceTypeDesktop,
	}
)

// Tüm cihaz listesi
var allDevices = []DeviceProfile{
	// Apple
	IPhone13Pro, IPhone14ProMax, IPhone15, IPadPro, IPadAir,
	// Samsung
	SamsungGalaxyS21, SamsungGalaxyS23Ultra, SamsungGalaxyA54, SamsungGalaxyTabS8,
	// Google
	GooglePixel6, GooglePixel7Pro, GooglePixel8,
	// Xiaomi
	XiaomiMi11, Xiaomi13Pro, RedmiNote12,
	// Huawei
	HuaweiP40Pro, HuaweiMate50Pro,
	// OnePlus
	OnePlus11, OnePlus9Pro,
	// Desktop
	WindowsDesktopChrome, WindowsDesktopFirefox, WindowsDesktopEdge, Windows11Desktop,
	MacDesktopSafari, MacDesktopChrome, MacBookPro,
	LinuxDesktopChrome, LinuxDesktopFirefox,
}

// Marka bazlı cihaz haritası
var devicesByBrand = map[DeviceBrand][]DeviceProfile{
	BrandApple:   {IPhone13Pro, IPhone14ProMax, IPhone15, IPadPro, IPadAir},
	BrandSamsung: {SamsungGalaxyS21, SamsungGalaxyS23Ultra, SamsungGalaxyA54, SamsungGalaxyTabS8},
	BrandGoogle:  {GooglePixel6, GooglePixel7Pro, GooglePixel8},
	BrandXiaomi:  {XiaomiMi11, Xiaomi13Pro, RedmiNote12},
	BrandHuawei:  {HuaweiP40Pro, HuaweiMate50Pro},
	BrandOnePlus: {OnePlus11, OnePlus9Pro},
	BrandWindows: {WindowsDesktopChrome, WindowsDesktopFirefox, WindowsDesktopEdge, Windows11Desktop},
	BrandMac:     {MacDesktopSafari, MacDesktopChrome, MacBookPro},
	BrandLinux:   {LinuxDesktopChrome, LinuxDesktopFirefox},
}

// Tip bazlı cihaz haritası
var devicesByType = map[DeviceType][]DeviceProfile{
	DeviceTypeMobile: {
		IPhone13Pro, IPhone14ProMax, IPhone15,
		SamsungGalaxyS21, SamsungGalaxyS23Ultra, SamsungGalaxyA54,
		GooglePixel6, GooglePixel7Pro, GooglePixel8,
		XiaomiMi11, Xiaomi13Pro, RedmiNote12,
		HuaweiP40Pro, HuaweiMate50Pro,
		OnePlus11, OnePlus9Pro,
	},
	DeviceTypeTablet: {IPadPro, IPadAir, SamsungGalaxyTabS8},
	DeviceTypeDesktop: {
		WindowsDesktopChrome, WindowsDesktopFirefox, WindowsDesktopEdge, Windows11Desktop,
		MacDesktopSafari, MacDesktopChrome, MacBookPro,
		LinuxDesktopChrome, LinuxDesktopFirefox,
	},
}

var mobileRng = rand.New(rand.NewSource(time.Now().UnixNano()))
var mobileMu sync.Mutex

func mobileRandInt(max int) int {
	mobileMu.Lock()
	defer mobileMu.Unlock()
	if max <= 0 {
		return 0
	}
	return mobileRng.Intn(max)
}

// GetAllDevices tüm cihazları döner
func GetAllDevices() []DeviceProfile {
	return append([]DeviceProfile{}, allDevices...)
}

// GetRandomDevice rastgele cihaz
func GetRandomDevice() DeviceProfile {
	return allDevices[mobileRandInt(len(allDevices))]
}

// GetDevicesByPlatform platforma göre filtreler
func GetDevicesByPlatform(platform string) []DeviceProfile {
	var out []DeviceProfile
	for _, d := range allDevices {
		if d.Platform == platform {
			out = append(out, d)
		}
	}
	return out
}

// GetDevicesByBrand markaya göre cihazları döner
func GetDevicesByBrand(brand DeviceBrand) []DeviceProfile {
	if devices, ok := devicesByBrand[brand]; ok {
		return append([]DeviceProfile{}, devices...)
	}
	return nil
}

// GetDevicesByBrands birden fazla markaya göre cihazları döner
func GetDevicesByBrands(brands []string) []DeviceProfile {
	var out []DeviceProfile
	for _, brandStr := range brands {
		brand := DeviceBrand(strings.ToLower(brandStr))
		if devices, ok := devicesByBrand[brand]; ok {
			out = append(out, devices...)
		}
	}
	if len(out) == 0 {
		return allDevices
	}
	return out
}

// GetDevicesByType tipe göre cihazları döner
func GetDevicesByType(deviceType DeviceType) []DeviceProfile {
	if devices, ok := devicesByType[deviceType]; ok {
		return append([]DeviceProfile{}, devices...)
	}
	return nil
}

// GetDevicesByTypeString string tipine göre cihazları döner
func GetDevicesByTypeString(deviceType string) []DeviceProfile {
	dt := DeviceType(strings.ToLower(deviceType))
	if dt == DeviceTypeMixed || dt == "" {
		return allDevices
	}
	return GetDevicesByType(dt)
}

// GetRandomDeviceByType tipe göre rastgele cihaz
func GetRandomDeviceByType(deviceType DeviceType) DeviceProfile {
	devices := GetDevicesByType(deviceType)
	if len(devices) == 0 {
		return GetRandomDevice()
	}
	return devices[mobileRandInt(len(devices))]
}

// GetRandomDeviceByBrand markaya göre rastgele cihaz
func GetRandomDeviceByBrand(brand DeviceBrand) DeviceProfile {
	devices := GetDevicesByBrand(brand)
	if len(devices) == 0 {
		return GetRandomDevice()
	}
	return devices[mobileRandInt(len(devices))]
}

// GetRandomDeviceFiltered tip ve markalara göre filtrelenmiş rastgele cihaz
func GetRandomDeviceFiltered(deviceType string, brands []string) DeviceProfile {
	var candidates []DeviceProfile

	// Önce tipe göre filtrele
	dt := DeviceType(strings.ToLower(deviceType))
	if dt == DeviceTypeMixed || dt == "" {
		candidates = allDevices
	} else {
		candidates = GetDevicesByType(dt)
	}

	// Sonra markalara göre filtrele
	if len(brands) > 0 {
		var filtered []DeviceProfile
		brandSet := make(map[DeviceBrand]bool)
		for _, b := range brands {
			brandSet[DeviceBrand(strings.ToLower(b))] = true
		}
		for _, d := range candidates {
			if brandSet[d.Brand] {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) > 0 {
			candidates = filtered
		}
	}

	if len(candidates) == 0 {
		return GetRandomDevice()
	}
	return candidates[mobileRandInt(len(candidates))]
}

// GetAvailableBrands mevcut markaları döner
func GetAvailableBrands() []DeviceBrand {
	return []DeviceBrand{
		BrandApple, BrandSamsung, BrandGoogle, BrandXiaomi,
		BrandHuawei, BrandOnePlus, BrandWindows, BrandMac, BrandLinux,
	}
}

// GetAvailableTypes mevcut tipleri döner
func GetAvailableTypes() []DeviceType {
	return []DeviceType{DeviceTypeDesktop, DeviceTypeMobile, DeviceTypeTablet, DeviceTypeMixed}
}
