package behavior

import (
	"math/rand"
	"sync"
	"time"
)

// ScrollType scroll davranış tipi
type ScrollType string

const (
	ScrollSmooth     ScrollType = "smooth"
	ScrollJumpy      ScrollType = "jumpy"
	ScrollSlow       ScrollType = "slow"
	ScrollAggressive ScrollType = "aggressive"
)

// MouseType mouse hareket tipi
type MouseType string

const (
	MouseDirect   MouseType = "direct"
	MouseCurved   MouseType = "curved"
	MouseHesitant MouseType = "hesitant"
)

// ClickType tıklama davranış tipi
type ClickType string

const (
	ClickPrecise ClickType = "precise"
	ClickRandom  ClickType = "random"
	ClickCareful ClickType = "careful"
)

// BehavioralProfile her ziyaretçi için benzersiz davranış profili
type BehavioralProfile struct {
	// Zamanlama
	ReadingSpeed int           // 150-400 WPM (Words Per Minute)
	TypingSpeed  int           // 200-600 CPM (Characters Per Minute)
	DecisionTime time.Duration // 1-5 saniye arası karar verme süresi

	// Scroll davranışı
	ScrollPattern ScrollType    // "smooth", "jumpy", "slow", "aggressive"
	ScrollPause   time.Duration // Scroll arası bekleme

	// Mouse davranışı
	MousePattern MouseType // "direct", "curved", "hesitant"
	MouseSpeed   int       // 1-10 hız

	// Tıklama davranışı
	ClickPattern    ClickType // "precise", "random", "careful"
	DoubleClickProb float64   // 0.0-0.1

	// Dikkat süresi
	AttentionSpan time.Duration
	TabSwitchProb float64 // Sekme değiştirme olasılığı

	// Ek davranış özellikleri
	ScrollReverseProb float64 // Scroll sırasında yukarı dönme olasılığı
	MousePauseProb    float64 // Mouse hareketinde duraklama olasılığı
	ReadingPauseProb  float64 // Okurken duraklama olasılığı
}

// FingerprintGenerator davranış parmak izi üretici
type FingerprintGenerator struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewFingerprintGenerator yeni bir fingerprint generator oluşturur
func NewFingerprintGenerator() *FingerprintGenerator {
	return &FingerprintGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomProfile rastgele bir davranış profili oluşturur
func GenerateRandomProfile() *BehavioralProfile {
	gen := NewFingerprintGenerator()
	return gen.GenerateProfile()
}

// GenerateProfile rastgele davranış profili oluşturur
func (fg *FingerprintGenerator) GenerateProfile() *BehavioralProfile {
	fg.mu.Lock()
	defer fg.mu.Unlock()

	return &BehavioralProfile{
		// Zamanlama (insan varyasyonuna göre)
		ReadingSpeed: fg.randomInt(150, 400),
		TypingSpeed:  fg.randomInt(200, 600),
		DecisionTime: fg.randomDuration(1*time.Second, 5*time.Second),

		// Scroll davranışı
		ScrollPattern: fg.randomScrollType(),
		ScrollPause:   fg.randomDuration(200*time.Millisecond, 2000*time.Millisecond),

		// Mouse davranışı
		MousePattern: fg.randomMouseType(),
		MouseSpeed:   fg.randomInt(1, 10),

		// Tıklama davranışı
		ClickPattern:    fg.randomClickType(),
		DoubleClickProb: fg.randomFloat(0.0, 0.1),

		// Dikkat süresi
		AttentionSpan: fg.randomDuration(5*time.Second, 120*time.Second),
		TabSwitchProb: fg.randomFloat(0.0, 0.3),

		// Ek davranışlar
		ScrollReverseProb: fg.randomFloat(0.05, 0.25),
		MousePauseProb:    fg.randomFloat(0.1, 0.4),
		ReadingPauseProb:  fg.randomFloat(0.2, 0.5),
	}
}

// GenerateProfileFromBase belirli bir profil tipine dayalı profil oluşturur
func (fg *FingerprintGenerator) GenerateProfileFromBase(baseType ProfileType) *BehavioralProfile {
	baseProfile := GetProfile(baseType)
	profile := fg.GenerateProfile()

	// Base profile özelliklerini uygula
	switch baseType {
	case ProfileFastReader:
		profile.ReadingSpeed = fg.randomInt(300, 400)
		profile.DecisionTime = fg.randomDuration(500*time.Millisecond, 1500*time.Millisecond)
		profile.AttentionSpan = fg.randomDuration(5*time.Second, 20*time.Second)
		profile.ScrollPattern = ScrollFast
		profile.MouseSpeed = fg.randomInt(7, 10)

	case ProfileDetailedReader:
		profile.ReadingSpeed = fg.randomInt(150, 250)
		profile.DecisionTime = fg.randomDuration(3*time.Second, 5*time.Second)
		profile.AttentionSpan = fg.randomDuration(60*time.Second, 300*time.Second)
		profile.ScrollPattern = ScrollSlow
		profile.MousePattern = MouseHesitant
		profile.ClickPattern = ClickCareful

	case ProfileShopper:
		profile.DecisionTime = fg.randomDuration(2*time.Second, 4*time.Second)
		profile.AttentionSpan = fg.randomDuration(30*time.Second, 90*time.Second)
		profile.ScrollPattern = ScrollSmooth
		profile.MousePattern = MouseCurved
		profile.MousePauseProb = 0.5 // Ürünleri incelemek için daha fazla duraklama

	case ProfileBouncer:
		profile.ReadingSpeed = fg.randomInt(350, 400)
		profile.DecisionTime = fg.randomDuration(200*time.Millisecond, 800*time.Millisecond)
		profile.AttentionSpan = fg.randomDuration(2*time.Second, 8*time.Second)
		profile.ScrollPattern = ScrollAggressive
		profile.MouseSpeed = fg.randomInt(8, 10)

	case ProfileMobile:
		profile.MousePattern = MouseDirect // Touch ekran
		profile.MouseSpeed = fg.randomInt(5, 8)
		profile.ScrollPattern = ScrollJumpy
		profile.ScrollPause = fg.randomDuration(100*time.Millisecond, 800*time.Millisecond)
	}

	// Base profile click probability'yi kullan
	_ = baseProfile.ClickProbability

	return profile
}

// randomInt [min, max] aralığında rastgele int döner
func (fg *FingerprintGenerator) randomInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + fg.rng.Intn(max-min+1)
}

