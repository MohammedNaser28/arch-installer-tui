package config

type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDiskSelect
	ScreenHostname
	ScreenUser
	ScreenConfirm
	ScreenInstall
	ScreenDone
	ScreenError
)

type Config struct {
	// Disk
	TargetDisk    string // /dev/sda
	BootMode      string // uefi | bios
	EFIPartition  string // /dev/sda1
	RootPartition string // /dev/sda2

	// System
	Hostname string

	// User
	Username     string
	UserPassword string
	RootPassword string

	// Runtime
	DryRun  bool
	LogPath string
	LastErr string
}

func New() *Config {
	return &Config{
		LogPath: "/tmp/arch-installer.log",
		DryRun:  false,
	}
}
