// Package mobile provides mobile device emulation including touch events,
// accelerometer/gyroscope spoofing, device orientation, and mobile keyboard simulation
package mobile

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

// TouchEventSimulator simulates realistic touch events
type TouchEventSimulator struct {
	EnableMultiTouch    bool
	EnableGestures      bool
	EnableHapticFeedback bool
	TouchPressure       float64
	TouchRadius         float64
}

// NewTouchEventSimulator creates a new touch event simulator
func NewTouchEventSimulator() *TouchEventSimulator {
	return &TouchEventSimulator{
		EnableMultiTouch:    true,
		EnableGestures:      true,
		EnableHapticFeedback: true,
		TouchPressure:       0.5,
		TouchRadius:         25.0,
	}
}

// GenerateTouchScript generates JavaScript for touch event simulation
func (t *TouchEventSimulator) GenerateTouchScript() string {
	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// Touch Event Simulation
	const touchConfig = {
		enableMultiTouch: %t,
		enableGestures: %t,
		enableHapticFeedback: %t,
		defaultPressure: %f,
		defaultRadius: %f
	};
	
	// Create realistic Touch object
	class SimulatedTouch {
		constructor(options) {
			this.identifier = options.identifier || Math.floor(Math.random() * 1000000);
			this.target = options.target || document.body;
			this.clientX = options.clientX || 0;
			this.clientY = options.clientY || 0;
			this.screenX = options.screenX || this.clientX;
			this.screenY = options.screenY || this.clientY;
			this.pageX = options.pageX || this.clientX + window.scrollX;
			this.pageY = options.pageY || this.clientY + window.scrollY;
			this.radiusX = options.radiusX || touchConfig.defaultRadius;
			this.radiusY = options.radiusY || touchConfig.defaultRadius;
			this.rotationAngle = options.rotationAngle || 0;
			this.force = options.force || touchConfig.defaultPressure;
		}
	}
	
	// Create TouchList
	class SimulatedTouchList {
		constructor(touches) {
			this._touches = touches || [];
			this.length = this._touches.length;
			this._touches.forEach((touch, i) => {
				this[i] = touch;
			});
		}
		
		item(index) {
			return this._touches[index] || null;
		}
		
		identifiedTouch(identifier) {
			return this._touches.find(t => t.identifier === identifier) || null;
		}
	}
	
	// Touch event dispatcher
	window.__simulateTouch = function(type, x, y, options = {}) {
		const target = document.elementFromPoint(x, y) || document.body;
		
		const touch = new SimulatedTouch({
			identifier: options.identifier || Date.now(),
			target: target,
			clientX: x,
			clientY: y,
			screenX: x,
			screenY: y,
			pageX: x + window.scrollX,
			pageY: y + window.scrollY,
			radiusX: options.radiusX || touchConfig.defaultRadius,
			radiusY: options.radiusY || touchConfig.defaultRadius,
			rotationAngle: options.rotationAngle || 0,
			force: options.force || touchConfig.defaultPressure
		});
		
		const touchList = new SimulatedTouchList([touch]);
		
		const touchEvent = new TouchEvent(type, {
			bubbles: true,
			cancelable: true,
			view: window,
			touches: type === 'touchend' ? new SimulatedTouchList([]) : touchList,
			targetTouches: type === 'touchend' ? new SimulatedTouchList([]) : touchList,
			changedTouches: touchList
		});
		
		target.dispatchEvent(touchEvent);
		return touchEvent;
	};
	
	// Gesture simulation
	window.__simulateGesture = function(gestureType, startX, startY, endX, endY, duration = 300) {
		return new Promise((resolve) => {
			const identifier = Date.now();
			const steps = Math.max(10, Math.floor(duration / 16));
			const deltaX = (endX - startX) / steps;
			const deltaY = (endY - startY) / steps;
			
			let currentStep = 0;
			
			// Start touch
			window.__simulateTouch('touchstart', startX, startY, { identifier });
			
			const moveInterval = setInterval(() => {
				currentStep++;
				const x = startX + deltaX * currentStep;
				const y = startY + deltaY * currentStep;
				
				window.__simulateTouch('touchmove', x, y, { identifier });
				
				if (currentStep >= steps) {
					clearInterval(moveInterval);
					window.__simulateTouch('touchend', endX, endY, { identifier });
					resolve();
				}
			}, duration / steps);
		});
	};
	
	// Swipe gestures
	window.__swipe = function(direction, startX, startY, distance = 200, duration = 300) {
		let endX = startX, endY = startY;
		
		switch(direction) {
			case 'up': endY = startY - distance; break;
			case 'down': endY = startY + distance; break;
			case 'left': endX = startX - distance; break;
			case 'right': endX = startX + distance; break;
		}
		
		return window.__simulateGesture('swipe', startX, startY, endX, endY, duration);
	};
	
	// Pinch gesture
	window.__pinch = function(centerX, centerY, startDistance, endDistance, duration = 300) {
		return new Promise((resolve) => {
			const id1 = Date.now();
			const id2 = Date.now() + 1;
			const steps = Math.max(10, Math.floor(duration / 16));
			const deltaDistance = (endDistance - startDistance) / steps;
			
			let currentStep = 0;
			let currentDistance = startDistance;
			
			// Start touches
			window.__simulateTouch('touchstart', centerX - currentDistance/2, centerY, { identifier: id1 });
			window.__simulateTouch('touchstart', centerX + currentDistance/2, centerY, { identifier: id2 });
			
			const moveInterval = setInterval(() => {
				currentStep++;
				currentDistance = startDistance + deltaDistance * currentStep;
				
				window.__simulateTouch('touchmove', centerX - currentDistance/2, centerY, { identifier: id1 });
				window.__simulateTouch('touchmove', centerX + currentDistance/2, centerY, { identifier: id2 });
				
				if (currentStep >= steps) {
					clearInterval(moveInterval);
					window.__simulateTouch('touchend', centerX - currentDistance/2, centerY, { identifier: id1 });
					window.__simulateTouch('touchend', centerX + currentDistance/2, centerY, { identifier: id2 });
					resolve();
				}
			}, duration / steps);
		});
	};
	
	// Tap gesture
	window.__tap = function(x, y, duration = 100) {
		return new Promise((resolve) => {
			const identifier = Date.now();
			window.__simulateTouch('touchstart', x, y, { identifier });
			
			setTimeout(() => {
				window.__simulateTouch('touchend', x, y, { identifier });
				resolve();
			}, duration);
		});
	};
	
	// Double tap
	window.__doubleTap = function(x, y) {
		return window.__tap(x, y, 50).then(() => {
			return new Promise(resolve => setTimeout(resolve, 100));
		}).then(() => {
			return window.__tap(x, y, 50);
		});
	};
	
	// Long press
	window.__longPress = function(x, y, duration = 500) {
		return window.__tap(x, y, duration);
	};
	
	console.log('[TouchSimulator] Touch event simulation initialized');
})();
`, t.EnableMultiTouch, t.EnableGestures, t.EnableHapticFeedback, t.TouchPressure, t.TouchRadius)
}

// AccelerometerGyroscopeSpoofer spoofs accelerometer and gyroscope data
type AccelerometerGyroscopeSpoofer struct {
	EnableAccelerometer bool
	EnableGyroscope     bool
	EnableMagnetometer  bool
	BaseAccelX          float64
	BaseAccelY          float64
	BaseAccelZ          float64
	NoiseLevel          float64
}

// NewAccelerometerGyroscopeSpoofer creates a new sensor spoofer
func NewAccelerometerGyroscopeSpoofer() *AccelerometerGyroscopeSpoofer {
	return &AccelerometerGyroscopeSpoofer{
		EnableAccelerometer: true,
		EnableGyroscope:     true,
		EnableMagnetometer:  true,
		BaseAccelX:          0.0,
		BaseAccelY:          0.0,
		BaseAccelZ:          9.81, // Earth gravity
		NoiseLevel:          0.1,
	}
}

// GenerateSensorScript generates JavaScript for sensor spoofing
func (s *AccelerometerGyroscopeSpoofer) GenerateSensorScript() string {
	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// Sensor Spoofing Configuration
	const sensorConfig = {
		enableAccelerometer: %t,
		enableGyroscope: %t,
		enableMagnetometer: %t,
		baseAccel: { x: %f, y: %f, z: %f },
		noiseLevel: %f
	};
	
	// Generate realistic noise
	function addNoise(value, level) {
		return value + (Math.random() - 0.5) * 2 * level;
	}
	
	// Accelerometer Spoofing
	if (sensorConfig.enableAccelerometer && window.Accelerometer) {
		const OriginalAccelerometer = window.Accelerometer;
		
		window.Accelerometer = function(options) {
			const sensor = new OriginalAccelerometer(options);
			
			Object.defineProperties(sensor, {
				x: {
					get: () => addNoise(sensorConfig.baseAccel.x, sensorConfig.noiseLevel),
					configurable: true
				},
				y: {
					get: () => addNoise(sensorConfig.baseAccel.y, sensorConfig.noiseLevel),
					configurable: true
				},
				z: {
					get: () => addNoise(sensorConfig.baseAccel.z, sensorConfig.noiseLevel),
					configurable: true
				}
			});
			
			return sensor;
		};
		window.Accelerometer.prototype = OriginalAccelerometer.prototype;
	}
	
	// Gyroscope Spoofing
	if (sensorConfig.enableGyroscope && window.Gyroscope) {
		const OriginalGyroscope = window.Gyroscope;
		
		window.Gyroscope = function(options) {
			const sensor = new OriginalGyroscope(options);
			
			Object.defineProperties(sensor, {
				x: {
					get: () => addNoise(0, sensorConfig.noiseLevel * 0.1),
					configurable: true
				},
				y: {
					get: () => addNoise(0, sensorConfig.noiseLevel * 0.1),
					configurable: true
				},
				z: {
					get: () => addNoise(0, sensorConfig.noiseLevel * 0.1),
					configurable: true
				}
			});
			
			return sensor;
		};
		window.Gyroscope.prototype = OriginalGyroscope.prototype;
	}
	
	// LinearAccelerationSensor Spoofing
	if (window.LinearAccelerationSensor) {
		const OriginalLinearAccelerationSensor = window.LinearAccelerationSensor;
		
		window.LinearAccelerationSensor = function(options) {
			const sensor = new OriginalLinearAccelerationSensor(options);
			
			Object.defineProperties(sensor, {
				x: { get: () => addNoise(0, sensorConfig.noiseLevel * 0.5), configurable: true },
				y: { get: () => addNoise(0, sensorConfig.noiseLevel * 0.5), configurable: true },
				z: { get: () => addNoise(0, sensorConfig.noiseLevel * 0.5), configurable: true }
			});
			
			return sensor;
		};
		window.LinearAccelerationSensor.prototype = OriginalLinearAccelerationSensor.prototype;
	}
	
	// GravitySensor Spoofing
	if (window.GravitySensor) {
		const OriginalGravitySensor = window.GravitySensor;
		
		window.GravitySensor = function(options) {
			const sensor = new OriginalGravitySensor(options);
			
			Object.defineProperties(sensor, {
				x: { get: () => addNoise(sensorConfig.baseAccel.x, sensorConfig.noiseLevel * 0.01), configurable: true },
				y: { get: () => addNoise(sensorConfig.baseAccel.y, sensorConfig.noiseLevel * 0.01), configurable: true },
				z: { get: () => addNoise(sensorConfig.baseAccel.z, sensorConfig.noiseLevel * 0.01), configurable: true }
			});
			
			return sensor;
		};
		window.GravitySensor.prototype = OriginalGravitySensor.prototype;
	}
	
	// Magnetometer Spoofing
	if (sensorConfig.enableMagnetometer && window.Magnetometer) {
		const OriginalMagnetometer = window.Magnetometer;
		
		window.Magnetometer = function(options) {
			const sensor = new OriginalMagnetometer(options);
			
			// Earth's magnetic field is roughly 25-65 microteslas
			Object.defineProperties(sensor, {
				x: { get: () => addNoise(25, 5), configurable: true },
				y: { get: () => addNoise(0, 5), configurable: true },
				z: { get: () => addNoise(-45, 5), configurable: true }
			});
			
			return sensor;
		};
		window.Magnetometer.prototype = OriginalMagnetometer.prototype;
	}
	
	// DeviceMotionEvent Spoofing
	let motionInterval = null;
	const originalAddEventListener = window.addEventListener;
	
	window.addEventListener = function(type, listener, options) {
		if (type === 'devicemotion' && sensorConfig.enableAccelerometer) {
			if (!motionInterval) {
				motionInterval = setInterval(() => {
					const event = new DeviceMotionEvent('devicemotion', {
						acceleration: {
							x: addNoise(0, sensorConfig.noiseLevel * 0.5),
							y: addNoise(0, sensorConfig.noiseLevel * 0.5),
							z: addNoise(0, sensorConfig.noiseLevel * 0.5)
						},
						accelerationIncludingGravity: {
							x: addNoise(sensorConfig.baseAccel.x, sensorConfig.noiseLevel),
							y: addNoise(sensorConfig.baseAccel.y, sensorConfig.noiseLevel),
							z: addNoise(sensorConfig.baseAccel.z, sensorConfig.noiseLevel)
						},
						rotationRate: {
							alpha: addNoise(0, sensorConfig.noiseLevel * 0.1),
							beta: addNoise(0, sensorConfig.noiseLevel * 0.1),
							gamma: addNoise(0, sensorConfig.noiseLevel * 0.1)
						},
						interval: 16
					});
					window.dispatchEvent(event);
				}, 16);
			}
		}
		return originalAddEventListener.apply(this, arguments);
	};
	
	console.log('[SensorSpoofer] Accelerometer/Gyroscope spoofing initialized');
})();
`, s.EnableAccelerometer, s.EnableGyroscope, s.EnableMagnetometer,
		s.BaseAccelX, s.BaseAccelY, s.BaseAccelZ, s.NoiseLevel)
}

