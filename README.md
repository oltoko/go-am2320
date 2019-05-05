[![Build Status](https://travis-ci.org/oltoko/go-am2320.svg?branch=master)](https://travis-ci.org/oltoko/go-am2320)

# go-am2320
Code to access an AM2320 via i2c on Raspberry Pi. Also see [Datasheet](https://akizukidenshi.com/download/ds/aosong/AM2320.pdf).

## Usage

```Go
package main

import (
    "log"

    "github.com/oltoko/go-am2320"
)

func main() {

    sensor := am2320.Create(am2320.DefaultI2CAddr)

    values, err := sensor.Read()
    if err != nil {
        log.Fatalln("Failed to read from Sensor", err)
    }

    log.Printf("%.2f °C", values.Temperature)
    log.Printf("%.2f %%", values.Humidity)
}
```
