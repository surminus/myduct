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
	"delta":      "0.18.2",
	"obsidian":   "1.8.9",
	"slack":      "4.43.51",
	"tidal-hifi": "5.19.0",
	"zoxide":     "0.9.7",
}

var dotFiles = []string{
	"default-go-packages",
	"gemrc",
	"gitconfig",
	"ripgreprc",
	"zshrc",
}

var packages = []string{
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
	"libreoffice-calc",
	"libsqlite3-dev",
	"libssl-dev",
	"libterm-readkey-perl",
	"libyaml-dev",
	"ncdu",
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
	"xclip",
	"xkcdpass",
	"zlib1g-dev",
}

// packages only installed for home installs
var homePackages = []string{
	// Allows configuring the install for use with music production software
	"ubuntustudio-installer",
}

var r = viaduct.New()

func main() {
	if viaduct.Attribute.User.Username != "root" {
		log.Fatal("Must run as root")
	}

	viaduct.Attribute.SetUser("laura")

	if isHomeInstall() {
		viaduct.Log("Detected home install!")
	}

	r.Add(&resources.Directory{Path: filepath.Join(viaduct.Attribute.User.HomeDir, "bin")})
	r.Add(&resources.Directory{Path: filepath.Join(viaduct.Attribute.User.HomeDir, "tmp")})

	// Core
	zsh()
	dotfiles()
	tools()
	user()

	// Other
	deleteSnap()
	docker()
	github()
	librewolf()
	mise()
	neovim()
	nodejs()
	obsidian()
	slack()
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

	r.Add(resources.CreateLink("~/.default-golang-pkgs", "~/.default-go-packages"))

	// Neovim configuration
	r.Add(&resources.Link{Path: "~/.config/nvim", Source: "~/.dotfiles/nvim"}, repo)

	// zsh-theme
	r.Add(&resources.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}, repo)

	// librewolf
	r.Add(resources.CreateLink("~/.config/librewolf/librewolf.overrides.cfg", "~/.dotfiles/librewolf.overrides.cfg"), repo, r.Add(resources.Dir("~/.config/librewolf")))

	// Mise
	r.Add(&resources.Link{Path: "~/.config/mise", Source: "~/.dotfiles/mise"}, repo)

	// Install kitty config
	kittyCfgDir := r.Add(resources.Dir("~/.config/kitty"))
	r.Add(&resources.Link{Path: "~/.config/kitty", Source: "~/.dotfiles/kitty"}, repo, kittyCfgDir)
	r.Add(resources.CreateFile("/usr/share/applications/kitty.desktop", resources.EmbeddedFile(files, "files/kitty.desktop")))

	// Configure fonts
	r.Add(resources.CreateLink("~/.local/share/fonts", "~/.dotfiles/fonts"), repo)
	if isKDE() {
		r.Add(resources.CreateFile("/etc/fonts/conf.avail/56-kubuntu-noto.conf", resources.EmbeddedFile(files, "files/56-kubuntu-noto.conf")))
	}
}

func tools() {
	r.Add(&resources.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git", Reference: "refs/heads/master"})

	var deps []*viaduct.Resource

	deps = append(deps, r.Add(&resources.Apt{
		Name:       "git",
		URI:        "https://ppa.launchpadcontent.net/git-core/ppa/ubuntu",
		SigningKey: "A1715D88E1DF1F24",
		Update:     true,
	}))

	pkgs := packages
	if isHomeInstall() {
		pkgs = append(pkgs, homePackages...)
	}

	r.Add(resources.Pkgs(pkgs...), deps...)

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
	installDebPkg("slack", v, fmt.Sprintf("https://downloads.slack-edge.com/desktop-releases/linux/x64/%s/slack-desktop-%s-amd64.deb", v, v))
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
	dep := r.Add(&resources.Apt{
		Distribution: "librewolf",
		Name:         "librewolf",
		Parameters:   map[string]string{"arch": "amd64"},
		PublicPgpKey: resources.EmbeddedFile(files, "files/librewolf.asc"),
		URI:          "https://repo.librewolf.net",
		Update:       true,
		Format:       resources.Sources,
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

// For a home install, simply touch ~/.myducthome to install additional
// packages
func isHomeInstall() bool {
	return viaduct.FileExists(viaduct.ExpandPath("~/.myducthome"))
}

func github() {
	r.Add(resources.Pkg("gh"),
		r.Add(&resources.Apt{
			Distribution:  "stable",
			Name:          "github",
			Parameters:    map[string]string{"arch": "amd64"},
			SigningKeyURL: "https://cli.github.com/packages/githubcli-archive-keyring.gpg",
			URI:           "https://cli.github.com/packages",
			Update:        true,
		}),
	)
}

func neovim() {
	// Install neovim manually since the upstream debian package broke my
	// install
	r.Add(&resources.Package{Names: []string{"neovim"}, Uninstall: true})

	tmp := viaduct.TmpFile("nvim-linux-x86_64.tar.gz")
	dl := r.Add(&resources.Download{URL: "https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.tar.gz", Path: tmp})
	rmdir := r.Add(&resources.Directory{Path: "/usr/share/nvim", Delete: true})
	unpack := r.Add(resources.Exec(fmt.Sprintf("tar -C /usr/share -xzf %s", tmp)), dl, rmdir)
	r.Add(resources.CreateLink("/usr/share/nvim", "/usr/share/nvim-linux-x86_64"), unpack)
}

func mise() {
	dep := r.Add(&resources.Apt{
		Distribution:  "stable",
		Name:          "mise",
		Parameters:    map[string]string{"arch": "amd64"},
		SigningKeyURL: "https://mise.jdx.dev/gpg-key.pub",
		URI:           "https://mise.jdx.dev/deb ",
		Update:        true,
	})

	r.Add(resources.Pkg("mise"), dep)
}

func obsidian() {
	v := packageVersions["obsidian"]
	installDebPkg("obsidian", v, fmt.Sprintf("https://github.com/obsidianmd/obsidian-releases/releases/download/v%s/obsidian_%s_amd64.deb", v, v))
	r.Add(resources.Repo("~/surminus/notes", "git@github.com:surminus/notes.git"))
}
