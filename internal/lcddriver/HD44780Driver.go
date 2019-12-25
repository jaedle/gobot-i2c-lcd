package lcddrive

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"time"
)

const defaultName = "HD44780"
const command2LineLcd = 0x08

type HD44780Driver struct {
	name       string
	connector  i2c.Connector
	connection i2c.Connection
	i2c.Config
}

func NewHD44780Driver(a i2c.Connector, options ...func(i2c.Config)) *HD44780Driver {
	result := &HD44780Driver{
		name:      gobot.DefaultName(defaultName),
		connector: a,
		Config:    i2c.NewConfig(),
	}

	for _, option := range options {
		option(result)
	}

	return result
}

func (h *HD44780Driver) Name() string { return h.name }

func (h *HD44780Driver) SetName(n string) { h.name = n }

func (h *HD44780Driver) Connection() gobot.Connection { return h.connector.(gobot.Connection) }

const enable byte = 0B00000100

const lcdStartTime = 1000 * time.Millisecond
const time4BitMode = 4500 * time.Millisecond

func (h *HD44780Driver) Start() error {
	if err := h.setupConnection(); err != nil {
		return err
	}

	time.Sleep(lcdStartTime)

	if err := h.set4BitMode(); err != nil {
		return err
	}

	if err := h.setLcdFunction(); err != nil {
		return err
	}

	if err := h.enableDisplay(); err != nil {
		return err
	}

	if err := h.Clear(); err != nil {
		return err
	}

	if err := h.Home(); err != nil {
		return err
	}

	return nil
}

func (h *HD44780Driver) enableDisplay() error {
	const displayControl = 0x08
	const displayOn = 0x04
	if err := h.sendCommand(displayControl | displayOn); err != nil {
		return err
	}
	return nil
}

func (h *HD44780Driver) setupConnection() error {
	bus := h.GetBusOrDefault(h.connector.GetDefaultBus())
	address := h.GetAddressOrDefault(0)
	connection, err := h.connector.GetConnection(address, bus)
	h.connection = connection

	return err
}

func (h *HD44780Driver) setLcdFunction() error {
	const commandLcdFunctionSet byte = 0x20
	var value = commandLcdFunctionSet | command2LineLcd
	return h.sendCommand(value)
}

const Retries4BitMode = 3

func (h *HD44780Driver) set4BitMode() error {
	for i := 0; i < Retries4BitMode; i++ {
		var command4BitMode byte = 0x03 << 4
		if err := h.send4BitCommand(command4BitMode); err != nil {
			return err
		}
		time.Sleep(time4BitMode)
	}

	return h.send4BitCommand(0x02 << 4)
}

const commandHome = 0x02
const homeCommandDuration = 2000 * time.Microsecond

func (h *HD44780Driver) Home() error {
	err := h.sendCommand(commandHome)
	if err != nil {
		return err
	}

	time.Sleep(homeCommandDuration)
	return nil
}

func (h *HD44780Driver) Write(c rune) error {
	return h.sendData(byte(c))
}

func (h *HD44780Driver) WriteString(s string) error {
	for _, c := range s {
		if err := h.Write(c); err != nil {
			return err
		}
	}

	return nil
}

const setAddressCommand = 0x80
func (h *HD44780Driver) SetCursor(row byte, col byte) error {
	if !validColumn(col) {
		return fmt.Errorf("invalid col coordinate: %d", col)
	} else if !validRow(row) {
		return fmt.Errorf("invalid row coordinate: %d", row)
	}

	return h.sendCommand(setAddressCommand | col + rowOffsets()[row])
}

func rowOffsets() []byte {
	return []byte{0x00, 0x40}
}

func validRow(r byte) bool {
	return !(r > 1 || r < 0)
}

func validColumn(c byte) bool {
	return !(c > 15 || c < 0)
}

func (h *HD44780Driver) sendCommand(command byte) error {
	if err := h.send4BitCommand(highNibble(command)); err != nil {
		return err
	}
	if err := h.send4BitCommand(lowNibble(command)); err != nil {
		return err
	}

	return nil
}

func (h *HD44780Driver) sendData(data byte) error {
	if err := h.send4BitCommand(highNibble(data) | 1); err != nil {
		return err
	}
	if err := h.send4BitCommand(lowNibble(data) | 1); err != nil {
		return err
	}

	return nil
}

func lowNibble(value byte) byte {
	return (value << 4) & 0xf0
}

func highNibble(value byte) byte {
	return value & 0xf0
}

const enablePulseTime = 1 * time.Microsecond
const commandSettleTime = 50 * time.Microsecond

func (h *HD44780Driver) send4BitCommand(command byte) error {
	if err := h.connection.WriteByte(withEnable(command)); err != nil {
		return err
	}
	time.Sleep(enablePulseTime)

	if err := h.connection.WriteByte(withoutEnable(command)); err != nil {
		return err
	}
	time.Sleep(commandSettleTime)

	return nil
}

func withoutEnable(command byte) byte {
	return command & ^enable
}

func withEnable(command byte) byte {
	return command | enable
}

func (h *HD44780Driver) Halt() error { return nil }

const commandClearDisplay = 0x01
const clearTime = 2000 * time.Microsecond

func (h *HD44780Driver) Clear() error {
	if err := h.sendCommand(commandClearDisplay); err != nil {
		return err
	}

	time.Sleep(clearTime)
	return nil
}
