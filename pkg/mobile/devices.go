package mobile

import (
	"math/rand"
	"sync"
	"time"
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
}

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
	}
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
	}
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
	}
	iPadPro = DeviceProfile{
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
	}
)

var deviceList = []DeviceProfile{IPhone13Pro, IPhone14ProMax, SamsungGalaxyS21, GooglePixel6, iPadPro}
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
	return append([]DeviceProfile{}, deviceList...)
}

// GetRandomDevice rastgele cihaz
func GetRandomDevice() DeviceProfile {
	return deviceList[mobileRandInt(len(deviceList))]
}

// GetDevicesByPlatform platforma göre filtreler
func GetDevicesByPlatform(platform string) []DeviceProfile {
	var out []DeviceProfile
	for _, d := range deviceList {
		if d.Platform == platform {
			out = append(out, d)
		}
	}
	return out
}
