package internal

import (
	"fmt"
	"github.com/HucciK/package-manager/config"
	"github.com/HucciK/package-manager/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultSSHPort = "22"

type sshClient struct {
	config config.Config
	cli    *sftp.Client
}

func NewSSHClient(config config.Config) (SSHClient, error) {
	addr := fmt.Sprintf("%s:%s", config.Address, DefaultSSHPort)

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("can't dial ssh conn: %w", err)
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("can't dial sftp conn: %w", err)
	}

	return &sshClient{cli: sftp, config: config}, nil
}

func (s *sshClient) UploadFile(name string) error {

	srcFile, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("can't open source file by specified path: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := s.cli.Create(path.Join(s.config.PacketsDir, name))
	if err != nil {
		return fmt.Errorf("can't create file on target server: %w", err)
	}
	defer dstFile.Close()

	bytes, err := dstFile.ReadFrom(srcFile)
	if err != nil {
		return fmt.Errorf("can't write to dst file: %w", err)
	}

	if bytes == 0 {
		return errors.ErrZeroBytesWriten
	}

	return nil
}

func (s *sshClient) DownloadFile(targetName, targetVer string) (string, error) {
	packets, err := s.cli.ReadDir(s.config.PacketsDir)
	if err != nil {
		return "", fmt.Errorf("can't get dir with packets: %w", err)
	}

	for _, packet := range packets {
		packetInfo := strings.Split(packet.Name(), "_v")
		if len(packetInfo) < 2 {
			continue
		}

		packetName := packetInfo[0]
		packetVer := strings.TrimSuffix(packetInfo[1], filepath.Ext(packetInfo[1]))

		if packetName == targetName {
			valid, err := s.compareVer(packetVer, targetVer)
			if err != nil {
				return "", fmt.Errorf("can't compare versions: %w", err)
			}

			if !valid {
				continue
			}

			rel, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("can't get current paht: %w", err)
			}
			dstPath := path.Join(rel, s.config.PacketsDir, packet.Name())

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return "", fmt.Errorf("can't create local file: %w", err)
			}

			srcFile, err := s.cli.Open(path.Join(s.config.PacketsDir, packet.Name()))
			if err != nil {
				return "", fmt.Errorf("can't read remote file: %w", err)
			}

			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return "", fmt.Errorf("can't copy remote file to local: %w", err)
			}

			if err := dstFile.Sync(); err != nil {
				return "", fmt.Errorf("can't sync local file: %w", err)
			}

			if err := dstFile.Close(); err != nil {
				return "", fmt.Errorf("can't close local file: %w", err)
			}

			return dstPath, nil
		}
	}

	return "", nil
}

func (s *sshClient) compareVer(packetVer, targetVer string) (bool, error) {
	if targetVer == "" {
		return true, nil
	}

	prefix, version, err := s.versionPrefix(targetVer)
	if err != nil {
		return false, err
	}

	pVer, err := strconv.ParseFloat(packetVer, 32)
	if err != nil {
		return false, err
	}

	tVer, err := strconv.ParseFloat(version, 32)
	if err != nil {
		return false, err
	}

	switch prefix {
	case HigherOrEqualPrefix:
		return pVer >= tVer, nil
	case BelowOrEqualPrefix:
		return pVer <= tVer, nil
	}

	return false, nil
}

const (
	HigherOrEqualPrefix = ">="
	BelowOrEqualPrefix  = "<="
)

func (s *sshClient) versionPrefix(v string) (string, string, error) {
	_, err := strconv.ParseFloat(v, 32)
	if err == nil {
		return "", v, nil
	}

	switch {
	case strings.Contains(v, HigherOrEqualPrefix):
		return HigherOrEqualPrefix, strings.TrimPrefix(v, HigherOrEqualPrefix), nil
	case strings.Contains(v, BelowOrEqualPrefix):
		return BelowOrEqualPrefix, strings.TrimPrefix(v, BelowOrEqualPrefix), nil
	default:
		return "", "", errors.ErrInvalidVerPrefix
	}
}
