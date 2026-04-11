package config

type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenNetwork
	ScreenDiskSelect
	ScreenPackages
	ScreenHostname
	ScreenUser
	ScreenConfirm
	ScreenInstall
	ScreenDone
	ScreenError
)

type Config struct {
	TargetDisk    string
	BootMode      string
	EFIPartition  string
	RootPartition string
	Hostname      string
	Username      string
	UserPassword  string
	RootPassword  string
	ExtraPackages []string
	JetbrainsIDEs []string
	DryRun        bool
	LogPath       string
	LastErr       string
}

func New() *Config {
	return &Config{
		LogPath: "/tmp/arch-installer.log",
		DryRun:  false,
	}
}
