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
	"delta":           "0.18.2",
	"kitty":           "0.47.2",
	"obsidian":        "1.8.9",
	"thorium-browser": "138.0.7204.303",
	"tidal-hifi":      "5.19.0",
	"tree-sitter":     "0.26.8",
	"zoxide":          "0.9.7",
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
	"fonts-noto-color-emoji",
	"git",
	"htop",
	"ipcalc",
	"jq",
	"kde-spectacle",
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
	"pinentry-gnome3",
	"pwgen",
	"resolvconf",
	"ripgrep",
	"sd",
	"shellcheck",
	"software-properties-common",
	"tmux",
	"vim",
	"vim-gui-common",
	"vim-nox",
	"xkcdpass",
	"zlib1g-dev",
}

// packages only installed for home installs
var homePackages = []string{
	// Allows configuring the install for use with music production software
	"ubuntustudio-installer",
}

// skills to symlink
var claudeSkills = []string{
	"git",
}

// agents to symlink
var claudeAgents = []string{
	"code-fixer",
	"code-reviewer",
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
	gpg()
	tools()
	user()

	// Other
	braveBrowser()
	thorium()
	deleteSnap()
	docker()
	github()
	kitty()
	mise()
	neovim()
	obsidian()
	tidal()
	treesitter()

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

	// Mise
	r.Add(&resources.Link{Path: "~/.config/mise", Source: "~/.dotfiles/mise"}, repo)

	// Kitty config
	r.Add(&resources.Link{Path: "~/.config/kitty", Source: "~/.dotfiles/kitty"}, repo)

	// Claude Code
	claudeCfgDir := r.Add(resources.Dir("~/.claude"))
	r.Add(&resources.Link{Path: "~/.claude/CLAUDE.md", Source: "~/.dotfiles/claude/CLAUDE.md"}, repo, claudeCfgDir)
	r.Add(&resources.Link{Path: "~/.claude/settings.json", Source: "~/.dotfiles/claude/settings.json"}, repo, claudeCfgDir)
	r.Add(&resources.Link{Path: "~/.claude/statusline-command.sh", Source: "~/.dotfiles/claude/statusline-command.sh"}, repo, claudeCfgDir)

	claudeSkillsDir := r.Add(resources.Dir("~/.claude/skills"))

	for _, skill := range claudeSkills {
		r.Add(&resources.Link{Path: "~/.claude/skills/" + skill, Source: "~/.dotfiles/claude/skills/" + skill}, repo, claudeCfgDir, claudeSkillsDir)
	}

	claudeAgentsDir := r.Add(resources.Dir("~/.claude/agents"))

	for _, agent := range claudeAgents {
		r.Add(&resources.Link{Path: "~/.claude/agents/" + agent + ".md", Source: "~/.dotfiles/claude/agents/" + agent + ".md"}, repo, claudeCfgDir, claudeAgentsDir)
	}

	// A local directory for general discussion using Claude Code
	r.Add(resources.Dir("~/claude"))

	// Configure fonts
	r.Add(resources.CreateLink("~/.local/share/fonts", "~/.dotfiles/fonts"), repo)
	if isKDE() {
		r.Add(resources.CreateFile("/etc/fonts/conf.avail/56-kubuntu-noto.conf", resources.EmbeddedFile(files, "files/56-kubuntu-noto.conf")))
	}
}

