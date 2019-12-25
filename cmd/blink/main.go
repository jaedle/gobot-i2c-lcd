package main

import (
	lcddriver "github.com/jaedle/gobot-i2c-lcd/internal/lcddriver"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
	r := raspi.NewAdaptor()
	led := gpio.NewLedDriver(r, "7")

	lcd := lcddriver.NewHD44780Driver(r, i2c.WithAddress(0x27))

	work := func() {
		lcd.Home()
		gobot.Every(1000*time.Millisecond, func() {
			led.Toggle()
			lcd.SetCursor(0, 1)
			lcd.WriteString("ASDF")
			lcd.SetCursor(1, 0)
			lcd.Write('A')
		})
	}

	robot := gobot.NewRobot("blinkBot",
		[]gobot.Connection{r},
		[]gobot.Device{led, lcd},
		work,
	)

	robot.Start()
}
