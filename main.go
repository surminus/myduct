package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/surminus/viaduct"
	"github.com/surminus/viaduct/resources"
)

//go:embed files
var files embed.FS

var packageVersions = map[string]string{
	"delta":      "0.17.0",
	"slack":      "4.33.84",
	"tidal-hifi": "5.16.0",
	"zoxide":     "0.9.4",
}

var dotFiles = []string{
	"gemrc",
	"gitconfig",
	"ripgreprc",
	"tool-versions",
	"zshrc",
}

var ubuntuPackages = []string{
	"apt-transport-https",
	"bat",
	"blueman",
	"ca-certificates",
	"colordiff",
	"curl",
	"exuberant-ctags",
	"fd-find",
	"flameshot",
	"fonts-noto-color-emoji",
	"git",
	"htop",
	"hub",
	"ipcalc",
	"jq",
	"kitty",
	"libbz2-dev",
	"libffi-dev",
	"libsqlite3-dev",
	"libssl-dev",
	"libterm-readkey-perl",
	"libyaml-dev",
	"ncdu",
	"neovim",
	"network-manager-openvpn-gnome",
	"openvpn",
	"pass",
	"pwgen",
	"resolvconf",
	"ripgrep",
	"sd",
	"shellcheck",
	"software-properties-common",
	"tldr",
	"tmux",
	"vim",
	"vim-gui-common",
	"vim-nox",
	"virtualbox",
	"xclip",
	"xkcdpass",
	"zlib1g-dev",
}

var r = viaduct.New()

func main() {
	if viaduct.Attribute.User.Username != "root" {
		log.Fatal("Must run as root")
	}

	viaduct.Attribute.SetUser("laura")

	r.Add(&resources.Directory{Path: filepath.Join(viaduct.Attribute.User.HomeDir, "bin")})
	r.Add(&resources.Directory{Path: filepath.Join(viaduct.Attribute.User.HomeDir, "tmp")})

	zsh()
	dotfiles()
	tools()
	asdf()
	docker()
	slack()
	nodejs()
	user()
	librewolf()
	deleteSnap()
	tidal()

	r.Run()
}

