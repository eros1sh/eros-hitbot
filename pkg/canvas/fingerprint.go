package canvas

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/chromedp/chromedp"
)

// CanvasFingerprint represents unique canvas properties
type CanvasFingerprint struct {
	CanvasNoise   float64  // 0.0001 - 0.001 arası gürültü
	WebGLVendor   string
	WebGLRenderer string
	AudioNoise    float64
	Fonts         []string
}

// GenerateFingerprint creates a unique fingerprint
func GenerateFingerprint() *CanvasFingerprint {
	return &CanvasFingerprint{
		CanvasNoise:   randomNoise(0.0001, 0.001),
		WebGLVendor:   randomWebGLVendor(),
		WebGLRenderer: randomWebGLRenderer(),
		AudioNoise:    randomNoise(0.00001, 0.0001),
		Fonts:         randomFonts(),
	}
}

// InjectCanvasNoise injects canvas fingerprinting noise (page load öncesi script)
func (cf *CanvasFingerprint) InjectCanvasNoise(ctx context.Context) error {
	script := fmt.Sprintf(`(function(){
		var originalGetImageData = CanvasRenderingContext2D.prototype.getImageData;
		var noise = %f;
		CanvasRenderingContext2D.prototype.getImageData = function() {
			var imageData = originalGetImageData.apply(this, arguments);
			var data = imageData.data;
			for (var i = 0; i < data.length; i += 4) {
				data[i] = Math.max(0, Math.min(255, data[i] + Math.floor((Math.random()-0.5)*noise)));
				data[i+1] = Math.max(0, Math.min(255, data[i+1] + Math.floor((Math.random()-0.5)*noise)));
				data[i+2] = Math.max(0, Math.min(255, data[i+2] + Math.floor((Math.random()-0.5)*noise)));
			}
			return imageData;
		};
	})();`, cf.CanvasNoise)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

// InjectWebGLFingerprint injects WebGL fingerprinting
func (cf *CanvasFingerprint) InjectWebGLFingerprint(ctx context.Context) error {
	vendor := escapeJS(cf.WebGLVendor)
	renderer := escapeJS(cf.WebGLRenderer)
	script := fmt.Sprintf(`(function(){
		var getParam = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(param) {
			if (param === 37445) return '%s';
			if (param === 37446) return '%s';
			return getParam.apply(this, arguments);
		};
	})();`, vendor, renderer)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

// InjectAudioFingerprint adds audio context noise
func (cf *CanvasFingerprint) InjectAudioFingerprint(ctx context.Context) error {
	script := fmt.Sprintf(`(function(){
		var AudioCtx = window.AudioContext || window.webkitAudioContext;
		if (!AudioCtx) return;
		var nativeCreate = AudioContext.prototype.createOscillator;
		var noise = %f;
		AudioContext.prototype.createOscillator = function() {
			var osc = nativeCreate.apply(this, arguments);
			var nativeStart = osc.start;
			osc.start = function() {
				osc.frequency.value += (Math.random()-0.5)*noise;
				return nativeStart.apply(this, arguments);
			};
			return osc;
		};
	})();`, cf.AudioNoise)
	return chromedp.Evaluate(script, nil).Do(ctx)
}

func escapeJS(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '\'':
			b = append(b, '\\')
			b = append(b, s[i])
		default:
			b = append(b, s[i])
		}
	}
	return string(b)
}

func randomNoise(min, max float64) float64 {
	b := make([]byte, 8)
	rand.Read(b)
	r := float64(b[0]) / 255.0
	return min + r*(max-min)
}

func randomWebGLVendor() string {
	vendors := []string{
		"Google Inc. (NVIDIA)",
		"Google Inc. (Intel)",
		"Google Inc. (AMD)",
		"Google Inc. (Apple)",
	}
	return vendors[randomInt(len(vendors))]
}

func randomWebGLRenderer() string {
	renderers := []string{
		"ANGLE (NVIDIA GeForce RTX 3060 Ti Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel(R) UHD Graphics 630 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (AMD Radeon RX 6700 XT Direct3D11 vs_5_0 ps_5_0)",
		"Apple M1",
		"Apple M2",
	}
	return renderers[randomInt(len(renderers))]
}

func randomFonts() []string {
	all := []string{
		"Arial", "Verdana", "Helvetica", "Times New Roman",
		"Courier New", "Georgia", "Palatino", "Garamond",
		"Bookman", "Comic Sans MS", "Trebuchet MS", "Impact",
	}
	count := 8 + randomInt(5)
	fonts := make([]string, count)
	for i := 0; i < count; i++ {
		fonts[i] = all[randomInt(len(all))]
	}
	return fonts
}

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
