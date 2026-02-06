// Package serp provides SERP (Search Engine Results Page) click simulation
// with realistic search behavior and click patterns
package serp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// SERPClickSimulator simulates realistic SERP click behavior
type SERPClickSimulator struct {
	SearchEngine     string
	Keyword          string
	TargetDomain     string
	TargetPosition   int
	EnableScrolling  bool
	EnableHovering   bool
	EnableOtherClicks bool
	ClickDelay       time.Duration
}

// SearchEngine constants
const (
	SearchEngineGoogle     = "google"
	SearchEngineBing       = "bing"
	SearchEngineYahoo      = "yahoo"
	SearchEngineDuckDuckGo = "duckduckgo"
	SearchEngineYandex     = "yandex"
)

// NewSERPClickSimulator creates a new SERP click simulator
func NewSERPClickSimulator(searchEngine, keyword, targetDomain string) *SERPClickSimulator {
	return &SERPClickSimulator{
		SearchEngine:     searchEngine,
		Keyword:          keyword,
		TargetDomain:     targetDomain,
		TargetPosition:   0, // Will be determined dynamically
		EnableScrolling:  true,
		EnableHovering:   true,
		EnableOtherClicks: true,
		ClickDelay:       time.Duration(randomInt(2000, 5000)) * time.Millisecond,
	}
}

// GenerateSearchURL generates the search URL for the given search engine
func (s *SERPClickSimulator) GenerateSearchURL() string {
	encodedKeyword := url.QueryEscape(s.Keyword)

	switch s.SearchEngine {
	case SearchEngineGoogle:
		return fmt.Sprintf("https://www.google.com/search?q=%s&hl=en", encodedKeyword)
	case SearchEngineBing:
		return fmt.Sprintf("https://www.bing.com/search?q=%s", encodedKeyword)
	case SearchEngineYahoo:
		return fmt.Sprintf("https://search.yahoo.com/search?p=%s", encodedKeyword)
	case SearchEngineDuckDuckGo:
		return fmt.Sprintf("https://duckduckgo.com/?q=%s", encodedKeyword)
	case SearchEngineYandex:
		return fmt.Sprintf("https://yandex.com/search/?text=%s", encodedKeyword)
	default:
		return fmt.Sprintf("https://www.google.com/search?q=%s&hl=en", encodedKeyword)
	}
}

