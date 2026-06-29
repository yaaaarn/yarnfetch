package main

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"slices"
	"strconv"
	"strings"

	_ "embed"

	osr "github.com/dominodatalab/os-release"
	"github.com/fatih/color"
	"github.com/mackerelio/go-osstat/uptime"
	"github.com/shirou/gopsutil/v4/disk"
)

//go:embed ascii.txt
var ascii string

const packagesCommand = `
	has() {
  	command -v "$1" >/dev/null 2>&1
	}

	has pacman-key && pacman -Qq || true
	has dpkg && dpkg-query -f '.\n' -W || true
	has rpm && rpm -qa || true
	has apk && apk info || true
	has brew && printf '%s\n' "$(brew --cellar)/"* || true

	has nix-store && {
  	nix-store -q -R /run/current-system/sw || true
  	nix-store -q -R ~/.nix-profile || true
	} || true	
`

type FetchItem int

const (
	User FetchItem = iota
	Hostname
	OS
	Kernel
	Uptime
	WM
	Packages
	Host
	Storage
)

func get(item FetchItem) string {
	var value string

	switch item {
	case User:
		user, err := user.Current()
		if err == nil {
			value = user.Username
		}
	case Hostname:
		hostname, err := os.Hostname()
		if err == nil {
			value = hostname
		}
	case OS:
		if runtime.GOOS == "darwin" {
			out, err := exec.Command("sw_vers", "-productVersion").Output()
			if err == nil {
				value = "macOS " + strings.TrimSpace(string(out))
			} else {
				value = "macOS"
			}
		} else {
			contents, err := os.ReadFile("/etc/os-release")
			if err == nil {
				release := osr.Parse(string(contents))
				value = release.PrettyName
			}
		}
	case Kernel:
		if runtime.GOOS == "darwin" {
			out, err := exec.Command("sysctl", "-n", "kern.osrelease").Output()
			if err == nil {
				value = strings.TrimSpace(string(out))
			}
		} else {
			b, err := os.ReadFile("/proc/sys/kernel/osrelease")
			if err == nil {
				value = strings.TrimSpace(string(b))
			}
		}
	case Packages:
		out, err := exec.Command("sh", "-c", packagesCommand).Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			count := 0
			for _, line := range lines {
				if line != "" {
					count++
				}
			}
			if count > 0 {
				value = strconv.Itoa(count)
			}
		}
	case Uptime:
		up, err := uptime.Get()
		if err == nil {
			value = up.String()
		}
	case WM:
		if runtime.GOOS == "darwin" {
			out, err := exec.Command("defaults", "read", "/System/Library/PrivateFrameworks/SkyLight.framework/Resources/Info.plist", "CFBundleShortVersionString").Output()
			if err == nil {
				value = "Quartz Compositor " + strings.TrimSpace(string(out))
			}
		} else {
			variables := [3]string{
				"XDG_CURRENT_DESKTOP",
				"XDG_SESSION_DESKTOP",
				"DESKTOP_SESSION",
			}
			for _, variable := range variables {
				envValue, ok := os.LookupEnv(variable)
				if ok {
					value = envValue
					break
				}
			}
		}
	case Host:
		if runtime.GOOS == "darwin" {
			out, err := exec.Command("sysctl", "-n", "hw.model").Output()
			if err == nil {
				value = strings.TrimSpace(string(out))
			}
		} else {
			paths := []string{
				"/sys/devices/virtual/dmi/id/product_name",
				"/sys/class/dmi/id/product_name",
				"/proc/device-tree/model",
			}
			for _, p := range paths {
				b, err := os.ReadFile(p)
				if err == nil {
					s := strings.TrimSpace(string(b))
					if s != "" {
						value = s
						break
					}
				}
			}
		}
	case Storage:
		usage, err := disk.Usage("/")
		if err == nil {
			value = fmt.Sprintf("%.0f%% full", usage.UsedPercent)
		}
	}

	if value == "" {
		return "unknown"
	}

	return value
}

type listItem struct {
	key   string
	value string
	color color.Attribute
}

func main() {
	items := []listItem{
		{key: "os", value: get(OS), color: color.FgBlue},
		{key: "host", value: get(Host), color: color.FgCyan},
		{key: "kernel", value: get(Kernel), color: color.FgHiMagenta},
		{key: "uptime", value: get(Uptime), color: color.FgGreen},
		{key: "pkgs", value: get(Packages), color: color.FgMagenta},
		{key: "root", value: get(Storage), color: color.FgHiCyan},
		{key: "wm", value: get(WM), color: color.FgYellow},
	}

	bgColors := []([]color.Attribute){
		{color.BgBlack, color.BgRed, color.BgGreen, color.BgBlue, color.BgMagenta, color.BgCyan, color.BgWhite},
		{color.BgHiBlack, color.BgHiRed, color.BgHiGreen, color.BgHiBlue, color.BgHiMagenta, color.BgHiCyan, color.BgHiWhite},
	}

	maxKeyLength := len(slices.MaxFunc(items, func(a listItem, b listItem) int {
		return cmp.Compare(len(a.key), len(b.key))
	}).key)

	ascii = strings.ReplaceAll(ascii, "\r\n", "\n")
	asciiLines := strings.Split(ascii, "\n")

	if len(asciiLines) > 0 && strings.TrimSpace(asciiLines[len(asciiLines)-1]) == "" {
		asciiLines = asciiLines[:len(asciiLines)-1]
	}

	asciiMaxLength := 0
	for _, line := range asciiLines {
		runeCount := len([]rune(line))
		if runeCount > asciiMaxLength {
			asciiMaxLength = runeCount
		}
	}

	var infoLines []string
	cyanBold := color.New(color.Bold, color.FgCyan).SprintFunc()
	infoLines = append(infoLines, cyanBold(get(User)+"@"+get(Hostname)))

	for _, item := range items {
		c := color.New(item.color).SprintFunc()
		formattedLine := fmt.Sprintf("%s %s", c(fmt.Sprintf("%-*s", maxKeyLength, item.key)), item.value)
		infoLines = append(infoLines, formattedLine)
	}

	maxTotalLines := max(len(asciiLines), len(infoLines))	

	gap := "   "

	fmt.Println()
	
	for i := range maxTotalLines {
		asciiPart := ""
		infoPart := ""

		if i < len(asciiLines) {
			line := asciiLines[i]
			padding := asciiMaxLength - len([]rune(line))
			asciiPart = line + strings.Repeat(" ", padding)
		} else {
			asciiPart = strings.Repeat(" ", asciiMaxLength)
		}

		if i < len(infoLines) {
			infoPart = infoLines[i]
		}

		fmt.Println(strings.ReplaceAll(asciiPart, "%%", color.GreenString("@_")) + gap + infoPart)
	}

	for _, row := range bgColors {
		fmt.Print("\n", gap+strings.Repeat(" ", asciiMaxLength))
		for _, bgColor := range row {
			x := color.New(bgColor)
			fmt.Print(x.Sprint("   "))
		}
	}

	fmt.Println()
}
