// Package pindata provides comprehensive pin mapping data for Raspberry Pi
// Pico boards and bare RP-series microcontroller chips.
package pindata

// Board identifies a specific board or bare chip variant.
type Board int

const (
	Pico        Board = iota // RP2040-based Pico board
	Pico2                    // RP2350A-based Pico 2 board
	RP2040Chip               // Bare RP2040 chip (30 GPIOs)
	RP2350AChip              // Bare RP2350A chip (30 GPIOs)
	RP2350BChip              // Bare RP2350B chip (48 GPIOs)
)

func (b Board) String() string {
	switch b {
	case Pico:
		return "Raspberry Pi Pico (RP2040)"
	case Pico2:
		return "Raspberry Pi Pico 2 (RP2350)"
	case RP2040Chip:
		return "RP2040 (bare chip)"
	case RP2350AChip:
		return "RP2350A (bare chip)"
	case RP2350BChip:
		return "RP2350B (bare chip)"
	default:
		return "Unknown"
	}
}

// IsChip returns true for bare chip variants (no physical board layout).
func (b Board) IsChip() bool {
	return b == RP2040Chip || b == RP2350AChip || b == RP2350BChip
}

// Function describes a peripheral function a GPIO can serve.
type Function struct {
	Name     string // e.g. "SPI0 RX", "I2C0 SDA", "UART0 TX"
	Category string // "SPI", "I2C", "UART", "PWM", "ADC", "PIO", "GPIO"
}

// Pin represents a single physical pin on the 40-pin header.
type Pin struct {
	PhysicalPin int        // 1-40 position on the header
	Label       string     // Silk-screen label, e.g. "GP0", "GND", "3V3(OUT)"
	GPIO        int        // GPIO number (-1 for power/ground pins)
	IsGPIO      bool       // true if this is a user-accessible GPIO
	IsPower     bool       // true for power pins (3V3, VSYS, VBUS)
	IsGround    bool       // true for GND pins
	IsSpecial   bool       // true for RUN, ADC_VREF, etc.
	Functions   []Function // available peripheral functions
	ADCChannel  int        // ADC channel number (-1 if not ADC capable)
	PWMSlice    int        // PWM slice number (-1 if not applicable)
	PWMChannel  string     // "A" or "B" within the slice
}

// BoardSpec holds the complete specification for a board variant.
type BoardSpec struct {
	Board       Board
	Name        string
	Chip        string
	FlashKB     int
	RAMKB       int
	CPUCores    int
	CPUArch     string
	MaxClockMHz int
	PWMChannels int
	PIOBlocks   int
	PIOSMs      int // state machines per PIO block
	USBSupport  bool
	Pins        []Pin
}

