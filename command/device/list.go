// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package device

import (
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// DeviceInfo contains the most interesting
// parameters of an Arduino IoT Cloud device.
type DeviceInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Board  string `json:"board"`
	Serial string `json:"serial-number"`
	FQBN   string `json:"fqbn"`
}

// List command is used to list
// the devices of Arduino IoT Cloud.
func List() ([]DeviceInfo, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	foundDevices, err := iotClient.DeviceList()
	if err != nil {
		return nil, err
	}

	var devices []DeviceInfo
	for _, foundDev := range foundDevices {
		dev := DeviceInfo{
			Name:   foundDev.Name,
			ID:     foundDev.Id,
			Board:  foundDev.Type,
			Serial: foundDev.Serial,
			FQBN:   foundDev.Fqbn,
		}
		devices = append(devices, dev)
	}

	return devices, nil
}
