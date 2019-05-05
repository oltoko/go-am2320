// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package am2320 contains the Code to read Temperature and Humidity from
// the environment with the Aosong AM2320 sensor.
// Please see the Datasheet for more information: https://akizukidenshi.com/download/ds/aosong/AM2320.pdf
package am2320

import (
	"errors"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// DefaultI2CAddr is the default Address of the AM2320.
	DefaultI2CAddr = 0x5c
	i2CSlave       = 0x0703
)

// AM2320 represents the sensor. In most cases you should initialize it with
// the DefaultI2CAddr as address.
type AM2320 struct {
	Address int
}

// SensorValues contains the results of reading the current Temperature and Humidity
// detected by the Sensor.
//
// The Temperature is in °C between -40 to 80
// Humidity is in % between 100 and 0
type SensorValues struct {
	Temperature, Humidity float32
}

// Read is used to read the actual Temperature and Humidity from the AM2320 Sensor
func (am2320 AM2320) Read() (*SensorValues, error) {

	f, err := os.OpenFile("/dev/i2c-1", syscall.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	unix.IoctlSetInt(int(f.Fd()), i2CSlave, am2320.Address)

	// wake AM2320 up, goes to sleep to not warm up and affect the humidity sensor
	// This write will fail as AM2320 won't ACK this write
	f.Write([]byte{0x00})
	// Wait at least 0.8ms, at most 3ms
	time.Sleep(1000 * time.Microsecond)
	// write at addr 0x03, start reg = 0x00, num regs = 0x04
	f.Write([]byte{0x03, 0x00, 0x04})
	// Wait at least 1.5ms for result
	time.Sleep(1600 * time.Microsecond)

	// Read out 8 bytes of result data
	// Byte 0: Should be Modbus function code 0x03
	// Byte 1: Should be number of registers to read (0x04)
	// Byte 2: Humidity msb
	// Byte 3: Humidity lsb
	// Byte 4: Temperature msb
	// Byte 5: Temperature lsb
	// Byte 6: CRC lsb byte
	// Byte 7: CRC msb byte
	result := make([]byte, 8)
	if _, err := f.Read(result); err != nil {
		return nil, err
	}

	if calcCrc16(result[0:6]) != combineBytes(result[7], result[6]) {
		return nil, errors.New("Failed to read from AM230 >> CRC check failed")
	}

	temp := float32(combineBytes(result[4], result[5])) / 10
	hum := float32(combineBytes(result[2], result[3])) / 10

	return &SensorValues{temp, hum}, nil
}

func calcCrc16(data []byte) int16 {

	crc := 0xffff

	for _, x := range data {
		crc = crc ^ int(x)

		for i := 0; i < 8; i++ {
			if (crc & 0x0001) == 0x0001 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return int16(crc)
}

func combineBytes(msb, lsb byte) int16 {
	return int16(msb)<<8 | int16(lsb)
}
