package pindata

// ── Full QFN package pinouts ─────────────────────────────────────────
// These match the physical pin ordering from the RP2040 and RP2350
// datasheets. All pins are included: GPIO, power, ground, and special.
// QFN pin numbering: top (L→R), right (T→B), bottom (R→L), left (B→T).

// buildRP2040QFN56 returns all 56 pins of the RP2040 QFN-56 package.
func buildRP2040QFN56() []Pin {
	const pwm = 8 // RP2040 has 8 PWM slices
	pins := []Pin{
		// ── Top side (pins 1–14, left → right) ──
		power("IOVDD"),  // 1
		rpGPIO(0, pwm),  // 2
		rpGPIO(1, pwm),  // 3
		rpGPIO(2, pwm),  // 4
		rpGPIO(3, pwm),  // 5
		rpGPIO(4, pwm),  // 6
		rpGPIO(5, pwm),  // 7
		rpGPIO(6, pwm),  // 8
		rpGPIO(7, pwm),  // 9
		power("IOVDD"),  // 10
		rpGPIO(8, pwm),  // 11
		rpGPIO(9, pwm),  // 12
		rpGPIO(10, pwm), // 13
		rpGPIO(11, pwm), // 14

		// ── Right side (pins 15–28, top → bottom) ──
		rpGPIO(12, pwm),   // 15
		rpGPIO(13, pwm),   // 16
		rpGPIO(14, pwm),   // 17
		rpGPIO(15, pwm),   // 18
		special("TESTEN"), // 19
		special("XIN"),    // 20
		special("XOUT"),   // 21
		power("IOVDD"),    // 22
		power("DVDD"),     // 23
		special("SWCLK"),  // 24
		special("SWD"),    // 25
		special("RUN"),    // 26
		rpGPIO(16, pwm),   // 27
		rpGPIO(17, pwm),   // 28

		// ── Bottom side (pins 29–42, right → left) ──
		rpGPIO(18, pwm), // 29
		rpGPIO(19, pwm), // 30
		rpGPIO(20, pwm), // 31
		rpGPIO(21, pwm), // 32
		rpGPIO(22, pwm), // 33
		rpGPIO(23, pwm), // 34
		rpGPIO(24, pwm), // 35
		rpGPIO(25, pwm), // 36
		power("IOVDD"),  // 37
		rpGPIO(26, pwm), // 38
		rpGPIO(27, pwm), // 39
		rpGPIO(28, pwm), // 40
		rpGPIO(29, pwm), // 41
		power("IOVDD"),  // 42

		// ── Left side (pins 43–56, bottom → top) ──
		power("ADC_AVDD"),   // 43
		power("VREG_VIN"),   // 44
		power("VREG_VOUT"),  // 45
		special("USB_DM"),   // 46
		special("USB_DP"),   // 47
		power("USB_VDD"),    // 48
		power("IOVDD"),      // 49
		power("DVDD"),       // 50
		special("QSPI_SD3"), // 51
		special("QSPI_SCK"), // 52
		special("QSPI_SD0"), // 53
		special("QSPI_SD2"), // 54
		special("QSPI_SD1"), // 55
		special("QSPI_SS"),  // 56
	}
	for i := range pins {
		pins[i].PhysicalPin = i + 1
	}
	return pins
}

