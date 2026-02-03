package geo

import (
	"math/rand"
	"sync"
	"time"
)

// GeoLocation coğrafi konum bilgisi
type GeoLocation struct {
	Country       string   // Ülke kodu (TR, US, DE, vb.)
	CountryName   string   // Ülke adı
	City          string   // Şehir
	Region        string   // Bölge/Eyalet
	Timezone      string   // Timezone (Europe/Istanbul)
	Language      string   // Dil kodu (tr, en, de)
	Languages     []string // Dil listesi
	AcceptLang    string   // Accept-Language header
	Locale        string   // Locale (tr-TR, en-US)
	CurrencyCode  string   // Para birimi (TRY, USD, EUR)
	LatLng        [2]float64 // Enlem, Boylam (yaklaşık)
}

// GeoConfig coğrafi hedefleme yapılandırması
type GeoConfig struct {
	Countries     []string // Hedef ülkeler
	Cities        []string // Hedef şehirler (opsiyonel)
	Distribution  map[string]int // Ülke dağılımı (yüzde)
}

// GeoManager coğrafi hedefleme yöneticisi
type GeoManager struct {
	config GeoConfig
	mu     sync.Mutex
	rng    *rand.Rand
}

// Ülke veritabanı
var countryDB = map[string]GeoLocation{
	"TR": {
		Country:      "TR",
		CountryName:  "Turkey",
		Timezone:     "Europe/Istanbul",
		Language:     "tr",
		Languages:    []string{"tr-TR", "tr", "en"},
		AcceptLang:   "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:       "tr-TR",
		CurrencyCode: "TRY",
	},
	"US": {
		Country:      "US",
		CountryName:  "United States",
		Timezone:     "America/New_York",
		Language:     "en",
		Languages:    []string{"en-US", "en"},
		AcceptLang:   "en-US,en;q=0.9",
		Locale:       "en-US",
		CurrencyCode: "USD",
	},
	"GB": {
		Country:      "GB",
		CountryName:  "United Kingdom",
		Timezone:     "Europe/London",
		Language:     "en",
		Languages:    []string{"en-GB", "en"},
		AcceptLang:   "en-GB,en;q=0.9",
		Locale:       "en-GB",
		CurrencyCode: "GBP",
	},
	"DE": {
		Country:      "DE",
		CountryName:  "Germany",
		Timezone:     "Europe/Berlin",
		Language:     "de",
		Languages:    []string{"de-DE", "de", "en"},
		AcceptLang:   "de-DE,de;q=0.9,en;q=0.8",
		Locale:       "de-DE",
		CurrencyCode: "EUR",
	},
	"FR": {
		Country:      "FR",
		CountryName:  "France",
		Timezone:     "Europe/Paris",
		Language:     "fr",
		Languages:    []string{"fr-FR", "fr", "en"},
		AcceptLang:   "fr-FR,fr;q=0.9,en;q=0.8",
		Locale:       "fr-FR",
		CurrencyCode: "EUR",
	},
	"NL": {
		Country:      "NL",
		CountryName:  "Netherlands",
		Timezone:     "Europe/Amsterdam",
		Language:     "nl",
		Languages:    []string{"nl-NL", "nl", "en"},
		AcceptLang:   "nl-NL,nl;q=0.9,en;q=0.8",
		Locale:       "nl-NL",
		CurrencyCode: "EUR",
	},
	"IT": {
		Country:      "IT",
		CountryName:  "Italy",
		Timezone:     "Europe/Rome",
		Language:     "it",
		Languages:    []string{"it-IT", "it", "en"},
		AcceptLang:   "it-IT,it;q=0.9,en;q=0.8",
		Locale:       "it-IT",
		CurrencyCode: "EUR",
	},
	"ES": {
		Country:      "ES",
		CountryName:  "Spain",
		Timezone:     "Europe/Madrid",
		Language:     "es",
		Languages:    []string{"es-ES", "es", "en"},
		AcceptLang:   "es-ES,es;q=0.9,en;q=0.8",
		Locale:       "es-ES",
		CurrencyCode: "EUR",
	},
	"RU": {
		Country:      "RU",
		CountryName:  "Russia",
		Timezone:     "Europe/Moscow",
		Language:     "ru",
		Languages:    []string{"ru-RU", "ru", "en"},
		AcceptLang:   "ru-RU,ru;q=0.9,en;q=0.8",
		Locale:       "ru-RU",
		CurrencyCode: "RUB",
	},
	"JP": {
		Country:      "JP",
		CountryName:  "Japan",
		Timezone:     "Asia/Tokyo",
		Language:     "ja",
		Languages:    []string{"ja-JP", "ja", "en"},
		AcceptLang:   "ja-JP,ja;q=0.9,en;q=0.8",
		Locale:       "ja-JP",
		CurrencyCode: "JPY",
	},
	"CN": {
		Country:      "CN",
		CountryName:  "China",
		Timezone:     "Asia/Shanghai",
		Language:     "zh",
		Languages:    []string{"zh-CN", "zh", "en"},
		AcceptLang:   "zh-CN,zh;q=0.9,en;q=0.8",
		Locale:       "zh-CN",
		CurrencyCode: "CNY",
	},
	"KR": {
		Country:      "KR",
		CountryName:  "South Korea",
		Timezone:     "Asia/Seoul",
		Language:     "ko",
		Languages:    []string{"ko-KR", "ko", "en"},
		AcceptLang:   "ko-KR,ko;q=0.9,en;q=0.8",
		Locale:       "ko-KR",
		CurrencyCode: "KRW",
	},
	"BR": {
		Country:      "BR",
		CountryName:  "Brazil",
		Timezone:     "America/Sao_Paulo",
		Language:     "pt",
		Languages:    []string{"pt-BR", "pt", "en"},
		AcceptLang:   "pt-BR,pt;q=0.9,en;q=0.8",
		Locale:       "pt-BR",
		CurrencyCode: "BRL",
	},
	"IN": {
		Country:      "IN",
		CountryName:  "India",
		Timezone:     "Asia/Kolkata",
		Language:     "en",
		Languages:    []string{"en-IN", "hi", "en"},
		AcceptLang:   "en-IN,en;q=0.9,hi;q=0.8",
		Locale:       "en-IN",
		CurrencyCode: "INR",
	},
	"AU": {
		Country:      "AU",
		CountryName:  "Australia",
		Timezone:     "Australia/Sydney",
		Language:     "en",
		Languages:    []string{"en-AU", "en"},
		AcceptLang:   "en-AU,en;q=0.9",
		Locale:       "en-AU",
		CurrencyCode: "AUD",
	},
	"CA": {
		Country:      "CA",
		CountryName:  "Canada",
		Timezone:     "America/Toronto",
		Language:     "en",
		Languages:    []string{"en-CA", "en", "fr"},
		AcceptLang:   "en-CA,en;q=0.9,fr;q=0.8",
		Locale:       "en-CA",
		CurrencyCode: "CAD",
	},
	"MX": {
		Country:      "MX",
		CountryName:  "Mexico",
		Timezone:     "America/Mexico_City",
		Language:     "es",
		Languages:    []string{"es-MX", "es", "en"},
		AcceptLang:   "es-MX,es;q=0.9,en;q=0.8",
		Locale:       "es-MX",
		CurrencyCode: "MXN",
	},
	"PL": {
		Country:      "PL",
		CountryName:  "Poland",
		Timezone:     "Europe/Warsaw",
		Language:     "pl",
		Languages:    []string{"pl-PL", "pl", "en"},
		AcceptLang:   "pl-PL,pl;q=0.9,en;q=0.8",
		Locale:       "pl-PL",
		CurrencyCode: "PLN",
	},
	"UA": {
		Country:      "UA",
		CountryName:  "Ukraine",
		Timezone:     "Europe/Kiev",
		Language:     "uk",
		Languages:    []string{"uk-UA", "uk", "ru", "en"},
		AcceptLang:   "uk-UA,uk;q=0.9,ru;q=0.8,en;q=0.7",
		Locale:       "uk-UA",
		CurrencyCode: "UAH",
	},
	"SA": {
		Country:      "SA",
		CountryName:  "Saudi Arabia",
		Timezone:     "Asia/Riyadh",
		Language:     "ar",
		Languages:    []string{"ar-SA", "ar", "en"},
		AcceptLang:   "ar-SA,ar;q=0.9,en;q=0.8",
		Locale:       "ar-SA",
		CurrencyCode: "SAR",
	},
	"AE": {
		Country:      "AE",
		CountryName:  "United Arab Emirates",
		Timezone:     "Asia/Dubai",
		Language:     "ar",
		Languages:    []string{"ar-AE", "ar", "en"},
		AcceptLang:   "ar-AE,ar;q=0.9,en;q=0.8",
		Locale:       "ar-AE",
		CurrencyCode: "AED",
	},
}

