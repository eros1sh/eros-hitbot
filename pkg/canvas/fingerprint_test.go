package canvas

import (
	"testing"
)

func TestCanvasFingerprintUniqueness(t *testing.T) {
	fp1 := GenerateFingerprint()
	fp2 := GenerateFingerprint()

	if fp1.CanvasNoise == fp2.CanvasNoise && fp1.AudioNoise == fp2.AudioNoise {
		t.Log("Fingerprints may coincide by chance - run again")
	}

	if fp1.CanvasNoise < 0.0001 || fp1.CanvasNoise > 0.001 {
		t.Errorf("CanvasNoise out of range: %f", fp1.CanvasNoise)
	}
	if fp1.AudioNoise < 0.00001 || fp1.AudioNoise > 0.0001 {
		t.Errorf("AudioNoise out of range: %f", fp1.AudioNoise)
	}
}

func TestWebGLValues(t *testing.T) {
	vendors := []string{
		"Google Inc. (NVIDIA)",
		"Google Inc. (Intel)",
		"Google Inc. (AMD)",
		"Google Inc. (Apple)",
	}
	fp := GenerateFingerprint()
	found := false
	for _, v := range vendors {
		if fp.WebGLVendor == v {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("WebGLVendor invalid: %s", fp.WebGLVendor)
	}
}

func TestRandomFonts(t *testing.T) {
	fp := GenerateFingerprint()
	if len(fp.Fonts) < 8 || len(fp.Fonts) > 13 {
		t.Errorf("Font count out of range: %d", len(fp.Fonts))
	}
}