// DeviceOrientationSimulator simulates device orientation
type DeviceOrientationSimulator struct {
	Alpha float64 // Compass direction (0-360)
	Beta  float64 // Front-back tilt (-180 to 180)
	Gamma float64 // Left-right tilt (-90 to 90)
}

// NewDeviceOrientationSimulator creates a new orientation simulator
func NewDeviceOrientationSimulator() *DeviceOrientationSimulator {
	return &DeviceOrientationSimulator{
		Alpha: 0,
		Beta:  0,
		Gamma: 0,
	}
}

// GenerateOrientationScript generates JavaScript for device orientation simulation
func (d *DeviceOrientationSimulator) GenerateOrientationScript() string {
	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// Device Orientation Simulation
	const orientationState = {
		alpha: %f,  // Compass direction
		beta: %f,   // Front-back tilt
		gamma: %f,  // Left-right tilt
		absolute: true
	};
	
	// Add noise to orientation values
	function addOrientationNoise(value, maxNoise) {
		return value + (Math.random() - 0.5) * 2 * maxNoise;
	}
	
	// Simulate orientation changes
	let orientationInterval = null;
	const originalAddEventListener = window.addEventListener;
	
	window.addEventListener = function(type, listener, options) {
		if (type === 'deviceorientation') {
			if (!orientationInterval) {
				orientationInterval = setInterval(() => {
					const event = new DeviceOrientationEvent('deviceorientation', {
						alpha: addOrientationNoise(orientationState.alpha, 0.5),
						beta: addOrientationNoise(orientationState.beta, 0.3),
						gamma: addOrientationNoise(orientationState.gamma, 0.3),
						absolute: orientationState.absolute
					});
					window.dispatchEvent(event);
				}, 50);
			}
		}
		if (type === 'deviceorientationabsolute') {
			if (!orientationInterval) {
				orientationInterval = setInterval(() => {
					const event = new DeviceOrientationEvent('deviceorientationabsolute', {
						alpha: addOrientationNoise(orientationState.alpha, 0.5),
						beta: addOrientationNoise(orientationState.beta, 0.3),
						gamma: addOrientationNoise(orientationState.gamma, 0.3),
						absolute: true
					});
					window.dispatchEvent(event);
				}, 50);
			}
		}
		return originalAddEventListener.apply(this, arguments);
	};
	
	// API to update orientation
	window.__setDeviceOrientation = function(alpha, beta, gamma) {
		orientationState.alpha = alpha;
		orientationState.beta = beta;
		orientationState.gamma = gamma;
	};
	
	// Simulate tilting the device
	window.__tiltDevice = function(direction, angle, duration = 500) {
		return new Promise((resolve) => {
			const startBeta = orientationState.beta;
			const startGamma = orientationState.gamma;
			const steps = Math.floor(duration / 16);
			let currentStep = 0;
			
			const interval = setInterval(() => {
				currentStep++;
				const progress = currentStep / steps;
				
				switch(direction) {
					case 'forward':
						orientationState.beta = startBeta + angle * progress;
						break;
					case 'backward':
						orientationState.beta = startBeta - angle * progress;
						break;
					case 'left':
						orientationState.gamma = startGamma - angle * progress;
						break;
					case 'right':
						orientationState.gamma = startGamma + angle * progress;
						break;
				}
				
				if (currentStep >= steps) {
					clearInterval(interval);
					resolve();
				}
			}, 16);
		});
	};
	
	// Simulate rotating the device (compass)
	window.__rotateDevice = function(targetAlpha, duration = 500) {
		return new Promise((resolve) => {
			const startAlpha = orientationState.alpha;
			const deltaAlpha = targetAlpha - startAlpha;
			const steps = Math.floor(duration / 16);
			let currentStep = 0;
			
			const interval = setInterval(() => {
				currentStep++;
				const progress = currentStep / steps;
				orientationState.alpha = startAlpha + deltaAlpha * progress;
				
				if (currentStep >= steps) {
					clearInterval(interval);
					resolve();
				}
			}, 16);
		});
	};
	
	console.log('[OrientationSimulator] Device orientation simulation initialized');
})();
`, d.Alpha, d.Beta, d.Gamma)
}

// MobileKeyboardSimulator simulates mobile keyboard behavior
type MobileKeyboardSimulator struct {
	KeyboardType    string // default, numeric, email, url, search
	AutoCorrect     bool
	AutoCapitalize  bool
	SpellCheck      bool
	TypingSpeed     int // characters per minute
}

// NewMobileKeyboardSimulator creates a new mobile keyboard simulator
func NewMobileKeyboardSimulator() *MobileKeyboardSimulator {
	return &MobileKeyboardSimulator{
		KeyboardType:   "default",
		AutoCorrect:    true,
		AutoCapitalize: true,
		SpellCheck:     true,
		TypingSpeed:    200,
	}
}

// GenerateKeyboardScript generates JavaScript for mobile keyboard simulation
func (k *MobileKeyboardSimulator) GenerateKeyboardScript() string {
	return fmt.Sprintf(`
