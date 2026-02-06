// Package session provides chromedp integration for session management.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"eroshit/pkg/canvas"
)

// ApplyToContext applies session data to a chromedp context
func (s *Session) ApplyToContext(ctx context.Context) error {
	// Apply cookies
	if len(s.Cookies) > 0 {
		params := make([]*network.CookieParam, 0, len(s.Cookies))
		for _, c := range s.Cookies {
			params = append(params, &network.CookieParam{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Secure:   c.Secure,
				HTTPOnly: c.HttpOnly,
			})
		}
		if err := network.SetCookies(params).Do(ctx); err != nil {
			return fmt.Errorf("failed to set cookies: %w", err)
		}
	}

	// Apply localStorage
	if len(s.LocalStorage) > 0 {
		script := SetLocalStorageJS(s.LocalStorage)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("failed to set localStorage: %w", err)
		}
	}

	// Apply sessionStorage
	if len(s.SessionStorage) > 0 {
		script := SetSessionStorageJS(s.SessionStorage)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("failed to set sessionStorage: %w", err)
		}
	}

	// Apply user agent if set (via JavaScript injection)
	if s.UserAgent != "" {
		script := fmt.Sprintf(`Object.defineProperty(navigator, 'userAgent', { value: '%s' });`, s.UserAgent)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("failed to set user agent: %w", err)
		}
	}

	// Apply screen resolution if set
	if s.ScreenResolution != "" {
		if err := applyScreenResolution(ctx, s.ScreenResolution); err != nil {
			return fmt.Errorf("failed to set screen resolution: %w", err)
		}
	}

	// Apply timezone if set
	if s.Timezone != "" {
		script := fmt.Sprintf(`Intl.DateTimeFormat().resolvedOptions().timeZone = '%s';`, s.Timezone)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("failed to set timezone: %w", err)
		}
	}

	// Apply language if set
	if s.Language != "" {
		script := fmt.Sprintf(`navigator.language = '%s';`, s.Language)
		if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
			return fmt.Errorf("failed to set language: %w", err)
		}
	}

	return nil
}

// ExtractFromContext extracts session data from a chromedp context
func (s *Session) ExtractFromContext(ctx context.Context) error {
	// Extract cookies
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	s.Cookies = make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		s.Cookies = append(s.Cookies, &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
		})
	}

	// Extract localStorage
	var localStorageJSON string
	if err := chromedp.Evaluate(GetLocalStorageJS(), &localStorageJSON).Do(ctx); err != nil {
		return fmt.Errorf("failed to get localStorage: %w", err)
	}
	if localStorageJSON != "" {
		if err := json.Unmarshal([]byte(localStorageJSON), &s.LocalStorage); err != nil {
			// Non-fatal: localStorage might be empty or inaccessible
			s.LocalStorage = make(map[string]string)
		}
	}

	// Extract sessionStorage
	var sessionStorageJSON string
	if err := chromedp.Evaluate(GetSessionStorageJS(), &sessionStorageJSON).Do(ctx); err != nil {
		return fmt.Errorf("failed to get sessionStorage: %w", err)
	}
	if sessionStorageJSON != "" {
		if err := json.Unmarshal([]byte(sessionStorageJSON), &s.SessionStorage); err != nil {
			// Non-fatal: sessionStorage might be empty or inaccessible
			s.SessionStorage = make(map[string]string)
		}
	}

	// Extract IndexedDB info (basic)
	var indexedDBJSON string
	if err := chromedp.Evaluate(GetIndexedDBJS(), &indexedDBJSON).Do(ctx); err == nil && indexedDBJSON != "" {
		var indexedDB map[string]any
		if err := json.Unmarshal([]byte(indexedDBJSON), &indexedDB); err == nil {
			s.IndexedDB = indexedDB
		}
	}

	// Update last used time
	s.UpdateLastUsed()

	return nil
}

// ApplyCanvasFingerprint applies canvas fingerprint to context
func ApplyCanvasFingerprint(ctx context.Context, fingerprint string) error {
	// Parse fingerprint components
	// Expected format: "WebGLVendor|WebGLRenderer|CanvasNoise"
	
	script := fmt.Sprintf(`(function() {
		// Store original methods
		var origGetParam = WebGLRenderingContext.prototype.getParameter;
		var origGetImageData = CanvasRenderingContext2D.prototype.getImageData;
		
		// Override WebGL getParameter
		WebGLRenderingContext.prototype.getParameter = function(param) {
			if (param === 37445) return '%s'.split('|')[0]; // UNMASKED_VENDOR_WEBGL
			if (param === 37446) return '%s'.split('|')[1]; // UNMASKED_RENDERER_WEBGL
			return origGetParam.apply(this, arguments);
		};
		
		// Override canvas getImageData with noise
		var noise = parseFloat('%s'.split('|')[2]) || 0.0001;
		CanvasRenderingContext2D.prototype.getImageData = function() {
			var imageData = origGetImageData.apply(this, arguments);
			var data = imageData.data;
			for (var i = 0; i < data.length; i += 4) {
				data[i] = Math.max(0, Math.min(255, data[i] + Math.floor((Math.random()-0.5)*noise)));
				data[i+1] = Math.max(0, Math.min(255, data[i+1] + Math.floor((Math.random()-0.5)*noise)));
				data[i+2] = Math.max(0, Math.min(255, data[i+2] + Math.floor((Math.random()-0.5)*noise)));
			}
			return imageData;
		};
	})();`, fingerprint, fingerprint, fingerprint)

	if err := chromedp.Evaluate(script, nil).Do(ctx); err != nil {
		return fmt.Errorf("failed to apply canvas fingerprint: %w", err)
	}

	return nil
}

// GenerateAndApplyCanvasFingerprint generates a new canvas fingerprint and applies it
func GenerateAndApplyCanvasFingerprint(ctx context.Context) (string, error) {
	cf := canvas.GenerateFingerprint()
	fingerprint := fmt.Sprintf("%s|%s|%f", cf.WebGLVendor, cf.WebGLRenderer, cf.CanvasNoise)
	
	if err := ApplyCanvasFingerprint(ctx, fingerprint); err != nil {
		return "", err
	}
	
	return fingerprint, nil
}

// Helper function to apply screen resolution
func applyScreenResolution(ctx context.Context, resolution string) error {
	var width, height int
	_, err := fmt.Sscanf(resolution, "%dx%d", &width, &height)
	if err != nil {
		return err
	}

	script := fmt.Sprintf(`
		Object.defineProperty(window.screen, 'width', { value: %d });
		Object.defineProperty(window.screen, 'height', { value: %d });
		Object.defineProperty(window.screen, 'availWidth', { value: %d });
		Object.defineProperty(window.screen, 'availHeight', { value: %d });
	`, width, height, width, height-40) // Subtract some pixels for taskbar

	return chromedp.Evaluate(script, nil).Do(ctx)
}