// GenerateSERPClickScript generates JavaScript for SERP click simulation
func (s *SERPClickSimulator) GenerateSERPClickScript() string {
	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// SERP Click Simulation Configuration
	const serpConfig = {
		searchEngine: '%s',
		targetDomain: '%s',
		enableScrolling: %t,
		enableHovering: %t,
		enableOtherClicks: %t,
		clickDelay: %d
	};
	
	// Search engine specific selectors
	const selectors = {
		google: {
			results: 'div.g, div[data-hveid]',
			links: 'a[href*="http"]',
			title: 'h3',
			snippet: '.VwiC3b, .IsZvec',
			pagination: '#pnnext',
			searchBox: 'input[name="q"], textarea[name="q"]'
		},
		bing: {
			results: 'li.b_algo',
			links: 'a[href*="http"]',
			title: 'h2',
			snippet: '.b_caption p',
			pagination: 'a.sb_pagN',
			searchBox: 'input[name="q"]'
		},
		yahoo: {
			results: 'div.algo',
			links: 'a[href*="http"]',
			title: 'h3',
			snippet: '.compText',
			pagination: 'a.next',
			searchBox: 'input[name="p"]'
		},
		duckduckgo: {
			results: 'article[data-testid="result"]',
			links: 'a[href*="http"]',
			title: 'h2',
			snippet: '[data-result="snippet"]',
			pagination: '.result--more',
			searchBox: 'input[name="q"]'
		}
	};
	
	const currentSelectors = selectors[serpConfig.searchEngine] || selectors.google;
	
	// Find all search results
	function findResults() {
		return document.querySelectorAll(currentSelectors.results);
	}
	
	// Find target result by domain
	function findTargetResult() {
		const results = findResults();
		for (let i = 0; i < results.length; i++) {
			const links = results[i].querySelectorAll(currentSelectors.links);
			for (const link of links) {
				if (link.href && link.href.includes(serpConfig.targetDomain)) {
					return { element: results[i], link: link, position: i + 1 };
				}
			}
		}
		return null;
	}
	
	// Simulate realistic scrolling to element
	async function scrollToElement(element, duration = 1000) {
		if (!serpConfig.enableScrolling) {
			element.scrollIntoView({ behavior: 'instant', block: 'center' });
			return;
		}
		
		const targetY = element.getBoundingClientRect().top + window.scrollY - window.innerHeight / 3;
		const startY = window.scrollY;
		const distance = targetY - startY;
		const startTime = performance.now();
		
		return new Promise((resolve) => {
			function step(currentTime) {
				const elapsed = currentTime - startTime;
				const progress = Math.min(elapsed / duration, 1);
				
				// Easing function for natural scrolling
				const easeProgress = 1 - Math.pow(1 - progress, 3);
				
				window.scrollTo(0, startY + distance * easeProgress);
				
				if (progress < 1) {
					requestAnimationFrame(step);
				} else {
					resolve();
				}
			}
			requestAnimationFrame(step);
		});
	}
	
	// Simulate hovering over elements
	async function hoverElement(element, duration = 500) {
		if (!serpConfig.enableHovering) return;
		
		const rect = element.getBoundingClientRect();
		const x = rect.left + rect.width / 2;
		const y = rect.top + rect.height / 2;
		
		// Dispatch mouse events
		const events = ['mouseenter', 'mouseover', 'mousemove'];
		for (const eventType of events) {
			const event = new MouseEvent(eventType, {
				bubbles: true,
				cancelable: true,
				view: window,
				clientX: x,
				clientY: y
			});
			element.dispatchEvent(event);
		}
		
		await new Promise(resolve => setTimeout(resolve, duration));
		
		// Mouse leave
		element.dispatchEvent(new MouseEvent('mouseleave', {
			bubbles: true,
			cancelable: true,
			view: window
		}));
	}
	
	// Simulate clicking on other results (to appear more natural)
	async function clickOtherResults() {
		if (!serpConfig.enableOtherClicks) return;
		
		const results = findResults();
		const numClicks = Math.floor(Math.random() * 2); // 0-1 other clicks
		
		for (let i = 0; i < numClicks && i < results.length; i++) {
			const randomIndex = Math.floor(Math.random() * Math.min(5, results.length));
			const result = results[randomIndex];
			
			if (result) {
				await scrollToElement(result, 500);
				await hoverElement(result, 300);
				
				// Don't actually click, just hover
				await new Promise(resolve => setTimeout(resolve, 500 + Math.random() * 1000));
			}
		}
	}
	
	// Main SERP click simulation
	window.__simulateSERPClick = async function() {
		console.log('[SERP] Starting SERP click simulation');
		
		// Wait for page to fully load
		await new Promise(resolve => setTimeout(resolve, 1000 + Math.random() * 1000));
		
		// Initial scroll to simulate reading
		if (serpConfig.enableScrolling) {
			await new Promise(resolve => setTimeout(resolve, 500));
			window.scrollTo({ top: 100, behavior: 'smooth' });
			await new Promise(resolve => setTimeout(resolve, 1000));
		}
		
		// Optionally interact with other results first
		if (Math.random() > 0.5) {
			await clickOtherResults();
		}
		
		// Find target result
		const target = findTargetResult();
		
		if (!target) {
			console.log('[SERP] Target domain not found in results');
			return { success: false, reason: 'target_not_found' };
		}
		
		console.log('[SERP] Found target at position ' + target.position);
		
		// Scroll to target
		await scrollToElement(target.element, 800 + Math.random() * 400);
		
		// Hover over target
		await hoverElement(target.element, 500 + Math.random() * 500);
		
		// Wait before clicking (reading time)
		await new Promise(resolve => setTimeout(resolve, serpConfig.clickDelay));
		
		// Click the link
		const rect = target.link.getBoundingClientRect();
		const clickX = rect.left + rect.width / 2 + (Math.random() - 0.5) * 10;
		const clickY = rect.top + rect.height / 2 + (Math.random() - 0.5) * 5;
		
		const clickEvent = new MouseEvent('click', {
			bubbles: true,
			cancelable: true,
			view: window,
			clientX: clickX,
			clientY: clickY
		});
		
		target.link.dispatchEvent(clickEvent);
		
		console.log('[SERP] Clicked target link');
		
		return {
			success: true,
			position: target.position,
			url: target.link.href
		};
	};
	
	// Get SERP results info
	window.__getSERPInfo = function() {
		const results = findResults();
		const info = [];
		
		results.forEach((result, index) => {
			const links = result.querySelectorAll(currentSelectors.links);
			const title = result.querySelector(currentSelectors.title);
			const snippet = result.querySelector(currentSelectors.snippet);
			
			if (links.length > 0) {
				info.push({
					position: index + 1,
					title: title ? title.textContent.trim() : '',
					url: links[0].href,
					snippet: snippet ? snippet.textContent.trim().substring(0, 200) : '',
					isTarget: links[0].href.includes(serpConfig.targetDomain)
				});
			}
		});
		
		return info;
	};
	
	console.log('[SERP] SERP click simulation initialized for ' + serpConfig.searchEngine);
})();
`, s.SearchEngine, s.TargetDomain, s.EnableScrolling, s.EnableHovering, s.EnableOtherClicks, s.ClickDelay.Milliseconds())
}

// GenerateReferrerURL generates a realistic referrer URL
func (s *SERPClickSimulator) GenerateReferrerURL() string {
	encodedKeyword := url.QueryEscape(s.Keyword)

	switch s.SearchEngine {
	case SearchEngineGoogle:
		// Google uses various referrer formats
		formats := []string{
			"https://www.google.com/",
			"https://www.google.com/search?q=%s",
			"https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&ved=2ahUKEw&url=https://%s/&usg=AOvVaw",
		}
		format := formats[randomInt(0, len(formats)-1)]
		if strings.Contains(format, "%s") {
			if strings.Count(format, "%s") == 2 {
				return fmt.Sprintf(format, encodedKeyword, s.TargetDomain)
			}
			return fmt.Sprintf(format, encodedKeyword)
		}
		return format
	case SearchEngineBing:
		return fmt.Sprintf("https://www.bing.com/search?q=%s", encodedKeyword)
	case SearchEngineYahoo:
		return fmt.Sprintf("https://search.yahoo.com/search?p=%s", encodedKeyword)
	case SearchEngineDuckDuckGo:
		return fmt.Sprintf("https://duckduckgo.com/?q=%s", encodedKeyword)
	default:
		return fmt.Sprintf("https://www.google.com/search?q=%s", encodedKeyword)
	}
}

// SERPBehaviorProfile represents a user's SERP behavior profile
type SERPBehaviorProfile struct {
	AvgTimeOnSERP      time.Duration
	AvgScrollDepth     float64
	ClickPosition      int
	BounceRate         float64
	PagesPerSession    int
	ReturnToSERP       bool
	TimeBeforeClick    time.Duration
	HoverBeforeClick   bool
}

// GenerateRandomProfile generates a random but realistic SERP behavior profile
func GenerateRandomProfile() *SERPBehaviorProfile {
	return &SERPBehaviorProfile{
		AvgTimeOnSERP:    time.Duration(randomInt(5, 30)) * time.Second,
		AvgScrollDepth:   float64(randomInt(30, 100)) / 100.0,
		ClickPosition:    randomInt(1, 10),
		BounceRate:       float64(randomInt(20, 60)) / 100.0,
		PagesPerSession:  randomInt(1, 5),
		ReturnToSERP:     randomInt(0, 100) < 30, // 30% chance
		TimeBeforeClick:  time.Duration(randomInt(2, 8)) * time.Second,
		HoverBeforeClick: randomInt(0, 100) < 70, // 70% chance
	}
}

// OrganicClickPattern represents an organic click pattern
type OrganicClickPattern struct {
	Position    int
	CTR         float64 // Click-through rate
	AvgDwell    time.Duration
	BounceProb  float64
}

// GetOrganicClickPatterns returns realistic organic click patterns by position
func GetOrganicClickPatterns() []OrganicClickPattern {
	return []OrganicClickPattern{
		{Position: 1, CTR: 0.284, AvgDwell: 120 * time.Second, BounceProb: 0.25},
		{Position: 2, CTR: 0.155, AvgDwell: 90 * time.Second, BounceProb: 0.30},
		{Position: 3, CTR: 0.109, AvgDwell: 75 * time.Second, BounceProb: 0.35},
		{Position: 4, CTR: 0.078, AvgDwell: 60 * time.Second, BounceProb: 0.40},
		{Position: 5, CTR: 0.058, AvgDwell: 50 * time.Second, BounceProb: 0.45},
		{Position: 6, CTR: 0.044, AvgDwell: 45 * time.Second, BounceProb: 0.48},
		{Position: 7, CTR: 0.035, AvgDwell: 40 * time.Second, BounceProb: 0.50},
		{Position: 8, CTR: 0.029, AvgDwell: 35 * time.Second, BounceProb: 0.52},
		{Position: 9, CTR: 0.024, AvgDwell: 30 * time.Second, BounceProb: 0.55},
		{Position: 10, CTR: 0.020, AvgDwell: 25 * time.Second, BounceProb: 0.58},
	}
}

// SelectPositionByCTR selects a position based on CTR distribution
func SelectPositionByCTR() int {
	patterns := GetOrganicClickPatterns()
	totalCTR := 0.0
	for _, p := range patterns {
		totalCTR += p.CTR
	}

	r := float64(randomInt(0, 1000)) / 1000.0 * totalCTR
	cumulative := 0.0

	for _, p := range patterns {
		cumulative += p.CTR
		if r <= cumulative {
			return p.Position
		}
	}

	return 1
}

// SearchQueryGenerator generates realistic search queries
type SearchQueryGenerator struct {
	BaseKeywords    []string
	LongTailEnabled bool
	Modifiers       []string
	Locations       []string
}

// NewSearchQueryGenerator creates a new search query generator
func NewSearchQueryGenerator(baseKeywords []string) *SearchQueryGenerator {
	return &SearchQueryGenerator{
		BaseKeywords:    baseKeywords,
		LongTailEnabled: true,
		Modifiers: []string{
			"best", "top", "how to", "what is", "where to",
			"cheap", "free", "online", "near me", "reviews",
			"vs", "alternative", "guide", "tutorial", "tips",
		},
		Locations: []string{
			"", "USA", "UK", "Canada", "Australia",
			"New York", "London", "Toronto", "Sydney",
		},
	}
}

// Generate generates a search query
func (g *SearchQueryGenerator) Generate() string {
	if len(g.BaseKeywords) == 0 {
		return ""
	}

	keyword := g.BaseKeywords[randomInt(0, len(g.BaseKeywords)-1)]

	if g.LongTailEnabled && randomInt(0, 100) < 60 {
		// Add modifier
		modifier := g.Modifiers[randomInt(0, len(g.Modifiers)-1)]
		if randomInt(0, 1) == 0 {
			keyword = modifier + " " + keyword
		} else {
			keyword = keyword + " " + modifier
		}
	}

	if randomInt(0, 100) < 20 {
		// Add location
		location := g.Locations[randomInt(0, len(g.Locations)-1)]
		if location != "" {
			keyword = keyword + " " + location
		}
	}

	return keyword
}

// GenerateMultiple generates multiple search queries
func (g *SearchQueryGenerator) GenerateMultiple(count int) []string {
	queries := make([]string, count)
	for i := 0; i < count; i++ {
		queries[i] = g.Generate()
	}
	return queries
}

func randomInt(min, max int) int {
	if max <= min {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}