// Türkiye şehirleri
var turkishCities = map[string]GeoLocation{
	"Istanbul": {
		Country:     "TR",
		CountryName: "Turkey",
		City:        "Istanbul",
		Region:      "Marmara",
		Timezone:    "Europe/Istanbul",
		Language:    "tr",
		Languages:   []string{"tr-TR", "tr", "en"},
		AcceptLang:  "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:      "tr-TR",
		LatLng:      [2]float64{41.0082, 28.9784},
	},
	"Ankara": {
		Country:     "TR",
		CountryName: "Turkey",
		City:        "Ankara",
		Region:      "Central Anatolia",
		Timezone:    "Europe/Istanbul",
		Language:    "tr",
		Languages:   []string{"tr-TR", "tr", "en"},
		AcceptLang:  "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:      "tr-TR",
		LatLng:      [2]float64{39.9334, 32.8597},
	},
	"Izmir": {
		Country:     "TR",
		CountryName: "Turkey",
		City:        "Izmir",
		Region:      "Aegean",
		Timezone:    "Europe/Istanbul",
		Language:    "tr",
		Languages:   []string{"tr-TR", "tr", "en"},
		AcceptLang:  "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:      "tr-TR",
		LatLng:      [2]float64{38.4237, 27.1428},
	},
	"Antalya": {
		Country:     "TR",
		CountryName: "Turkey",
		City:        "Antalya",
		Region:      "Mediterranean",
		Timezone:    "Europe/Istanbul",
		Language:    "tr",
		Languages:   []string{"tr-TR", "tr", "en"},
		AcceptLang:  "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:      "tr-TR",
		LatLng:      [2]float64{36.8969, 30.7133},
	},
	"Bursa": {
		Country:     "TR",
		CountryName: "Turkey",
		City:        "Bursa",
		Region:      "Marmara",
		Timezone:    "Europe/Istanbul",
		Language:    "tr",
		Languages:   []string{"tr-TR", "tr", "en"},
		AcceptLang:  "tr-TR,tr;q=0.9,en;q=0.8",
		Locale:      "tr-TR",
		LatLng:      [2]float64{40.1885, 29.0610},
	},
}