// buildRP2350AQFN60 returns all 60 pins of the RP2350A QFN-60 package.
func buildRP2350AQFN60() []Pin {
	const pwm = 12 // RP2350 has 12 PWM slices
	pins := []Pin{
		// ── Top side (pins 1–15, left → right) ──
		power("IOVDD"),  // 1
		rpGPIO(0, pwm),  // 2
		rpGPIO(1, pwm),  // 3
		rpGPIO(2, pwm),  // 4
		rpGPIO(3, pwm),  // 5
		rpGPIO(4, pwm),  // 6
		rpGPIO(5, pwm),  // 7
		rpGPIO(6, pwm),  // 8
		rpGPIO(7, pwm),  // 9
		power("IOVDD"),  // 10
		power("DVDD"),   // 11
		rpGPIO(8, pwm),  // 12
		rpGPIO(9, pwm),  // 13
		rpGPIO(10, pwm), // 14
		rpGPIO(11, pwm), // 15

		// ── Right side (pins 16–30, top → bottom) ──
		rpGPIO(12, pwm),   // 16
		rpGPIO(13, pwm),   // 17
		rpGPIO(14, pwm),   // 18
		rpGPIO(15, pwm),   // 19
		special("TESTEN"), // 20
		special("XIN"),    // 21
		special("XOUT"),   // 22
		power("IOVDD"),    // 23
		power("DVDD"),     // 24
		special("SWCLK"),  // 25
		special("SWDIO"),  // 26
		special("RUN"),    // 27
		rpGPIO(16, pwm),   // 28
		rpGPIO(17, pwm),   // 29
		rpGPIO(18, pwm),   // 30

		// ── Bottom side (pins 31–45, right → left) ──
		rpGPIO(19, pwm),   // 31
		rpGPIO(20, pwm),   // 32
		rpGPIO(21, pwm),   // 33
		rpGPIO(22, pwm),   // 34
		rpGPIO(23, pwm),   // 35
		rpGPIO(24, pwm),   // 36
		rpGPIO(25, pwm),   // 37
		power("IOVDD"),    // 38
		rpGPIO(26, pwm),   // 39
		rpGPIO(27, pwm),   // 40
		rpGPIO(28, pwm),   // 41
		rpGPIO(29, pwm),   // 42
		power("IOVDD"),    // 43
		power("ADC_AVDD"), // 44
		power("VREG_VIN"), // 45

		// ── Left side (pins 46–60, bottom → top) ──
		power("VREG_VOUT"),  // 46
		special("USB_DM"),   // 47
		special("USB_DP"),   // 48
		power("USB_VDD"),    // 49
		power("IOVDD"),      // 50
		power("DVDD"),       // 51
		special("QSPI_SD3"), // 52
		special("QSPI_SCK"), // 53
		special("QSPI_SD0"), // 54
		special("QSPI_SD2"), // 55
		special("QSPI_SD1"), // 56
		special("QSPI_SS"),  // 57
		power("IOVDD"),      // 58
		power("DVDD"),       // 59
		power("IOVDD"),      // 60
	}
	for i := range pins {
		pins[i].PhysicalPin = i + 1
	}
	return pins
}

// buildRP2350BQFN80 returns all 80 pins of the RP2350B QFN-80 package.
func buildRP2350BQFN80() []Pin {
	const pwm = 12
	pins := []Pin{
		// ── Top side (pins 1–20, left → right) ──
		power("IOVDD"),  // 1
		rpGPIO(0, pwm),  // 2
		rpGPIO(1, pwm),  // 3
		rpGPIO(2, pwm),  // 4
		rpGPIO(3, pwm),  // 5
		rpGPIO(4, pwm),  // 6
		rpGPIO(5, pwm),  // 7
		rpGPIO(6, pwm),  // 8
		rpGPIO(7, pwm),  // 9
		power("IOVDD"),  // 10
		power("DVDD"),   // 11
		rpGPIO(8, pwm),  // 12
		rpGPIO(9, pwm),  // 13
		rpGPIO(10, pwm), // 14
		rpGPIO(11, pwm), // 15
		rpGPIO(12, pwm), // 16
		rpGPIO(13, pwm), // 17
		rpGPIO(14, pwm), // 18
		rpGPIO(15, pwm), // 19
		power("IOVDD"),  // 20

		// ── Right side (pins 21–40, top → bottom) ──
		rpGPIO(16, pwm),   // 21
		rpGPIO(17, pwm),   // 22
		rpGPIO(18, pwm),   // 23
		rpGPIO(19, pwm),   // 24
		special("TESTEN"), // 25
		special("XIN"),    // 26
		special("XOUT"),   // 27
		power("IOVDD"),    // 28
		power("DVDD"),     // 29
		special("SWCLK"),  // 30
		special("SWDIO"),  // 31
		special("RUN"),    // 32
		rpGPIO(20, pwm),   // 33
		rpGPIO(21, pwm),   // 34
		rpGPIO(22, pwm),   // 35
		rpGPIO(23, pwm),   // 36
		rpGPIO(24, pwm),   // 37
		rpGPIO(25, pwm),   // 38
		rpGPIO(26, pwm),   // 39
		rpGPIO(27, pwm),   // 40

		// ── Bottom side (pins 41–60, right → left) ──
		rpGPIO(28, pwm), // 41
		rpGPIO(29, pwm), // 42
		rpGPIO(30, pwm), // 43
		rpGPIO(31, pwm), // 44
		rpGPIO(32, pwm), // 45
		rpGPIO(33, pwm), // 46
		rpGPIO(34, pwm), // 47
		rpGPIO(35, pwm), // 48
		power("IOVDD"),  // 49
		rpGPIO(36, pwm), // 50
		rpGPIO(37, pwm), // 51
		rpGPIO(38, pwm), // 52
		rpGPIO(39, pwm), // 53
		rpGPIO(40, pwm), // 54
		rpGPIO(41, pwm), // 55
		rpGPIO(42, pwm), // 56
		rpGPIO(43, pwm), // 57
		power("IOVDD"),  // 58
		rpGPIO(44, pwm), // 59
		rpGPIO(45, pwm), // 60

		// ── Left side (pins 61–80, bottom → top) ──
		rpGPIO(46, pwm),     // 61
		rpGPIO(47, pwm),     // 62
		power("IOVDD"),      // 63
		power("ADC_AVDD"),   // 64
		power("VREG_VIN"),   // 65
		power("VREG_VOUT"),  // 66
		special("USB_DM"),   // 67
		special("USB_DP"),   // 68
		power("USB_VDD"),    // 69
		power("IOVDD"),      // 70
		power("DVDD"),       // 71
		special("QSPI_SD3"), // 72
		special("QSPI_SCK"), // 73
		special("QSPI_SD0"), // 74
		special("QSPI_SD2"), // 75
		special("QSPI_SD1"), // 76
		special("QSPI_SS"),  // 77
		power("IOVDD"),      // 78
		power("DVDD"),       // 79
		power("IOVDD"),      // 80
	}
	for i := range pins {
		pins[i].PhysicalPin = i + 1
	}
	return pins
}

