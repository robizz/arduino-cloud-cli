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
	"errors"
	"fmt"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/sirupsen/logrus"
)

// CreateParams contains the parameters needed
// to find the device to be provisioned.
type CreateParams struct {
	Name string  // Device name
	Port *string // Serial port - Optional - If omitted then each serial port is analyzed
	Fqbn *string // Board FQBN - Optional - If omitted then the first device found gets selected
}

type board struct {
	fqbn   string
	serial string
	dType  string
	port   string
}

// Create command is used to provision a new arduino device
// and to add it to Arduino IoT Cloud.
func Create(params *CreateParams) (*DeviceInfo, error) {
	comm, err := cli.NewCommander()
	if err != nil {
		return nil, err
	}

	ports, err := comm.BoardList()
	if err != nil {
		return nil, err
	}
	board := boardFromPorts(ports, params)
	if board == nil {
		err = errors.New("no board found")
		return nil, err
	}

	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	logrus.Info("Creating a new device on the cloud")
	dev, err := iotClient.DeviceCreate(board.fqbn, params.Name, board.serial, board.dType)
	if err != nil {
		return nil, err
	}

	prov := &provision{
		Commander: comm,
		Client:    iotClient,
		board:     board,
		id:        dev.Id,
	}
	if err = prov.run(); err != nil {
		// TODO: retry to delete the device if it returns an error.
		// In alternative: encapsulate also this error.
		iotClient.DeviceDelete(dev.Id)
		err = fmt.Errorf("%s: %w", "cannot provision device", err)
		return nil, err
	}

	devInfo := &DeviceInfo{
		Name:   dev.Name,
		ID:     dev.Id,
		Board:  dev.Type,
		Serial: dev.Serial,
		FQBN:   dev.Fqbn,
	}
	return devInfo, nil
}

// boardFromPorts returns a board that matches all the criteria
// passed in. If no criteria are passed, it returns the first board found.
func boardFromPorts(ports []*rpc.DetectedPort, params *CreateParams) *board {
	for _, port := range ports {
		if portFilter(port, params) {
			continue
		}
		boardFound := boardFilter(port.Boards, params)
		if boardFound != nil {
			t := strings.Split(boardFound.Fqbn, ":")[2]
			b := &board{boardFound.Fqbn, port.SerialNumber, t, port.Address}
			return b
		}
	}

	return nil
}

// portFilter filters out the given port in the following cases:
// - if the port parameter does not match the actual port address.
// - if the the detected port does not contain any board.
// It returns:
// true -> to skip the port
// false -> to keep the port
func portFilter(port *rpc.DetectedPort, params *CreateParams) bool {
	if len(port.Boards) == 0 {
		return true
	}
	if params.Port != nil && *params.Port != port.Address {
		return true
	}
	return false
}

// boardFilter looks for a board which has the same fqbn passed as parameter.
// If fqbn parameter is nil, then the first board found is returned.
// It returns:
// - a board if it is found.
// - nil if no board matching the fqbn parameter is found.
func boardFilter(boards []*rpc.BoardListItem, params *CreateParams) (board *rpc.BoardListItem) {
	if params.Fqbn == nil {
		return boards[0]
	}
	for _, b := range boards {
		if b.Fqbn == *params.Fqbn {
			return b
		}
	}
	return
}