(function() {
	'use strict';
	
	// Mobile Keyboard Simulation
	const keyboardConfig = {
		type: '%s',
		autoCorrect: %t,
		autoCapitalize: %t,
		spellCheck: %t,
		typingSpeed: %d  // chars per minute
	};
	
	// Calculate delay between keystrokes
	const msPerChar = 60000 / keyboardConfig.typingSpeed;
	
	// Virtual keyboard state
	let keyboardVisible = false;
	let activeInput = null;
	
	// Simulate keyboard appearance
	window.__showKeyboard = function(inputElement) {
		if (keyboardVisible) return;
		
		activeInput = inputElement || document.activeElement;
		keyboardVisible = true;
		
		// Dispatch visualViewport resize (keyboard appearing)
		if (window.visualViewport) {
			const originalHeight = window.visualViewport.height;
			const keyboardHeight = 300; // Approximate keyboard height
			
			Object.defineProperty(window.visualViewport, 'height', {
				get: () => originalHeight - keyboardHeight,
				configurable: true
			});
			
			window.visualViewport.dispatchEvent(new Event('resize'));
		}
		
		// Scroll input into view
		if (activeInput && activeInput.scrollIntoView) {
			activeInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	};
	
	// Simulate keyboard hiding
	window.__hideKeyboard = function() {
		if (!keyboardVisible) return;
		
		keyboardVisible = false;
		activeInput = null;
		
		// Restore visualViewport
		if (window.visualViewport) {
			window.visualViewport.dispatchEvent(new Event('resize'));
		}
	};
	
	// Simulate typing with realistic delays
	window.__mobileType = function(text, inputElement) {
		return new Promise((resolve) => {
			const input = inputElement || document.activeElement;
			if (!input) {
				resolve();
				return;
			}
			
			window.__showKeyboard(input);
			
			let index = 0;
			const chars = text.split('');
			
			function typeNextChar() {
				if (index >= chars.length) {
					resolve();
					return;
				}
				
				const char = chars[index];
				let delay = msPerChar;
				
				// Add variation to typing speed
				delay += (Math.random() - 0.5) * msPerChar * 0.5;
				
				// Longer delay for special characters
				if (/[^a-zA-Z0-9]/.test(char)) {
					delay *= 1.5;
				}
				
				// Simulate keydown
				const keydownEvent = new KeyboardEvent('keydown', {
					key: char,
					code: 'Key' + char.toUpperCase(),
					bubbles: true,
					cancelable: true
				});
				input.dispatchEvent(keydownEvent);
				
				// Update input value
				if (input.tagName === 'INPUT' || input.tagName === 'TEXTAREA') {
					const start = input.selectionStart || 0;
					const end = input.selectionEnd || 0;
					const value = input.value;
					
					// Apply auto-capitalize
					let charToInsert = char;
					if (keyboardConfig.autoCapitalize && start === 0) {
						charToInsert = char.toUpperCase();
					}
					
					input.value = value.substring(0, start) + charToInsert + value.substring(end);
					input.selectionStart = input.selectionEnd = start + 1;
					
					// Dispatch input event
					input.dispatchEvent(new Event('input', { bubbles: true }));
				}
				
				// Simulate keyup
				const keyupEvent = new KeyboardEvent('keyup', {
					key: char,
					code: 'Key' + char.toUpperCase(),
					bubbles: true,
					cancelable: true
				});
				input.dispatchEvent(keyupEvent);
				
				index++;
				setTimeout(typeNextChar, delay);
			}
			
			typeNextChar();
		});
	};
	
	// Simulate backspace
	window.__mobileBackspace = function(count = 1, inputElement) {
		return new Promise((resolve) => {
			const input = inputElement || document.activeElement;
			if (!input) {
				resolve();
				return;
			}
			
			let remaining = count;
			
			function deleteNext() {
				if (remaining <= 0) {
					resolve();
					return;
				}
				
				const keydownEvent = new KeyboardEvent('keydown', {
					key: 'Backspace',
					code: 'Backspace',
					bubbles: true,
					cancelable: true
				});
				input.dispatchEvent(keydownEvent);
				
				if (input.tagName === 'INPUT' || input.tagName === 'TEXTAREA') {
					const start = input.selectionStart || 0;
					const end = input.selectionEnd || 0;
					const value = input.value;
					
					if (start === end && start > 0) {
						input.value = value.substring(0, start - 1) + value.substring(end);
						input.selectionStart = input.selectionEnd = start - 1;
					} else if (start !== end) {
						input.value = value.substring(0, start) + value.substring(end);
						input.selectionStart = input.selectionEnd = start;
					}
					
					input.dispatchEvent(new Event('input', { bubbles: true }));
				}
				
				const keyupEvent = new KeyboardEvent('keyup', {
					key: 'Backspace',
					code: 'Backspace',
					bubbles: true,
					cancelable: true
				});
				input.dispatchEvent(keyupEvent);
				
				remaining--;
				setTimeout(deleteNext, msPerChar * 0.5);
			}
			
			deleteNext();
		});
	};
	
	// Simulate pressing Enter/Return
	window.__mobileEnter = function(inputElement) {
		const input = inputElement || document.activeElement;
		if (!input) return;
		
		const keydownEvent = new KeyboardEvent('keydown', {
			key: 'Enter',
			code: 'Enter',
			bubbles: true,
			cancelable: true
		});
		input.dispatchEvent(keydownEvent);
		
		const keyupEvent = new KeyboardEvent('keyup', {
			key: 'Enter',
			code: 'Enter',
			bubbles: true,
			cancelable: true
		});
		input.dispatchEvent(keyupEvent);
		
		// If it's a form, submit it
		if (input.form) {
			input.form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
		}
	};
	
	// Auto-focus handling for mobile
	document.addEventListener('click', function(e) {
		const target = e.target;
		if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') {
			window.__showKeyboard(target);
		} else if (keyboardVisible) {
			window.__hideKeyboard();
		}
	});
	
	console.log('[MobileKeyboard] Mobile keyboard simulation initialized');
})();
`, k.KeyboardType, k.AutoCorrect, k.AutoCapitalize, k.SpellCheck, k.TypingSpeed)
}

// GenerateRandomDeviceMotion generates random but realistic device motion values
func GenerateRandomDeviceMotion() (accelX, accelY, accelZ float64) {
	// Generate small random accelerations (device at rest with slight movements)
	randX, _ := rand.Int(rand.Reader, big.NewInt(100))
	randY, _ := rand.Int(rand.Reader, big.NewInt(100))
	randZ, _ := rand.Int(rand.Reader, big.NewInt(100))

	accelX = (float64(randX.Int64()) - 50) / 500.0 // -0.1 to 0.1
	accelY = (float64(randY.Int64()) - 50) / 500.0 // -0.1 to 0.1
	accelZ = 9.81 + (float64(randZ.Int64())-50)/500.0 // ~9.81 with noise

	return
}

// GenerateRandomOrientation generates random but realistic device orientation
func GenerateRandomOrientation() (alpha, beta, gamma float64) {
	randAlpha, _ := rand.Int(rand.Reader, big.NewInt(360))
	randBeta, _ := rand.Int(rand.Reader, big.NewInt(20))
	randGamma, _ := rand.Int(rand.Reader, big.NewInt(20))

	alpha = float64(randAlpha.Int64())           // 0-360 compass
	beta = float64(randBeta.Int64()) - 10        // -10 to 10 (slight tilt)
	gamma = float64(randGamma.Int64()) - 10      // -10 to 10 (slight tilt)

	return
}

// NormalizeAngle normalizes an angle to 0-360 range
func NormalizeAngle(angle float64) float64 {
	for angle < 0 {
		angle += 360
	}
	for angle >= 360 {
		angle -= 360
	}
	return angle
}

// CalculateDistance calculates distance between two touch points
func CalculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
