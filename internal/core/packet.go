package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Packet struct {
	Name    string   `json:"name"`
	Ver     string   `json:"ver"`
	Targets []Target `json:"targets"`
}

func NewPacket(path string) (Packet, error) {
	var pack Packet

	file, err := os.Open(path)
	if err != nil {
		return pack, fmt.Errorf("can't open packet file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return pack, fmt.Errorf("can't read packet file: %w", err)
	}

	if err := json.Unmarshal(content, &pack); err != nil {
		return pack, fmt.Errorf("can't unmarshall packet file to json: %w", err)
	}

	return pack, nil
}

func (p Packet) ZipName() string {
	return fmt.Sprintf("%s_v%s.zip", p.Name, p.Ver)
}

type Packages struct {
	Packets []Packet `json:"packages"`
}

func NewPackages(path string) (Packages, error) {
	var pack Packages

	file, err := os.Open(path)
	if err != nil {
		return pack, fmt.Errorf("can't open packet file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return pack, fmt.Errorf("can't read packet file: %w", err)
	}

	if err := json.Unmarshal(content, &pack); err != nil {
		return pack, fmt.Errorf("can't unmarshall packet file to json: %w", err)
	}

	return pack, nil
}