// GetSpec returns the full board specification for the given board or chip.
func GetSpec(b Board) BoardSpec {
	switch b {
	case Pico2:
		return BoardSpec{
			Board:       Pico2,
			Name:        "Raspberry Pi Pico 2",
			Chip:        "RP2350A",
			FlashKB:     4096,
			RAMKB:       520,
			CPUCores:    2,
			CPUArch:     "ARM Cortex-M33 / Hazard3 RISC-V",
			MaxClockMHz: 150,
			PWMChannels: 24,
			PIOBlocks:   3,
			PIOSMs:      4,
			USBSupport:  true,
			Pins:        buildPins(),
		}
	case RP2040Chip:
		return BoardSpec{
			Board:       RP2040Chip,
			Name:        "RP2040",
			Chip:        "RP2040",
			FlashKB:     0, // external flash required
			RAMKB:       264,
			CPUCores:    2,
			CPUArch:     "ARM Cortex-M0+",
			MaxClockMHz: 133,
			PWMChannels: 16,
			PIOBlocks:   2,
			PIOSMs:      4,
			USBSupport:  true,
			Pins:        buildRP2040QFN56(),
		}
	case RP2350AChip:
		return BoardSpec{
			Board:       RP2350AChip,
			Name:        "RP2350A",
			Chip:        "RP2350A",
			FlashKB:     0, // external flash required
			RAMKB:       520,
			CPUCores:    2,
			CPUArch:     "ARM Cortex-M33 / Hazard3 RISC-V",
			MaxClockMHz: 150,
			PWMChannels: 24,
			PIOBlocks:   3,
			PIOSMs:      4,
			USBSupport:  true,
			Pins:        buildRP2350AQFN60(),
		}
	case RP2350BChip:
		return BoardSpec{
			Board:       RP2350BChip,
			Name:        "RP2350B",
			Chip:        "RP2350B",
			FlashKB:     0, // external flash required
			RAMKB:       520,
			CPUCores:    2,
			CPUArch:     "ARM Cortex-M33 / Hazard3 RISC-V",
			MaxClockMHz: 150,
			PWMChannels: 24,
			PIOBlocks:   3,
			PIOSMs:      4,
			USBSupport:  true,
			Pins:        buildRP2350BQFN80(),
		}
	default: // Pico
		return BoardSpec{
			Board:       Pico,
			Name:        "Raspberry Pi Pico",
			Chip:        "RP2040",
			FlashKB:     2048,
			RAMKB:       264,
			CPUCores:    2,
			CPUArch:     "ARM Cortex-M0+",
			MaxClockMHz: 133,
			PWMChannels: 16,
			PIOBlocks:   2,
			PIOSMs:      4,
			USBSupport:  true,
			Pins:        buildPins(),
		}
	}
}

// GPIOPins returns only the user-accessible GPIO pins.
func GPIOPins(spec BoardSpec) []Pin {
	var out []Pin
	for _, p := range spec.Pins {
		if p.IsGPIO {
			out = append(out, p)
		}
	}
	return out
}