// randomFloat [min, max] aralığında rastgele float64 döner
func (fg *FingerprintGenerator) randomFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + fg.rng.Float64()*(max-min)
}

// randomDuration [min, max] aralığında rastgele duration döner
func (fg *FingerprintGenerator) randomDuration(min, max time.Duration) time.Duration {
	if min >= max {
		return min
	}
	diff := max - min
	return min + time.Duration(fg.rng.Int63n(int64(diff)))
}

// randomScrollType rastgele scroll tipi döner
func (fg *FingerprintGenerator) randomScrollType() ScrollType {
	types := []ScrollType{ScrollSmooth, ScrollJumpy, ScrollSlow, ScrollAggressive}
	return types[fg.rng.Intn(len(types))]
}

// randomMouseType rastgele mouse tipi döner
func (fg *FingerprintGenerator) randomMouseType() MouseType {
	types := []MouseType{MouseDirect, MouseCurved, MouseHesitant}
	weights := []int{30, 50, 20} // curved daha yaygın

	total := 0
	for _, w := range weights {
		total += w
	}

	roll := fg.rng.Intn(total)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return types[i]
		}
	}
	return MouseCurved
}

// randomClickType rastgele click tipi döner
func (fg *FingerprintGenerator) randomClickType() ClickType {
	types := []ClickType{ClickPrecise, ClickRandom, ClickCareful}
	weights := []int{60, 15, 25} // precise daha yaygın

	total := 0
	for _, w := range weights {
		total += w
	}

	roll := fg.rng.Intn(total)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return types[i]
		}
	}
	return ClickPrecise
}

// GetScrollSpeedMultiplier scroll tipine göre hız çarpanı döner
func (bp *BehavioralProfile) GetScrollSpeedMultiplier() float64 {
	switch bp.ScrollPattern {
	case ScrollSmooth:
		return 1.0
	case ScrollJumpy:
		return 1.5 + (rand.Float64() * 0.5)
	case ScrollSlow:
		return 0.5 + (rand.Float64() * 0.3)
	case ScrollAggressive:
		return 2.0 + (rand.Float64() * 0.5)
	default:
		return 1.0
	}
}

// GetMouseSpeedMultiplier mouse hız çarpanı döner
func (bp *BehavioralProfile) GetMouseSpeedMultiplier() float64 {
	// MouseSpeed 1-10 arası
	base := float64(bp.MouseSpeed) / 5.0

	switch bp.MousePattern {
	case MouseDirect:
		return base * 1.2
	case MouseCurved:
		return base * 1.0
	case MouseHesitant:
		return base * 0.7
	default:
		return base
	}
}

// ShouldDoubleClick double click yapmalı mı
func (bp *BehavioralProfile) ShouldDoubleClick() bool {
	return rand.Float64() < bp.DoubleClickProb
}

// ShouldTabSwitch sekme değiştirmeli mi
func (bp *BehavioralProfile) ShouldTabSwitch() bool {
	return rand.Float64() < bp.TabSwitchProb
}

