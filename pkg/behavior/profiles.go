package behavior

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// ProfileType davranış profili tipi
type ProfileType string

const (
	ProfileFastReader     ProfileType = "fast_reader"     // Hızlı okuyucu
	ProfileDetailedReader ProfileType = "detailed_reader" // Detaylı okuyucu
	ProfileShopper        ProfileType = "shopper"         // Alışverişçi
	ProfileResearcher     ProfileType = "researcher"      // Araştırmacı
	ProfileBouncer        ProfileType = "bouncer"         // Hemen çıkan
	ProfileEngaged        ProfileType = "engaged"         // İlgili kullanıcı
	ProfileMobile         ProfileType = "mobile"          // Mobil kullanıcı
	ProfileRandom         ProfileType = "random"          // Rastgele
)

// BehaviorProfile davranış profili
type BehaviorProfile struct {
	Type              ProfileType
	Name              string
	Description       string
	MinDwellTime      time.Duration // Minimum sayfa süresi
	MaxDwellTime      time.Duration // Maksimum sayfa süresi
	ScrollDepth       float64       // Scroll derinliği (0-1)
	ScrollSpeed       string        // "slow", "medium", "fast"
	ClickProbability  float64       // İç link tıklama olasılığı (0-1)
	MouseMovement     bool          // Mouse hareketi simülasyonu
	ReadingPauses     bool          // Okuma duraklamaları
	BackNavigation    float64       // Geri dönme olasılığı (0-1)
	FormInteraction   float64       // Form etkileşimi olasılığı (0-1)
	VideoWatch        float64       // Video izleme olasılığı (0-1)
	PageViews         int           // Ortalama sayfa görüntüleme
	BounceRate        float64       // Bounce rate (0-1)
}

// Önceden tanımlı profiller
var profiles = map[ProfileType]BehaviorProfile{
	ProfileFastReader: {
		Type:             ProfileFastReader,
		Name:             "Hızlı Okuyucu",
		Description:      "Sayfayı hızlıca tarar, kısa süre kalır",
		MinDwellTime:     5 * time.Second,
		MaxDwellTime:     15 * time.Second,
		ScrollDepth:      0.6,
		ScrollSpeed:      "fast",
		ClickProbability: 0.2,
		MouseMovement:    false,
		ReadingPauses:    false,
		BackNavigation:   0.1,
		FormInteraction:  0.05,
		VideoWatch:       0.1,
		PageViews:        2,
		BounceRate:       0.4,
	},
	ProfileDetailedReader: {
		Type:             ProfileDetailedReader,
		Name:             "Detaylı Okuyucu",
		Description:      "İçeriği dikkatli okur, uzun süre kalır",
		MinDwellTime:     45 * time.Second,
		MaxDwellTime:     180 * time.Second,
		ScrollDepth:      0.95,
		ScrollSpeed:      "slow",
		ClickProbability: 0.5,
		MouseMovement:    true,
		ReadingPauses:    true,
		BackNavigation:   0.3,
		FormInteraction:  0.2,
		VideoWatch:       0.6,
		PageViews:        5,
		BounceRate:       0.1,
	},
	ProfileShopper: {
		Type:             ProfileShopper,
		Name:             "Alışverişçi",
		Description:      "Ürünlere bakar, fiyat karşılaştırır",
		MinDwellTime:     20 * time.Second,
		MaxDwellTime:     60 * time.Second,
		ScrollDepth:      0.8,
		ScrollSpeed:      "medium",
		ClickProbability: 0.7,
		MouseMovement:    true,
		ReadingPauses:    true,
		BackNavigation:   0.5,
		FormInteraction:  0.3,
		VideoWatch:       0.2,
		PageViews:        8,
		BounceRate:       0.15,
	},
	ProfileResearcher: {
		Type:             ProfileResearcher,
		Name:             "Araştırmacı",
		Description:      "Bilgi toplar, birden fazla sayfa ziyaret eder",
		MinDwellTime:     30 * time.Second,
		MaxDwellTime:     120 * time.Second,
		ScrollDepth:      0.9,
		ScrollSpeed:      "medium",
		ClickProbability: 0.6,
		MouseMovement:    true,
		ReadingPauses:    true,
		BackNavigation:   0.4,
		FormInteraction:  0.1,
		VideoWatch:       0.4,
		PageViews:        6,
		BounceRate:       0.1,
	},
	ProfileBouncer: {
		Type:             ProfileBouncer,
		Name:             "Hemen Çıkan",
		Description:      "Sayfaya bakar ve hemen çıkar",
		MinDwellTime:     2 * time.Second,
		MaxDwellTime:     8 * time.Second,
		ScrollDepth:      0.2,
		ScrollSpeed:      "fast",
		ClickProbability: 0.05,
		MouseMovement:    false,
		ReadingPauses:    false,
		BackNavigation:   0.0,
		FormInteraction:  0.0,
		VideoWatch:       0.0,
		PageViews:        1,
		BounceRate:       0.95,
	},
	ProfileEngaged: {
		Type:             ProfileEngaged,
		Name:             "İlgili Kullanıcı",
		Description:      "Siteyle yoğun etkileşim kurar",
		MinDwellTime:     60 * time.Second,
		MaxDwellTime:     300 * time.Second,
		ScrollDepth:      1.0,
		ScrollSpeed:      "slow",
		ClickProbability: 0.8,
		MouseMovement:    true,
		ReadingPauses:    true,
		BackNavigation:   0.3,
		FormInteraction:  0.5,
		VideoWatch:       0.7,
		PageViews:        10,
		BounceRate:       0.05,
	},
	ProfileMobile: {
		Type:             ProfileMobile,
		Name:             "Mobil Kullanıcı",
		Description:      "Mobil cihazdan kısa oturumlar",
		MinDwellTime:     10 * time.Second,
		MaxDwellTime:     45 * time.Second,
		ScrollDepth:      0.7,
		ScrollSpeed:      "medium",
		ClickProbability: 0.4,
		MouseMovement:    false, // Touch kullanır
		ReadingPauses:    true,
		BackNavigation:   0.2,
		FormInteraction:  0.15,
		VideoWatch:       0.3,
		PageViews:        3,
		BounceRate:       0.25,
	},
}

