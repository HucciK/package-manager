package internal

import (
	"archive/zip"
	"fmt"
	"github.com/HucciK/package-manager/internal/core"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ZipHandler interface {
	GetWriter() *zip.Writer
	WriteFile(file *os.File, subPath string) error
	CloseWriter() error
	Unzip(dst, zipPath string) error
}

type SSHClient interface {
	UploadFile(filePath string) error
	DownloadFile(name, ver string) (string, error)
}

type Manager struct {
	zip ZipHandler
	ssh SSHClient
}

func NewManager(z ZipHandler, s SSHClient) *Manager {
	return &Manager{
		zip: z,
		ssh: s,
	}
}

func (m *Manager) CreatePacket(packet core.Packet) error {
	zipWriter := m.zip.GetWriter()

	for _, target := range packet.Targets {
		builder := core.NewTargetBuilder(target)
		files, err := builder.Build()
		if err != nil {
			return fmt.Errorf("can't parse files from packet targets: %w", err)
		}

		var subDir string
		for _, f := range files {
			file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("can't open file from target dir: %w", err)
			}

			stat, err := file.Stat()
			if err != nil {
				return fmt.Errorf("can't get file stat: %w", err)
			}

			if stat.IsDir() {
				if subDir == "" {
					subDir = filepath.Base(f)
				}

				if filepath.Base(f) != subDir {
					subDir = path.Join(subDir, filepath.Base(f))
				}

				_, err = zipWriter.Create(subDir + "/")
				if err != nil {
					return fmt.Errorf("can't create sub dir in zip file: %w", err)
				}
				continue
			}

			if err := m.zip.WriteFile(file, subDir); err != nil {
				return fmt.Errorf("can't write file into zip: %w", err)
			}
		}
	}

	if err := m.zip.CloseWriter(); err != nil {
		return fmt.Errorf("can't close zip writer: %w", err)
	}

	if err := m.ssh.UploadFile(packet.ZipName()); err != nil {
		return fmt.Errorf("can't upload file: %w", err)
	}
	log.Printf("Successfully uploaded. %s", packet.ZipName())

	return nil
}

func (m *Manager) UpdatePackets(packages core.Packages) error {
	var downloaded []string

	for _, p := range packages.Packets {
		filePath, err := m.ssh.DownloadFile(p.Name, p.Ver)
		if err != nil {
			return fmt.Errorf("can't download file: %w", err)
		}

		if filePath == "" {
			continue
		}

		dst := path.Join(strings.TrimSuffix(filePath, filepath.Base(filePath)), p.Name)
		if err := m.zip.Unzip(dst, filePath); err != nil {
			return fmt.Errorf("can't unzip downloaded file: %w", err)
		}
		downloaded = append(downloaded, p.Name)
	}
	log.Printf("Succesfully updated %d/%d packets.", len(downloaded), len(packages.Packets))

	return nil
}
