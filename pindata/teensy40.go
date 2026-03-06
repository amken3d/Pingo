package pindata

// buildTeensy40Pins constructs the Teensy 4.0 pin list.
//
// Teensy 4.0 (NXP i.MX RT1062) physical layout (top view, USB at top):
//
//	Left side:  GND, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 3.3V
//	Right side: Vin, GND, 3.3V, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13
//	Bottom pads: 24-33 (active), 34-39 (under SD card pads)
//
// Pin data sourced from PJRC Teensy 4.0 pinout card and NXP i.MX RT1062 docs.
//
// Serial buses: Serial1(0,1) Serial2(7,8) Serial3(14,15) Serial4(16,17)
//
//	Serial5(20,21) Serial6(24,25) Serial7(28,29) Serial8(32,3-alt)
//
// SPI buses:    SPI(10-CS,11-MOSI,12-MISO,13-SCK) SPI1(37-SCK,35-MOSI,34-MISO,36-CS)
// I2C buses:    Wire(18-SDA,19-SCL) Wire1(17-SDA,16-SCL) Wire2(25-SDA,24-SCL)
// CAN buses:    CAN1(22-TX,23-RX) CAN2(1-TX,0-RX) CAN3(31-TX,30-RX)
// ADC:          A0-A9 = pins 14-23, A10-A13 = pins 24-27
func buildTeensy40Pins() []Pin {
	ser := func(port int, isTX bool) Function {
		dir := "RX"
		if isTX {
			dir = "TX"
		}
		return Function{Name: "Serial" + itoa(port) + " " + dir, Category: "Serial"}
	}

	spi := func(bus int, role string) Function {
		name := "SPI"
		if bus > 0 {
			name += itoa(bus)
		}
		return Function{Name: name + " " + role, Category: "SPI"}
	}

	i2c := func(bus string, isSCL bool) Function {
		pin := "SDA"
		if isSCL {
			pin = "SCL"
		}
		return Function{Name: bus + " " + pin, Category: "I2C"}
	}

	pwm := func(label string) Function {
		return Function{Name: label, Category: "PWM"}
	}

	adc := func(ch int) Function {
		return Function{Name: "A" + itoa(ch), Category: "ADC"}
	}

	can := func(bus int, isTX bool) Function {
		dir := "RX"
		if isTX {
			dir = "TX"
		}
		return Function{Name: "CAN" + itoa(bus) + " " + dir, Category: "CAN"}
	}

	gpioF := Function{Name: "GPIO", Category: "GPIO"}

	// Arranged as physical pairs: (left, right) from top to bottom.
	// Left side pins: GND, 0-12, 3.3V
	// Right side pins: Vin, GND, 3.3V, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13
	// Bottom pads: 24-39

	pins := []Pin{
		// Physical 1-2: GND / Vin
		ground(),
		power("Vin"),

		// Physical 3-4: Pin 0 / GND
		{Label: "0", GPIO: 0, IsGPIO: true, Functions: []Function{
			gpioF, ser(1, false), can(2, false), pwm("FlexPWM1.1X"),
		}, ADCChannel: -1, PWMSlice: 11, PWMChannel: "X"},
		ground(),

		// Physical 5-6: Pin 1 / 3.3V
		{Label: "1", GPIO: 1, IsGPIO: true, Functions: []Function{
			gpioF, ser(1, true), can(2, true), pwm("FlexPWM1.0X"),
		}, ADCChannel: -1, PWMSlice: 10, PWMChannel: "X"},
		power("3.3V"),

		// Physical 7-8: Pin 2 / Pin 23
		{Label: "2", GPIO: 2, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM4.2A"),
		}, ADCChannel: -1, PWMSlice: 42, PWMChannel: "A"},
		{Label: "23", GPIO: 23, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM4.0B"), can(1, false), adc(9),
		}, ADCChannel: 9, PWMSlice: 40, PWMChannel: "B"},

		// Physical 9-10: Pin 3 / Pin 22
		{Label: "3", GPIO: 3, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM4.2B"), ser(8, false),
		}, ADCChannel: -1, PWMSlice: 42, PWMChannel: "B"},
		{Label: "22", GPIO: 22, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM4.0A"), can(1, true), adc(8),
		}, ADCChannel: 8, PWMSlice: 40, PWMChannel: "A"},

		// Physical 11-12: Pin 4 / Pin 21
		{Label: "4", GPIO: 4, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM2.0A"),
		}, ADCChannel: -1, PWMSlice: 20, PWMChannel: "A"},
		{Label: "21", GPIO: 21, IsGPIO: true, Functions: []Function{
			gpioF, ser(5, false), adc(7),
		}, ADCChannel: 7, PWMSlice: -1},

		// Physical 13-14: Pin 5 / Pin 20
		{Label: "5", GPIO: 5, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM2.1A"),
		}, ADCChannel: -1, PWMSlice: 21, PWMChannel: "A"},
		{Label: "20", GPIO: 20, IsGPIO: true, Functions: []Function{
			gpioF, ser(5, true), adc(6),
		}, ADCChannel: 6, PWMSlice: -1},

		// Physical 15-16: Pin 6 / Pin 19
		{Label: "6", GPIO: 6, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM2.2A"),
		}, ADCChannel: -1, PWMSlice: 22, PWMChannel: "A"},
		{Label: "19", GPIO: 19, IsGPIO: true, Functions: []Function{
			gpioF, i2c("Wire", true), adc(5),
		}, ADCChannel: 5, PWMSlice: -1},

		// Physical 17-18: Pin 7 / Pin 18
		{Label: "7", GPIO: 7, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.3B"), ser(2, false),
		}, ADCChannel: -1, PWMSlice: 13, PWMChannel: "B"},
		{Label: "18", GPIO: 18, IsGPIO: true, Functions: []Function{
			gpioF, i2c("Wire", false), adc(4),
		}, ADCChannel: 4, PWMSlice: -1},

		// Physical 19-20: Pin 8 / Pin 17
		{Label: "8", GPIO: 8, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.3A"), ser(2, true),
		}, ADCChannel: -1, PWMSlice: 13, PWMChannel: "A"},
		{Label: "17", GPIO: 17, IsGPIO: true, Functions: []Function{
			gpioF, ser(4, true), i2c("Wire1", false), spi(1, "MOSI"), adc(3),
		}, ADCChannel: 3, PWMSlice: -1},

		// Physical 21-22: Pin 9 / Pin 16
		{Label: "9", GPIO: 9, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM2.2B"),
		}, ADCChannel: -1, PWMSlice: 22, PWMChannel: "B"},
		{Label: "16", GPIO: 16, IsGPIO: true, Functions: []Function{
			gpioF, ser(4, false), i2c("Wire1", true), spi(1, "SCK"), adc(2),
		}, ADCChannel: 2, PWMSlice: -1},

		// Physical 23-24: Pin 10 / Pin 15
		{Label: "10", GPIO: 10, IsGPIO: true, Functions: []Function{
			gpioF, pwm("QuadTimer3.0"), spi(0, "CS"),
		}, ADCChannel: -1, PWMSlice: -1},
		{Label: "15", GPIO: 15, IsGPIO: true, Functions: []Function{
			gpioF, ser(3, false), adc(1),
		}, ADCChannel: 1, PWMSlice: -1},

		// Physical 25-26: Pin 11 / Pin 14
		{Label: "11", GPIO: 11, IsGPIO: true, Functions: []Function{
			gpioF, pwm("QuadTimer1.2"), spi(0, "MOSI"),
		}, ADCChannel: -1, PWMSlice: -1},
		{Label: "14", GPIO: 14, IsGPIO: true, Functions: []Function{
			gpioF, ser(3, true), adc(0),
		}, ADCChannel: 0, PWMSlice: -1},

		// Physical 27-28: Pin 12 / Pin 13
		{Label: "12", GPIO: 12, IsGPIO: true, Functions: []Function{
			gpioF, pwm("QuadTimer1.1"), spi(0, "MISO"),
		}, ADCChannel: -1, PWMSlice: -1},
		{Label: "13", GPIO: 13, IsGPIO: true, Functions: []Function{
			gpioF, pwm("QuadTimer2.0"), spi(0, "SCK"), /* built-in LED */
		}, ADCChannel: -1, PWMSlice: -1},

		// Physical 29-30: 3.3V / Program
		power("3.3V"),
		special("Program"),

		// ── Bottom pads: pins 24-39 ──────────────────────────────────

		// Physical 31-32: Pin 24 / Pin 33
		{Label: "24", GPIO: 24, IsGPIO: true, Functions: []Function{
			gpioF, i2c("Wire2", true), ser(6, true), adc(10),
		}, ADCChannel: 10, PWMSlice: -1},
		{Label: "33", GPIO: 33, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM2.0B"),
		}, ADCChannel: -1, PWMSlice: 20, PWMChannel: "B"},

		// Physical 33-34: Pin 25 / Pin 34
		{Label: "25", GPIO: 25, IsGPIO: true, Functions: []Function{
			gpioF, i2c("Wire2", false), ser(6, false), adc(11),
		}, ADCChannel: 11, PWMSlice: -1},
		{Label: "34", GPIO: 34, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.1B"), spi(1, "MISO"),
		}, ADCChannel: -1, PWMSlice: 11, PWMChannel: "B"},

		// Physical 35-36: Pin 26 / Pin 35
		{Label: "26", GPIO: 26, IsGPIO: true, Functions: []Function{
			gpioF, adc(12),
		}, ADCChannel: 12, PWMSlice: -1},
		{Label: "35", GPIO: 35, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.1A"), spi(1, "MOSI"),
		}, ADCChannel: -1, PWMSlice: 11, PWMChannel: "A"},

		// Physical 37-38: Pin 27 / Pin 36
		{Label: "27", GPIO: 27, IsGPIO: true, Functions: []Function{
			gpioF, adc(13),
		}, ADCChannel: 13, PWMSlice: -1},
		{Label: "36", GPIO: 36, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.0B"), spi(1, "CS"),
		}, ADCChannel: -1, PWMSlice: 10, PWMChannel: "B"},

		// Physical 39-40: Pin 28 / Pin 37
		{Label: "28", GPIO: 28, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM3.1B"), ser(7, false),
		}, ADCChannel: -1, PWMSlice: 31, PWMChannel: "B"},
		{Label: "37", GPIO: 37, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.0A"), spi(1, "SCK"),
		}, ADCChannel: -1, PWMSlice: 10, PWMChannel: "A"},

		// Physical 41-42: Pin 29 / Pin 38
		{Label: "29", GPIO: 29, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM3.1A"), ser(7, true),
		}, ADCChannel: -1, PWMSlice: 31, PWMChannel: "A"},
		{Label: "38", GPIO: 38, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.2B"),
		}, ADCChannel: -1, PWMSlice: 12, PWMChannel: "B"},

		// Physical 43-44: Pin 30 / Pin 39
		{Label: "30", GPIO: 30, IsGPIO: true, Functions: []Function{
			gpioF, can(3, false),
		}, ADCChannel: -1, PWMSlice: -1},
		{Label: "39", GPIO: 39, IsGPIO: true, Functions: []Function{
			gpioF, pwm("FlexPWM1.2A"),
		}, ADCChannel: -1, PWMSlice: 12, PWMChannel: "A"},

		// Physical 45-46: Pin 31 / VBAT
		{Label: "31", GPIO: 31, IsGPIO: true, Functions: []Function{
			gpioF, can(3, true),
		}, ADCChannel: -1, PWMSlice: -1},
		power("VBAT"),

		// Physical 47-48: Pin 32 / 3.3V
		{Label: "32", GPIO: 32, IsGPIO: true, Functions: []Function{
			gpioF, ser(8, true),
		}, ADCChannel: -1, PWMSlice: -1},
		power("3.3V"),
	}

	// Assign physical pin numbers
	for i := range pins {
		pins[i].PhysicalPin = i + 1
	}

	return pins
}