func tools() {
	r.Add(&resources.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git", Reference: "refs/heads/master"})

	pkgs := packages
	if isHomeInstall() {
		pkgs = append(pkgs, homePackages...)
	}

	r.Add(resources.Pkgs(pkgs...))

	// Install delta
	v := packageVersions["delta"]
	installDebPkg("git-delta", v, fmt.Sprintf("https://github.com/dandavison/delta/releases/download/%s/git-delta_%s_amd64.deb", v, v))

	// Install zoxide
	v = packageVersions["zoxide"]
	installDebPkg("zoxide", v, fmt.Sprintf("https://github.com/ajeetdsouza/zoxide/releases/download/v%s/zoxide_%s-1_amd64.deb", v, v))

	// toolkit is always on PATH
	r.Add(&resources.Git{Path: "~/surminus/toolkit", URL: "git@github.com:surminus/toolkit", Reference: "refs/heads/main"})

	// Ensure hub is uninstalled
	r.Add(&resources.Package{Names: []string{"hub"}, Uninstall: true})
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

	// Add the user to the docker group so they can talk to the daemon
	r.Add(&resources.User{
		Name:   viaduct.Attribute.User.Username,
		Groups: []string{"docker"},
	}, install)
}

func user() {
	r.Add(resources.ExecUnless("usermod -s /bin/zsh laura", "grep laura /etc/passwd | grep -q zsh"))

	r.Add(resources.DeleteFile("~/.face"))
	r.Add(resources.DeleteFile("/var/lib/AccountsService/icons/laura"))
}

// Cache the GPG passphrase for a day so that pass-backed secrets don't prompt
// for it on every new shell. The stock cache is only 10 minutes.
//
// Use pinentry-gnome3, which (unlike pinentry-qt) integrates with libsecret and
// offers a "Save in password manager" checkbox. Ticking it stores the passphrase
// in KWallet (which provides the Secret Service on this KDE box), so GPG stops
// prompting entirely while logged in with the wallet unlocked.
func gpg() {
	// On a fresh install ~/.gnupg doesn't exist yet, so dropping gpg-agent.conf
	// into it fails. Create it first (0700, or gpg refuses to use it).
	dir := r.Add(&resources.Directory{Path: "~/.gnupg", Permissions: resources.Permissions{Mode: 0o700}})

	r.Add(resources.CreateFile("~/.gnupg/gpg-agent.conf", `default-cache-ttl 86400
max-cache-ttl 86400
pinentry-program /usr/bin/pinentry-gnome3
`), dir)
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

	// AptHold has no built-in idempotency, and re-holding the now-purged
	// snapd package errors, so guard on it not already being held
	hold := resources.AptHold("snapd")
	hold.Unless = "apt-mark showhold | grep -q snapd"
	holdSnap := r.Add(hold, deleteSnap)
	r.Add(resources.CreateFile("/etc/apt/preferences.d/nosnap.pref", resources.EmbeddedFile(files, "files/nosnap.pref")), deleteSnap, holdSnap)

	// Clean up any lingering snap mount units and data left behind after snapd removal
	r.Add(&resources.Execute{
		Command: "find /etc/systemd/system -name 'snap*.mount' -o -name 'snap.*.service' -o -name 'snap.*.timer' | xargs --no-run-if-empty rm -f && find /etc/systemd/system -name 'snapd*' | xargs --no-run-if-empty rm -f && systemctl daemon-reload",
		Unless:  "test ! -d /var/lib/snapd",
	}, deleteSnap)
	r.Add(&resources.Execute{
		Command: "umount /snap/*/* 2>/dev/null; umount /snap/* 2>/dev/null; rm -rf /snap /var/snap /var/lib/snapd",
		Unless:  "test ! -d /var/lib/snapd",
	}, deleteSnap)
}

func braveBrowser() {
	dep := r.Add(&resources.Apt{
		Name:          "brave-browser",
		URI:           "https://brave-browser-apt-release.s3.brave.com",
		SigningKeyURL: "https://brave-browser-apt-release.s3.brave.com/brave-browser-archive-keyring.gpg",
		Distribution:  "stable",
		Update:        true,
		Format:        resources.Sources,
	})

	r.Add(resources.Pkg("brave-browser"), dep)
}

// Thorium is an optimised Chromium fork. I install the AVX2 build to match
// this machine's CPU. Note the Linux releases lag upstream Chromium, so this
// pins to the newest Linux .deb rather than the latest overall tag.
func thorium() {
	v := packageVersions["thorium-browser"]
	installDebPkg("thorium-browser", v, fmt.Sprintf("https://github.com/Alex313031/thorium/releases/download/M%s/thorium-browser_%s_AVX2.deb", v, v))
}

func tidal() {
	v := packageVersions["tidal-hifi"]
	installDebPkg("tidal-hifi", v, fmt.Sprintf("https://github.com/Mastermindzh/tidal-hifi/releases/download/%s/tidal-hifi_%s_amd64.deb", v, v))
}

func installDebPkg(name, version, source string, deps ...*viaduct.Resource) {
	currentVersion := viaduct.CommandOutput(fmt.Sprintf("dpkg -l | awk '/%s/ {print $3}'", name))

	if !strings.HasPrefix(currentVersion, version) {
		viaduct.Log(name, " =>", currentVersion)
		pkg := viaduct.TmpFile(fmt.Sprintf("%s.deb", name))
		deb := r.Add(resources.Wget(source, pkg), deps...)
		r.Add(resources.InstallDeb(pkg), deb)
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

func kitty() {
	// Remove the apt package in favour of a manual install
	r.Add(&resources.Package{Names: []string{"kitty"}, Uninstall: true})

	v := packageVersions["kitty"]
	currentVersion := viaduct.CommandOutput("kitty --version | awk '{print $2}'")

	installDir := fmt.Sprintf("/usr/share/kitty-%s", v)

	if currentVersion != v {
		viaduct.Log("kitty", " =>", currentVersion)
		tmp := viaduct.TmpFile("kitty-x86_64.txz")
		dl := r.Add(&resources.Download{
			URL:  fmt.Sprintf("https://github.com/kovidgoyal/kitty/releases/download/v%s/kitty-%s-x86_64.txz", v, v),
			Path: tmp,
		})
		rmdir := r.Add(&resources.Directory{Path: installDir, Delete: true})
		mkdir := r.Add(resources.Dir(installDir), rmdir)
		unpack := r.Add(resources.Exec(fmt.Sprintf("tar -C %s -xJf %s", installDir, tmp)), dl, mkdir)
		r.Add(resources.CreateLink("/usr/share/kitty", installDir), unpack)
		r.Add(resources.CreateLink("/usr/bin/kitty", "/usr/share/kitty/bin/kitty"), unpack)
		r.Add(resources.CreateLink("/usr/bin/kitten", "/usr/share/kitty/bin/kitten"), unpack)
	} else {
		viaduct.Log("kitty", " up to date")
	}

	r.Add(resources.CreateFile("/usr/share/applications/kitty.desktop", resources.EmbeddedFile(files, "files/kitty.desktop")))
	r.Add(resources.CreateLink("/usr/share/icons/hicolor/scalable/apps/kitty.svg", "/usr/share/kitty/share/icons/hicolor/scalable/apps/kitty.svg"))
}

func neovim() {
	// Install neovim manually since the upstream debian package broke my
	// install
	r.Add(&resources.Package{Names: []string{"neovim"}, Uninstall: true})

	tmp := viaduct.TmpFile("nvim-linux-x86_64.tar.gz")
	r.Chain(
		&resources.Download{URL: "https://github.com/neovim/neovim/releases/latest/download/nvim-linux-x86_64.tar.gz", Path: tmp},
		&resources.Directory{Path: "/usr/share/nvim", Delete: true},
		resources.Extract(tmp, "/usr/share"),
		resources.CreateLink("/usr/share/nvim", "/usr/share/nvim-linux-x86_64"),
	)
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

func treesitter() {
	v := packageVersions["tree-sitter"]
	source := fmt.Sprintf("https://github.com/tree-sitter/tree-sitter/releases/download/v%s/tree-sitter-cli-linux-x64.zip", v)
	tmp := viaduct.TmpFile("tree-sitter.zip")
	binDir := viaduct.ExpandPath("~/bin")
	r.Chain(
		&resources.Download{URL: source, Path: tmp},
		&resources.Archive{Path: tmp, Dest: binDir, Pick: []string{"tree-sitter"}},
		resources.Exec(fmt.Sprintf("chmod +x %s/tree-sitter", binDir)),
	)
}
