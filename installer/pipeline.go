package installer

import (
	"arch-installer/config"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Step is one unit of the install pipeline
type Step struct {
	Name string
	Run  func(cfg *config.Config) error
}

// Pipeline returns the ordered install steps
func Pipeline(cfg *config.Config) []Step {
	return []Step{
		{"Wiping disk", wipe},
		{"Partitioning disk", partition},
		{"Formatting partitions", format},
		{"Mounting partitions", mount},
		{"Installing base system", pacstrap},
		{"Generating fstab", genfstab},
		{"Setting timezone", timezone},
		{"Setting locale", locale},
		{"Setting hostname", hostname},
		{"Setting root password", rootPassword},
		{"Creating user", createUser},
		{"Installing GRUB", grub},
		{"Unmounting", unmount},
	}
}

// run executes a shell command, returns error with stderr
func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, out)
	}
	return nil
}

// chroot runs a command inside /mnt via arch-chroot
func chroot(args ...string) error {
	return run("arch-chroot", append([]string{"/mnt"}, args...)...)
}

func wipe(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	return run("wipefs", "-a", cfg.TargetDisk)
}

func partition(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	disk := cfg.TargetDisk

	if cfg.BootMode == "uefi" {
		if err := run("parted", "-s", disk, "mklabel", "gpt"); err != nil {
			return err
		}
		if err := run("parted", "-s", disk, "mkpart", "ESP", "fat32", "1MiB", "513MiB"); err != nil {
			return err
		}
		if err := run("parted", "-s", disk, "set", "1", "esp", "on"); err != nil {
			return err
		}
		if err := run("parted", "-s", disk, "mkpart", "primary", "ext4", "513MiB", "100%"); err != nil {
			return err
		}
	} else {
		if err := run("parted", "-s", disk, "mklabel", "msdos"); err != nil {
			return err
		}
		if err := run("parted", "-s", disk, "mkpart", "primary", "ext4", "1MiB", "100%"); err != nil {
			return err
		}
		if err := run("parted", "-s", disk, "set", "1", "boot", "on"); err != nil {
			return err
		}
	}

	// Let kernel re-read partition table
	_ = run("partprobe", disk)
	return nil
}

func format(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if cfg.BootMode == "uefi" {
		if err := run("mkfs.fat", "-F32", cfg.EFIPartition); err != nil {
			return err
		}
	}
	return run("mkfs.ext4", "-F", cfg.RootPartition)
}

func mount(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if err := run("mount", cfg.RootPartition, "/mnt"); err != nil {
		return err
	}
	if cfg.BootMode == "uefi" {
		if err := os.MkdirAll("/mnt/boot", 0755); err != nil {
			return err
		}
		return run("mount", cfg.EFIPartition, "/mnt/boot")
	}
	return nil
}

func pacstrap(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	pkgs := []string{
		"base", "linux", "linux-firmware",
		"networkmanager", "grub", "sudo", "vim",
	}
	if cfg.BootMode == "uefi" {
		pkgs = append(pkgs, "efibootmgr", "dosfstools")
	}
	args := append([]string{"/mnt"}, pkgs...)
	return run("pacstrap", args...)
}

func genfstab(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	out, err := exec.Command("genfstab", "-U", "/mnt").Output()
	if err != nil {
		return fmt.Errorf("genfstab failed: %w", err)
	}
	return os.WriteFile("/mnt/etc/fstab", out, 0644)
}

func timezone(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if err := chroot("ln", "-sf", "/usr/share/zoneinfo/UTC", "/etc/localtime"); err != nil {
		return err
	}
	return chroot("hwclock", "--systohc")
}

func locale(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if err := os.WriteFile("/mnt/etc/locale.gen", []byte("en_US.UTF-8 UTF-8\n"), 0644); err != nil {
		return err
	}
	if err := chroot("locale-gen"); err != nil {
		return err
	}
	return os.WriteFile("/mnt/etc/locale.conf", []byte("LANG=en_US.UTF-8\n"), 0644)
}

func hostname(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if err := os.WriteFile("/mnt/etc/hostname", []byte(cfg.Hostname+"\n"), 0644); err != nil {
		return err
	}
	hosts := fmt.Sprintf("127.0.0.1 localhost\n::1 localhost\n127.0.1.1 %s\n", cfg.Hostname)
	return os.WriteFile("/mnt/etc/hosts", []byte(hosts), 0644)
}

func rootPassword(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	cmd := exec.Command("arch-chroot", "/mnt", "chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("root:%s\n", cfg.RootPassword))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chpasswd root failed: %w\n%s", err, out)
	}
	return nil
}

func createUser(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if err := chroot("useradd", "-m", "-G", "wheel,audio,video,storage", "-s", "/bin/bash", cfg.Username); err != nil {
		return err
	}
	cmd := exec.Command("arch-chroot", "/mnt", "chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s\n", cfg.Username, cfg.UserPassword))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chpasswd user failed: %w\n%s", err, out)
	}
	// Enable sudo for wheel group
	sudoers := "%wheel ALL=(ALL:ALL) ALL\n"
	return os.WriteFile("/mnt/etc/sudoers.d/wheel", []byte(sudoers), 0440)
}

func grub(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	if cfg.BootMode == "uefi" {
		if err := chroot("grub-install",
			"--target=x86_64-efi",
			"--efi-directory=/boot",
			"--bootloader-id=GRUB"); err != nil {
			return err
		}
	} else {
		if err := chroot("grub-install",
			"--target=i386-pc",
			cfg.TargetDisk); err != nil {
			return err
		}
	}
	return chroot("grub-mkconfig", "-o", "/boot/grub/grub.cfg")
}

func unmount(cfg *config.Config) error {
	if cfg.DryRun {
		return nil
	}
	_ = run("umount", "-R", "/mnt") // best effort
	return nil
}
