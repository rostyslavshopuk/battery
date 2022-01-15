// battery
// Copyright (C) 2016-2017 Karol 'Kenji Takahashi' Woźniak
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package battery_test

import (
	"fmt"
	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
	"time"
)

type MockedIOUtil struct {
	fileSystem map[string][]byte
	dirs       map[string][]fs.FileInfo
}

func (m *MockedIOUtil) ReadDir(dirname string) ([]fs.FileInfo, error) {
	v, ok := m.dirs[dirname]
	if !ok {
		return nil, fmt.Errorf("directory %v does not exist", dirname)
	}

	return v, nil
}

func (m *MockedIOUtil) ReadFile(fileName string) ([]byte, error) {
	f, ok := m.fileSystem[fileName]
	if !ok {
		return nil, fmt.Errorf("file %s does not exist", fileName)
	}

	return f, nil
}

type DirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
	t     fs.FileMode
	info  fs.FileInfo
}

func (d DirEntry) Size() int64 {
	return 0
}

func (d DirEntry) Mode() fs.FileMode {
	return d.mode
}

func (d DirEntry) ModTime() time.Time {
	return time.Now()
}

func (d DirEntry) Sys() interface{} {
	return nil
}

func (d DirEntry) Name() string {
	return d.name
}

func (d DirEntry) IsDir() bool {
	return d.isDir
}

var _ fs.FileInfo = (*DirEntry)(nil)

const sysClassPowerSupply = "/sys/class/power_supply"

func (m *MockedIOUtil) addBattery(batteryName string) {
	v, ok := m.dirs[sysClassPowerSupply]
	if !ok {
		m.dirs[sysClassPowerSupply] = []fs.FileInfo{}
		v = m.dirs[sysClassPowerSupply]
	}

	v = append(v, DirEntry{isDir: true, name: batteryName})
	m.dirs[sysClassPowerSupply] = v
}

func (m *MockedIOUtil) createBatteryEntry(batteryName string, fileName string, fileContent string) {
	filePath := fmt.Sprintf("%s/%s/%s",
		"/sys/class/power_supply",
		batteryName,
		fileName,
	)
	m.fileSystem[filePath] = []byte(fileContent)
}

/*
	Some devices, like the Apple Magic Mouse 2, report with their HID driver a battery in the
	/sys/class/power_supply path, named hid-MACADDR-battery.

	Although this is a battery, only a couple of fields are present, and this should return an error.
	This test is used to reproduce this situation, and to verify that the call to GetBatteries
	returns a battery.Errors{nil, error}

	$ ls -la /sys/class/power_supply/hid-MACADDR-battery/
	total 0
	drwxr-xr-x 5 root root    0 Jan 15 13:40 .
	drwxr-xr-x 3 root root    0 Jan 15 13:40 ..
	-r--r--r-- 1 root root 4096 Jan 15 13:40 capacity
	lrwxrwxrwx 1 root root    0 Jan 15 13:48 device -> ../../../0005:004C:0269.0013
	drwxr-xr-x 3 root root    0 Jan 15 13:40 hwmon9
	-r--r--r-- 1 root root 4096 Jan 15 13:40 model_name
	-r--r--r-- 1 root root 4096 Jan 15 13:48 online
	drwxr-xr-x 2 root root    0 Jan 15 13:48 power
	lrwxrwxrwx 1 root root    0 Jan 15 13:48 powers -> ../../../0005:004C:0269.0013
	-r--r--r-- 1 root root 4096 Jan 15 13:48 present
	-r--r--r-- 1 root root 4096 Jan 15 13:40 scope
	-r--r--r-- 1 root root 4096 Jan 15 13:40 status
	lrwxrwxrwx 1 root root    0 Jan 15 13:40 subsystem -> ../../../../../../../../../../../../class/power_supply
	-r--r--r-- 1 root root 4096 Jan 15 13:40 type
	-rw-r--r-- 1 root root 4096 Jan 15 13:40 uevent
	drwxr-xr-x 2 root root    0 Jan 15 13:40 wakeup66

	$ cat /sys/class/power_supply/hid-MACADDR-battery/model_name
	John Doe's Mouse
*/
func TestHidBattery(t *testing.T) {
	mockedIo := MockedIOUtil{
		fileSystem: map[string][]byte{},
		dirs:       map[string][]fs.FileInfo{},
	}
	battery.MyIOUtil = &mockedIo

	// Add real battery
	bat0 := "BAT0"
	mockedIo.addBattery(bat0)
	mockedIo.createBatteryEntry(bat0, "type", "Battery\n")
	mockedIo.createBatteryEntry(bat0, "energy_now", "10520000\n")
	mockedIo.createBatteryEntry(bat0, "voltage_now", "14427000\n")
	mockedIo.createBatteryEntry(bat0, "voltage_min_design", "15440000\n")
	mockedIo.createBatteryEntry(bat0, "energy_full", "55890000\n")
	mockedIo.createBatteryEntry(bat0, "energy_full_design", "57000000\n")
	mockedIo.createBatteryEntry(bat0, "power_now", "24037000\n")
	mockedIo.createBatteryEntry(bat0, "status", "Discharging\n")

	hidBatteryDir := "hid-aa:bb:cc:dd:ee:ff-battery"
	mockedIo.addBattery(hidBatteryDir)
	mockedIo.createBatteryEntry(hidBatteryDir, "type", "Battery\n")
	mockedIo.createBatteryEntry(hidBatteryDir, "status", "Discharging\n")

	allBatteries, err := battery.GetAll()
	assert.Len(t, allBatteries, 2)
	assert.IsType(t, battery.Errors{}, err)
	assert.NotNil(t, err)

	errors := err.(battery.Errors)
	assert.Nil(t, errors[0])
	assert.NotNil(t, errors[1])
	fmt.Printf("batteries=%v\n", allBatteries)
}