// NewGeoManager yeni geo manager oluşturur
func NewGeoManager(config GeoConfig) *GeoManager {
	return &GeoManager{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomLocation rastgele bir konum döner
func (g *GeoManager) GetRandomLocation() GeoLocation {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Dağılım varsa ona göre seç
	if len(g.config.Distribution) > 0 {
		roll := g.rng.Intn(100)
		cumulative := 0
		for country, percent := range g.config.Distribution {
			cumulative += percent
			if roll < cumulative {
				if loc, ok := countryDB[country]; ok {
					return loc
				}
			}
		}
	}
	
	// Ülke listesi varsa ondan seç
	if len(g.config.Countries) > 0 {
		country := g.config.Countries[g.rng.Intn(len(g.config.Countries))]
		if loc, ok := countryDB[country]; ok {
			return loc
		}
	}
	
	// Varsayılan: Türkiye
	return countryDB["TR"]
}

// GetLocationByCountry ülke koduna göre konum döner
func GetLocationByCountry(countryCode string) (GeoLocation, bool) {
	loc, ok := countryDB[countryCode]
	return loc, ok
}

// GetTurkishCity Türkiye şehri döner
func GetTurkishCity(cityName string) (GeoLocation, bool) {
	loc, ok := turkishCities[cityName]
	return loc, ok
}

// GetRandomTurkishCity rastgele Türkiye şehri döner
func GetRandomTurkishCity() GeoLocation {
	cities := make([]GeoLocation, 0, len(turkishCities))
	for _, city := range turkishCities {
		cities = append(cities, city)
	}
	if len(cities) == 0 {
		return countryDB["TR"]
	}
	return cities[rand.Intn(len(cities))]
}

// GetAvailableCountries mevcut ülkeleri döner
func GetAvailableCountries() []string {
	countries := make([]string, 0, len(countryDB))
	for code := range countryDB {
		countries = append(countries, code)
	}
	return countries
}

// GetAvailableTurkishCities mevcut Türkiye şehirlerini döner
func GetAvailableTurkishCities() []string {
	cities := make([]string, 0, len(turkishCities))
	for name := range turkishCities {
		cities = append(cities, name)
	}
	return cities
}

// ApplyToFingerprint fingerprint'e geo bilgilerini uygular
func (loc *GeoLocation) ApplyToFingerprint() map[string]interface{} {
	return map[string]interface{}{
		"timezone":    loc.Timezone,
		"language":    loc.Language,
		"languages":   loc.Languages,
		"locale":      loc.Locale,
		"acceptLang":  loc.AcceptLang,
	}
}