// rpGPIO builds a single GPIO pin with all peripheral functions
// computed from the GPIO number and chip parameters.
//
// Peripheral mapping rules (shared by RP2040 and RP2350):
//
//	SPI:  bus = (gpio/8) % 2,  role cycles [RX, CSn, SCK, TX] on gpio % 4
//	I2C:  bus = (gpio/2) % 2,  SDA on even GPIOs, SCL on odd
//	UART: bus pattern [0,1,1,0] per group of 4; TX on offset 0,3; RX on 1,2
//	PWM:  slice = (gpio/2) % pwmSlices, channel A if even, B if odd
//	ADC:  GP26=ADC0, GP27=ADC1, GP28=ADC2, GP29=ADC3
func rpGPIO(num, pwmSlices int) Pin {
	gpioF := Function{Name: "GPIO", Category: "GPIO"}
	pio := Function{Name: "PIO", Category: "PIO"}

	funcs := []Function{gpioF}

	// SPI
	spiBus := (num / 8) % 2
	spiRoles := [4]string{"RX", "CSn", "SCK", "TX"}
	funcs = append(funcs, Function{
		Name:     "SPI" + itoa(spiBus) + " " + spiRoles[num%4],
		Category: "SPI",
	})

	// I2C
	i2cBus := (num / 2) % 2
	i2cPin := "SDA"
	if num%2 == 1 {
		i2cPin = "SCL"
	}
	funcs = append(funcs, Function{
		Name:     "I2C" + itoa(i2cBus) + " " + i2cPin,
		Category: "I2C",
	})

	// UART
	uartBusLookup := [4]int{0, 1, 1, 0}
	uartBus := uartBusLookup[(num/4)%4]
	uartDir := "RX"
	if num%4 == 0 || num%4 == 3 {
		uartDir = "TX"
	}
	funcs = append(funcs, Function{
		Name:     "UART" + itoa(uartBus) + " " + uartDir,
		Category: "UART",
	})

	// PWM
	pwmSlice := (num / 2) % pwmSlices
	pwmCh := "A"
	if num%2 == 1 {
		pwmCh = "B"
	}
	funcs = append(funcs, Function{
		Name:     "PWM" + itoa(pwmSlice) + " " + pwmCh,
		Category: "PWM",
	})

	funcs = append(funcs, pio)

	// ADC (GP26-GP29 only)
	adcCh := -1
	if num >= 26 && num <= 29 {
		adcCh = num - 26
		funcs = append(funcs, Function{
			Name:     "ADC" + itoa(adcCh),
			Category: "ADC",
		})
	}

	return Pin{
		Label:      "GP" + itoa(num),
		GPIO:       num,
		IsGPIO:     true,
		Functions:  funcs,
		ADCChannel: adcCh,
		PWMSlice:   pwmSlice,
		PWMChannel: pwmCh,
	}
}