// ShouldScrollReverse scroll sırasında yukarı dönmeli mi
func (bp *BehavioralProfile) ShouldScrollReverse() bool {
	return rand.Float64() < bp.ScrollReverseProb
}

// ShouldMousePause mouse hareketinde duraklamalı mı
func (bp *BehavioralProfile) ShouldMousePause() bool {
	return rand.Float64() < bp.MousePauseProb
}

// ShouldReadingPause okurken duraklamalı mı
func (bp *BehavioralProfile) ShouldReadingPause() bool {
	return rand.Float64() < bp.ReadingPauseProb
}

// GetReadingTime metin uzunluğuna göre okuma süresi hesaplar
func (bp *BehavioralProfile) GetReadingTime(textLength int) time.Duration {
	wordCount := textLength / 5
	if wordCount <= 0 {
		wordCount = 50
	}

	// WPM'den saniyeye çevir
	seconds := (wordCount * 60) / bp.ReadingSpeed

	// Min/max sınırları
	if seconds < 3 {
		seconds = 3
	}
	if seconds > 300 {
		seconds = 300
	}

	return time.Duration(seconds) * time.Second
}

// GetTypingDelay bir karakter yazma gecikmesi
func (bp *BehavioralProfile) GetTypingDelay() time.Duration {
	// CPM'den karakter başına ms hesapla
	cpm := bp.TypingSpeed
	if cpm <= 0 {
		cpm = 300
	}

	msPerChar := 60000 / cpm

	// Doğal varyasyon ekle
	variation := float64(msPerChar) * 0.3
	delay := float64(msPerChar) + (rand.Float64()*variation - variation/2)

	return time.Duration(delay) * time.Millisecond
}

// GetDecisionPause karar verme bekleme süresi
func (bp *BehavioralProfile) GetDecisionPause() time.Duration {
	// DecisionTime etrafında varyasyon
	variation := float64(bp.DecisionTime) * 0.4
	pause := float64(bp.DecisionTime) + (rand.Float64()*variation - variation/2)

	if pause < 0 {
		pause = 100 // minimum 100ms
	}

	return time.Duration(pause)
}

// ProfilePool önceden oluşturulmuş profiller havuzu
type ProfilePool struct {
	profiles []*BehavioralProfile
	index    int
	mu       sync.Mutex
	rng      *rand.Rand
}

// NewProfilePool yeni profil havuzu oluşturur
func NewProfilePool(size int) *ProfilePool {
	pool := &ProfilePool{
		profiles: make([]*BehavioralProfile, size),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for i := 0; i < size; i++ {
		pool.profiles[i] = GenerateRandomProfile()
	}

	return pool
}

// GetNext sıradaki profili döner (rotasyon)
func (pp *ProfilePool) GetNext() *BehavioralProfile {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	profile := pp.profiles[pp.index]
	pp.index = (pp.index + 1) % len(pp.profiles)
	return profile
}

// GetRandom rastgele profil döner
func (pp *ProfilePool) GetRandom() *BehavioralProfile {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	return pp.profiles[pp.rng.Intn(len(pp.profiles))]
}

// Shuffle profilleri karıştırır
func (pp *ProfilePool) Shuffle() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.rng.Shuffle(len(pp.profiles), func(i, j int) {
		pp.profiles[i], pp.profiles[j] = pp.profiles[j], pp.profiles[i]
	})
}

// GenerateMixedProfiles farklı profil tiplerinden karışık profiller oluşturur
func GenerateMixedProfiles(count int, mixRatio map[ProfileType]float64) []*BehavioralProfile {
	if len(mixRatio) == 0 {
		// Varsayılan dağılım
		mixRatio = map[ProfileType]float64{
			ProfileFastReader:     0.25,
			ProfileDetailedReader: 0.15,
			ProfileShopper:        0.20,
			ProfileResearcher:     0.15,
			ProfileBouncer:        0.10,
			ProfileEngaged:        0.10,
			ProfileMobile:         0.05,
		}
	}

	gen := NewFingerprintGenerator()
	profiles := make([]*BehavioralProfile, count)

	for i := 0; i < count; i++ {
		// Rastgele bir profil tipi seç
		profileType := weightedRandomProfileType(gen.rng, mixRatio)
		profiles[i] = gen.GenerateProfileFromBase(profileType)
	}

	return profiles
}

// weightedRandomProfileType ağırlıklı rastgele profil tipi seçer
func weightedRandomProfileType(rng *rand.Rand, weights map[ProfileType]float64) ProfileType {
	total := 0.0
	for _, w := range weights {
		total += w
	}

	if total <= 0 {
		return ProfileFastReader
	}

	roll := rng.Float64() * total
	cumulative := 0.0

	for profileType, weight := range weights {
		cumulative += weight
		if roll < cumulative {
			return profileType
		}
	}

	return ProfileFastReader
}