// PinsForCategory returns GPIO pins that support the given function category.
func PinsForCategory(spec BoardSpec, category string) []Pin {
	var out []Pin
	for _, p := range spec.Pins {
		if !p.IsGPIO {
			continue
		}
		for _, f := range p.Functions {
			if f.Category == category {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// FunctionsForGPIO returns all available functions for a given GPIO number.
func FunctionsForGPIO(spec BoardSpec, gpio int) []Function {
	for _, p := range spec.Pins {
		if p.GPIO == gpio {
			return p.Functions
		}
	}
	return nil
}

// PinConflict describes a conflict when two peripherals share a GPIO.
type PinConflict struct {
	GPIO      int
	Function1 Function
	Function2 Function
}

// CheckConflicts checks a set of selected (gpio, function) pairs for conflicts.
func CheckConflicts(selections map[int]Function) []PinConflict {
	var conflicts []PinConflict
	gpioUsage := make(map[int]Function)
	for gpio, fn := range selections {
		if existing, ok := gpioUsage[gpio]; ok {
			conflicts = append(conflicts, PinConflict{
				GPIO:      gpio,
				Function1: existing,
				Function2: fn,
			})
		} else {
			gpioUsage[gpio] = fn
		}
	}
	return conflicts
}

func gpio(num int, funcs []Function, adcCh, pwmSlice int, pwmCh string) Pin {
	label := "GP" + itoa(num)
	return Pin{
		Label:      label,
		GPIO:       num,
		IsGPIO:     true,
		Functions:  funcs,
		ADCChannel: adcCh,
		PWMSlice:   pwmSlice,
		PWMChannel: pwmCh,
	}
}

func power(label string) Pin {
	return Pin{Label: label, GPIO: -1, IsPower: true}
}

func ground() Pin {
	return Pin{Label: "GND", GPIO: -1, IsGround: true}
}

func special(label string) Pin {
	return Pin{Label: label, GPIO: -1, IsSpecial: true}
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// buildPins constructs the 40-pin header in physical order.
// The Pico and Pico 2 share the same physical pinout.
func buildPins() []Pin {
	// SPI bus assignments repeat every 8 GPIOs:
	//   GP0-GP7:  SPI0 (RX=0,4  CSn=1,5  SCK=2,6  TX=3,7)
	//   GP8-GP15: SPI1 (RX=8,12 CSn=9,13 SCK=10,14 TX=11,15)
	//   GP16-GP23: SPI0 (RX=16,20 CSn=17,21 SCK=18,22 TX=19,23)

	// I2C bus assignments repeat every 8 GPIOs:
	//   Even GPIOs: SDA, Odd GPIOs: SCL
	//   GP0-GP7:  I2C0 (0,1) I2C1 (2,3) I2C0 (4,5) I2C1 (6,7)
	//   GP8-GP15: I2C0 (8,9) I2C1 (10,11) I2C0 (12,13) I2C1 (14,15)

	// UART assignments:
	//   GP0,1:   UART0 TX,RX
	//   GP4,5:   UART1 TX,RX
	//   GP8,9:   UART1 TX,RX
	//   GP12,13: UART0 TX,RX
	//   GP16,17: UART0 TX,RX
	//   GP20,21: UART1 TX,RX

	// PWM slices: GPIO / 2 (mod 8 for RP2040), channel A=even B=odd

	spi := func(bus, role int) Function {
		busStr := itoa(bus)
		roles := []string{"RX", "CSn", "SCK", "TX"}
		return Function{Name: "SPI" + busStr + " " + roles[role], Category: "SPI"}
	}

	i2c := func(bus int, isSCL bool) Function {
		busStr := itoa(bus)
		pin := "SDA"
		if isSCL {
			pin = "SCL"
		}
		return Function{Name: "I2C" + busStr + " " + pin, Category: "I2C"}
	}

	uart := func(bus int, isTX bool) Function {
		busStr := itoa(bus)
		dir := "RX"
		if isTX {
			dir = "TX"
		}
		return Function{Name: "UART" + busStr + " " + dir, Category: "UART"}
	}

	pwm := func(slice int, ch string) Function {
		return Function{Name: "PWM" + itoa(slice) + " " + ch, Category: "PWM"}
	}

	adc := func(ch int) Function {
		return Function{Name: "ADC" + itoa(ch), Category: "ADC"}
	}

	pio := Function{Name: "PIO", Category: "PIO"}
	gpioF := Function{Name: "GPIO", Category: "GPIO"}

	// Build all 40 pins in physical order (left=odd, right=even looking from top)
	pins := []Pin{
		// Pin 1-2
		gpio(0, []Function{gpioF, spi(0, 0), i2c(0, false), uart(0, true), pwm(0, "A"), pio}, -1, 0, "A"),
		gpio(1, []Function{gpioF, spi(0, 1), i2c(0, true), uart(0, false), pwm(0, "B"), pio}, -1, 0, "B"),
		// Pin 3-4
		ground(),
		gpio(2, []Function{gpioF, spi(0, 2), i2c(1, false), uart(0, false), pwm(1, "A"), pio}, -1, 1, "A"),
		// Pin 5-6
		gpio(3, []Function{gpioF, spi(0, 3), i2c(1, true), uart(0, true), pwm(1, "B"), pio}, -1, 1, "B"),
		gpio(4, []Function{gpioF, spi(0, 0), i2c(0, false), uart(1, true), pwm(2, "A"), pio}, -1, 2, "A"),
		// Pin 7-8
		gpio(5, []Function{gpioF, spi(0, 1), i2c(0, true), uart(1, false), pwm(2, "B"), pio}, -1, 2, "B"),
		ground(),
		// Pin 9-10
		gpio(6, []Function{gpioF, spi(0, 2), i2c(1, false), uart(1, false), pwm(3, "A"), pio}, -1, 3, "A"),
		gpio(7, []Function{gpioF, spi(0, 3), i2c(1, true), uart(1, true), pwm(3, "B"), pio}, -1, 3, "B"),
		// Pin 11-12
		gpio(8, []Function{gpioF, spi(1, 0), i2c(0, false), uart(1, true), pwm(4, "A"), pio}, -1, 4, "A"),
		gpio(9, []Function{gpioF, spi(1, 1), i2c(0, true), uart(1, false), pwm(4, "B"), pio}, -1, 4, "B"),
		// Pin 13-14
		ground(),
		gpio(10, []Function{gpioF, spi(1, 2), i2c(1, false), uart(1, false), pwm(5, "A"), pio}, -1, 5, "A"),
		// Pin 15-16
		gpio(11, []Function{gpioF, spi(1, 3), i2c(1, true), uart(1, true), pwm(5, "B"), pio}, -1, 5, "B"),
		gpio(12, []Function{gpioF, spi(1, 0), i2c(0, false), uart(0, true), pwm(6, "A"), pio}, -1, 6, "A"),
		// Pin 17-18
		gpio(13, []Function{gpioF, spi(1, 1), i2c(0, true), uart(0, false), pwm(6, "B"), pio}, -1, 6, "B"),
		ground(),
		// Pin 19-20
		gpio(14, []Function{gpioF, spi(1, 2), i2c(1, false), uart(0, false), pwm(7, "A"), pio}, -1, 7, "A"),
		gpio(15, []Function{gpioF, spi(1, 3), i2c(1, true), uart(0, true), pwm(7, "B"), pio}, -1, 7, "B"),
		// Pin 21-22
		gpio(16, []Function{gpioF, spi(0, 0), i2c(0, false), uart(0, true), pwm(0, "A"), pio}, -1, 0, "A"),
		gpio(17, []Function{gpioF, spi(0, 1), i2c(0, true), uart(0, false), pwm(0, "B"), pio}, -1, 0, "B"),
		// Pin 23-24
		ground(),
		gpio(18, []Function{gpioF, spi(0, 2), i2c(1, false), uart(0, false), pwm(1, "A"), pio}, -1, 1, "A"),
		// Pin 25-26
		gpio(19, []Function{gpioF, spi(0, 3), i2c(1, true), uart(0, true), pwm(1, "B"), pio}, -1, 1, "B"),
		gpio(20, []Function{gpioF, spi(0, 0), i2c(0, false), uart(1, true), pwm(2, "A"), pio}, -1, 2, "A"),
		// Pin 27-28
		gpio(21, []Function{gpioF, spi(0, 1), i2c(0, true), uart(1, false), pwm(2, "B"), pio}, -1, 2, "B"),
		ground(),
		// Pin 29-30
		gpio(22, []Function{gpioF, spi(0, 2), i2c(1, false), uart(1, false), pwm(3, "A"), pio}, -1, 3, "A"),
		special("RUN"),
		// Pin 31-32
		gpio(26, []Function{gpioF, spi(1, 2), i2c(1, false), uart(1, false), pwm(5, "A"), pio, adc(0)}, 0, 5, "A"),
		gpio(27, []Function{gpioF, spi(1, 3), i2c(1, true), uart(1, true), pwm(5, "B"), pio, adc(1)}, 1, 5, "B"),
		// Pin 33-34
		ground(),
		gpio(28, []Function{gpioF, spi(1, 0), i2c(0, false), uart(0, true), pwm(6, "A"), pio, adc(2)}, 2, 6, "A"),
		// Pin 35-36
		special("ADC_VREF"),
		power("3V3(OUT)"),
		// Pin 37-38
		special("3V3_EN"),
		ground(),
		// Pin 39-40
		power("VSYS"),
		power("VBUS"),
	}

	// Assign physical pin numbers
	for i := range pins {
		pins[i].PhysicalPin = i + 1
	}

	return pins
}

// CategoryColor returns a hex color string for a function category.
func CategoryColor(category string) uint32 {
	switch category {
	case "SPI":
		return 0xFF9800 // orange
	case "I2C":
		return 0x2196F3 // blue
	case "UART":
		return 0x4CAF50 // green
	case "PWM":
		return 0x9C27B0 // purple
	case "ADC":
		return 0xF44336 // red
	case "PIO":
		return 0x00BCD4 // teal
	case "GPIO":
		return 0x607D8B // blue-grey
	default:
		return 0x9E9E9E // grey
	}
}

// AllCategories returns the list of peripheral categories.
func AllCategories() []string {
	return []string{"GPIO", "SPI", "I2C", "UART", "PWM", "ADC", "PIO"}
}