// ProfileManager profil yöneticisi
type ProfileManager struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewProfileManager yeni profil yöneticisi oluşturur
func NewProfileManager() *ProfileManager {
	return &ProfileManager{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetProfile profil döner
func GetProfile(profileType ProfileType) BehaviorProfile {
	if profile, ok := profiles[profileType]; ok {
		return profile
	}
	return profiles[ProfileFastReader]
}

// GetRandomProfile rastgele profil döner
func (pm *ProfileManager) GetRandomProfile() BehaviorProfile {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	profileTypes := []ProfileType{
		ProfileFastReader,
		ProfileDetailedReader,
		ProfileShopper,
		ProfileResearcher,
		ProfileEngaged,
		ProfileMobile,
	}
	
	// Ağırlıklı dağılım (gerçekçi)
	weights := []int{25, 15, 20, 15, 10, 15} // Toplam 100
	
	roll := pm.rng.Intn(100)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return profiles[profileTypes[i]]
		}
	}
	
	return profiles[ProfileFastReader]
}

// GetRandomProfileWeighted ağırlıklı rastgele profil döner
func (pm *ProfileManager) GetRandomProfileWeighted(weights map[ProfileType]int) BehaviorProfile {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	total := 0
	for _, w := range weights {
		total += w
	}
	
	if total == 0 {
		return pm.GetRandomProfile()
	}
	
	roll := pm.rng.Intn(total)
	cumulative := 0
	for profileType, w := range weights {
		cumulative += w
		if roll < cumulative {
			if profile, ok := profiles[profileType]; ok {
				return profile
			}
		}
	}
	
	return profiles[ProfileFastReader]
}

// GetDwellTime profil için dwell time hesaplar
func (p *BehaviorProfile) GetDwellTime() time.Duration {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	diff := p.MaxDwellTime - p.MinDwellTime
	return p.MinDwellTime + time.Duration(rng.Int63n(int64(diff)))
}

// ShouldClick tıklama yapılmalı mı
func (p *BehaviorProfile) ShouldClick() bool {
	return rand.Float64() < p.ClickProbability
}

// ShouldBounce bounce olmalı mı
func (p *BehaviorProfile) ShouldBounce() bool {
	return rand.Float64() < p.BounceRate
}

// ShouldWatchVideo video izlemeli mi
func (p *BehaviorProfile) ShouldWatchVideo() bool {
	return rand.Float64() < p.VideoWatch
}

// ShouldInteractForm form etkileşimi olmalı mı
func (p *BehaviorProfile) ShouldInteractForm() bool {
	return rand.Float64() < p.FormInteraction
}

// ShouldNavigateBack geri dönmeli mi
func (p *BehaviorProfile) ShouldNavigateBack() bool {
	return rand.Float64() < p.BackNavigation
}

// SimulateBehavior profil davranışını simüle eder
func (p *BehaviorProfile) SimulateBehavior(ctx context.Context) error {
	dwellTime := p.GetDwellTime()
	start := time.Now()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Scroll davranışı
	scrollSteps := int(p.ScrollDepth * 10)
	if scrollSteps < 1 {
		scrollSteps = 1
	}
	
	var scrollDelay time.Duration
	switch p.ScrollSpeed {
	case "slow":
		scrollDelay = 800 * time.Millisecond
	case "fast":
		scrollDelay = 200 * time.Millisecond
	default:
		scrollDelay = 400 * time.Millisecond
	}
	
	for i := 0; i < scrollSteps && time.Since(start) < dwellTime; i++ {
		// Scroll
		scrollAmount := 100 + rng.Intn(200)
		script := `window.scrollBy(0, ` + string(rune(scrollAmount)) + `)`
		
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			// Scroll hatası kritik değil
			_ = err
		}
		
		// Okuma duraklaması
		if p.ReadingPauses && rng.Float64() < 0.3 {
			pauseTime := time.Duration(1000+rng.Intn(3000)) * time.Millisecond
			time.Sleep(pauseTime)
		} else {
			jitter := time.Duration(rng.Intn(200)) * time.Millisecond
			time.Sleep(scrollDelay + jitter)
		}
		
		// Bazen yukarı scroll
		if rng.Float64() < 0.15 {
			upScroll := 30 + rng.Intn(70)
			upScript := `window.scrollBy(0, -` + string(rune(upScroll)) + `)`
			_ = chromedp.Run(ctx, chromedp.Evaluate(upScript, nil))
		}
	}
	
	// Kalan süreyi bekle
	remaining := dwellTime - time.Since(start)
	if remaining > 0 {
		time.Sleep(remaining)
	}
	
	return nil
}

// GetAllProfiles tüm profilleri döner
func GetAllProfiles() map[ProfileType]BehaviorProfile {
	result := make(map[ProfileType]BehaviorProfile)
	for k, v := range profiles {
		result[k] = v
	}
	return result
}

// GetProfileTypes profil tiplerini döner
func GetProfileTypes() []ProfileType {
	return []ProfileType{
		ProfileFastReader,
		ProfileDetailedReader,
		ProfileShopper,
		ProfileResearcher,
		ProfileBouncer,
		ProfileEngaged,
		ProfileMobile,
	}
}