func zsh() {
	r.Add(resources.Pkg("zsh"))
	zsh := r.Add(&resources.Git{Path: "~/.oh-my-zsh", URL: "https://github.com/ohmyzsh/ohmyzsh.git", Reference: "refs/heads/master"})
	r.Add(&resources.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-autosuggestions", URL: "https://github.com/zsh-users/zsh-autosuggestions", Reference: "refs/heads/master"}, zsh)
	r.Add(&resources.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-completions", URL: "https://github.com/zsh-users/zsh-completions", Reference: "refs/heads/master"}, zsh)
}

func dotfiles() {
	repo := r.Add(resources.Repo(
		"~/.dotfiles",
		"git@github.com:surminus/dotfiles.git",
	))

	for _, file := range dotFiles {
		r.Add(&resources.Link{
			Path:   "~/." + file,
			Source: filepath.Join("~/.dotfiles", file),
		}, repo)
	}

	// Neovim configuration
	r.Add(&resources.Link{Path: "~/.config/nvim", Source: "~/.dotfiles/nvim"}, repo)

	// zsh-theme
	r.Add(&resources.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}, repo)

	// librewolf
	r.Add(resources.CreateLink("~/.config/librewolf/librewolf.overrides.cfg", "~/.dotfiles/librewolf.overrides.cfg"), repo, r.Add(resources.Dir("~/.config/librewolf")))

	// Install kitty config
	kittyCfgDir := r.Add(resources.Dir("~/.config/kitty"))
	r.Add(&resources.Link{Path: "~/.config/kitty", Source: "~/.dotfiles/kitty"}, repo, kittyCfgDir)
	r.Add(resources.CreateFile("/usr/share/applications/kitty.desktop", resources.EmbeddedFile(files, "files/kitty.desktop")), kittyCfgDir) // Default to start in fullscreen mode

	// Configure fonts
	r.Add(resources.CreateLink("~/.local/share/fonts", "~/.dotfiles/fonts"), repo)
	if isKDE() {
		r.Add(resources.CreateFile("/etc/fonts/conf.avail/56-kubuntu-noto.conf", resources.EmbeddedFile(files, "files/56-kubuntu-noto.conf")))
	}
}

func tools() {
	r.Add(&resources.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git", Reference: "refs/heads/master"})

	var deps []*viaduct.Resource

	// The PPAs for the latest version have not been built yet
	if viaduct.Attribute.Platform.VersionID != "24.04" {
		deps = append(deps, r.Add(&resources.Apt{
			Name:         "vim",
			URI:          "https://ppa.launchpadcontent.net/jonathonf/vim/ubuntu",
			Distribution: ubuntuDistribution(),
			SigningKey:   "8CF63AD3F06FC659",
			Update:       true,
		}))

		deps = append(deps, r.Add(&resources.Apt{
			Name:       "git",
			URI:        "https://ppa.launchpadcontent.net/git-core/ppa/ubuntu",
			SigningKey: "A1715D88E1DF1F24",
			Update:     true,
		}))
	}

	r.Add(resources.Pkgs(ubuntuPackages...), deps...)

	// Install delta
	v := packageVersions["delta"]
	installDebPkg("git-delta", v, fmt.Sprintf("https://github.com/dandavison/delta/releases/download/%s/git-delta_%s_amd64.deb", v, v))

	// Install zoxide
	v = packageVersions["zoxide"]
	installDebPkg("zoxide", v, fmt.Sprintf("https://github.com/ajeetdsouza/zoxide/releases/download/v%s/zoxide_%s-1_amd64.deb", v, v))

	toolkit := r.Add(&resources.Git{Path: "~/surminus/toolkit", URL: "git@github.com:surminus/toolkit", Reference: "refs/heads/main"})
	for _, file := range []string{"awsexport", "discord-updater", "goinstall"} {
		r.Add(&resources.Link{
			Path:   "~/bin/" + file,
			Source: filepath.Join("~/surminus/toolkit", file),
		}, toolkit)
	}
}

func slack() {
	// Don't bother installing on WSL
	if viaduct.Attribute.Hostname == "win-hub" {
		return
	}

	v := packageVersions["slack"]
	installDebPkg("slack", v, fmt.Sprintf("https://downloads.slack-edge.com/releases/linux/%s/prod/x64/slack-desktop-%s-amd64.deb", v, v))
}

func asdf() {
	repo := r.Add(&resources.Git{
		Path:      "~/.asdf",
		URL:       "https://github.com/asdf-vm/asdf",
		Reference: "refs/tags/v0.10.2",
	})

	dir := r.Add(&resources.Directory{Path: "~/.asdf/plugins"}, repo)

	// refs/heads/master
	for plugin, url := range map[string]string{
		"golang": "kennyp/asdf-golang",
		"goss": "raimon49/asdf-goss",
		"jq": "azmcode/asdf-jq",
		"nodejs": "asdf-vm/asdf-nodejs",
		"python": "danhper/asdf-python",
		"ruby":   "asdf-vm/asdf-ruby",
		"rust": "asdf-community/asdf-rust",
	} {
		r.Add(&resources.Git{
			Path:      fmt.Sprintf("~/.asdf/plugins/%s", plugin),
			URL:       fmt.Sprintf("https://github.com/%s", url),
			Reference: "refs/heads/master",
			Ensure:    true,
		}, dir)
	}

	// refs/heads/main
	for plugin, url := range map[string]string{
		"awscli": "MetricMike/asdf-awscli",
		"opentofu": "virtualroot/asdf-opentofu",
	} {
		r.Add(&resources.Git{
			Path:      fmt.Sprintf("~/.asdf/plugins/%s", plugin),
			URL:       fmt.Sprintf("https://github.com/%s", url),
			Reference: "refs/heads/main",
			Ensure:    true,
		}, dir)
	}
}

func docker() {
	apt := r.Add(&resources.Apt{
		Name:          "docker",
		URI:           "https://download.docker.com/linux/ubuntu",
		Parameters:    map[string]string{"arch": viaduct.Attribute.Arch},
		Source:        "stable",
		Distribution:  ubuntuDistribution(),
		SigningKeyURL: "https://download.docker.com/linux/ubuntu/gpg",
	})

	install := r.Add(resources.Pkg("docker-ce"), r.Add(resources.AptUpdate(), apt))

	// We need to add a User resource here to manage users, so we can
	// add the docker group to the user
	r.Add(&resources.Execute{
		Command: fmt.Sprintf("usermod -aG docker %s", viaduct.Attribute.User.Username),
		Unless:  fmt.Sprintf("grep %s /etc/group | grep -q docker", viaduct.Attribute.User.Username),
	}, install)
}

func nodejs() {
	if viaduct.Attribute.Platform.VersionID != "24.04" {
		r.Add(&resources.Apt{
			Name:          "node",
			URI:           "https://deb.nodesource.com/node_18.x",
			SigningKeyURL: "https://deb.nodesource.com/gpgkey/nodesource.gpg.key",
			Distribution:  ubuntuDistribution(),
			Update:        true,
		})
	}

	r.Add(resources.Pkg("nodejs"))
}

func user() {
	r.Add(resources.ExecUnless("usermod -s /bin/zsh laura", "grep laura /etc/passwd | grep -q zsh"))

	r.Add(resources.DeleteFile("~/.face"))
	r.Add(resources.DeleteFile("/var/lib/AccountsService/icons/laura"))
}

func ubuntuDistribution() string {
	distribution := viaduct.Attribute.Platform.UbuntuCodename
	if distribution == "mantic" || distribution == "lunar" {
		distribution = "jammy"
	}

	return distribution
}

// Snap is a fucking pain in the ass
func deleteSnap() {
	deleteSnap := r.Add(&resources.Package{Names: []string{"snapd"}, Uninstall: true})
	holdSnap := r.Add(&resources.Execute{Command: "apt-mark hold snapd", Unless: "apt-mark showhold | grep -q snapd"}, deleteSnap)
	r.Add(resources.CreateFile("/etc/apt/preferences.d/nosnap.pref", resources.EmbeddedFile(files, "files/nosnap.pref")), deleteSnap, holdSnap)
}

func librewolf() {
	// https://librewolf.net/installation/debian/
	distro := func() string {
		if viaduct.Attribute.Platform.UbuntuCodename == "noble" || viaduct.Attribute.Platform.UbuntuCodename == "jammy" {
			return "jammy"
		}

		return "focal"
	}

	dep := r.Add(&resources.Apt{
		Distribution:  distro(),
		Name:          "librewolf",
		Parameters:    map[string]string{"arch": "amd64"},
		SigningKeyURL: "https://deb.librewolf.net/keyring.gpg",
		URI:           "https://deb.librewolf.net",
		Update:        true,
	})

	r.Add(resources.Pkg("librewolf"), dep)
}

func tidal() {
	v := packageVersions["tidal-hifi"]
	installDebPkg("tidal-hifi", v, fmt.Sprintf("https://github.com/Mastermindzh/tidal-hifi/releases/download/%s/tidal-hifi_%s_amd64.deb", v, v))
}

func installDebPkg(name, version, source string) {
	currentVersion := viaduct.CommandOutput(fmt.Sprintf("dpkg -l | awk '/%s/ {print $3}'", name))

	if !strings.HasPrefix(currentVersion, version) {
		viaduct.Log(name, " =>", currentVersion)
		pkg := viaduct.TmpFile(fmt.Sprintf("%s.deb", name))
		deb := r.Add(resources.Wget(source, pkg))
		r.WithLock(r.Add(resources.Exec("sudo dpkg -i "+pkg), deb))
	} else {
		viaduct.Log(name, " up to date")
	}
}

func isKDE() bool {
	return os.Getenv("XDG_CURRENT_DESKTOP") == "KDE"
}
