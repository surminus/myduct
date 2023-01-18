package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/surminus/viaduct"
	"github.com/surminus/viaduct/resources"
)

const (
	deltaVersion = "0.13.0"
	slackVersion = "4.27.156"
)

var dotFiles = []string{
	"colordiffrc",
	"gemrc",
	"gitconfig",
	"gitignore_global",
	"ripgreprc",
	"terraformrc",
	"tmux.conf",
	"tool-versions",
	"vale.ini",
	"vimrc",
	"zshrc",
}

var archPackages = []string{
	"bat",
	"ctags",
	"flameshot",
	"fzf",
	"git-delta",
	"github-cli",
	"nodejs",
	"noto-fonts-emoji", // https://chrpaul.de/2019/07/Enable-colour-emoji-support-on-Manjaro-Linux.html we should add this config here
	"perl-term-readkey",
	"seahorse",
	"tmux",
	"yarn",
}

var ubuntuPackages = []string{
	"apt-transport-https",
	"awscli",
	"bat",
	"ca-certificates",
	"chromium-browser",
	"colordiff",
	"curl",
	"exuberant-ctags",
	"fd-find",
	"flameshot",
	"git",
	"htop",
	"hub",
	"ipcalc",
	"jq",
	"libbz2-dev",
	"libssl-dev",
	"libterm-readkey-perl",
	"ncdu",
	"network-manager-openvpn-gnome",
	"openvpn",
	"pass",
	"pwgen",
	"resolvconf",
	"ripgrep",
	"shellcheck",
	"software-properties-common",
	"terminator",
	"tldr",
	"tmux",
	"vagrant",
	"vim",
	"vim-gtk",
	"vim-nox",
	"virtualbox",
	"xclip",
	"xkcdpass",
}

var r = viaduct.New()

func main() {
	if viaduct.Attribute.User.Username != "root" {
		log.Fatal("Must run as root")
	}

	viaduct.Attribute.SetUser("laura")

	r.Add(&resources.Directory{Path: filepath.Join(viaduct.Attribute.User.HomeDir, "bin")})

	if viaduct.Attribute.Platform.IDLike == "arch" {
		r.WithLock(r.Add(resources.Exec("sudo pacman -Syy --needed")))
	}

	zsh()
	vim()
	dotfiles()
	tools()
	tmux()
	asdf()
	docker()
	slack()
	nodejs()

	r.Run()
}

func zsh() {
	r.Add(resources.Pkg("zsh"))
	zsh := r.Add(&resources.Git{Path: "~/.oh-my-zsh", URL: "https://github.com/ohmyzsh/ohmyzsh.git", Reference: "refs/heads/master"})
	r.Add(&resources.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-autosuggestions", URL: "https://github.com/zsh-users/zsh-autosuggestions", Reference: "refs/heads/master"}, zsh)
}

func vim() {
	r.Add(resources.Dir("~/.vim/swapfiles"))
}

func dotfiles() {
	repo := r.Add(resources.Repo(
		"~/.dotfiles",
		"git@github.com:surminus/dotfiles.git",
	))

	stylespath := "~/.vale/styles"
	valedir := r.Add(resources.Dir(stylespath))
	valeStyles := []string{
		"alex",
	}

	for _, style := range valeStyles {
		r.Add(&resources.Git{
			Path: fmt.Sprintf("%s/%s", stylespath, style),
			URL:  fmt.Sprintf("git@github.com:errata-ai/%s", style),
		}, valedir)
	}

	for _, file := range dotFiles {
		r.Add(&resources.Link{
			Path:   "~/." + file,
			Source: filepath.Join("~/.dotfiles", file),
		}, repo)
	}

	r.Add(&resources.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}, repo)

	// Add terminator configuration
	termdir := r.Add(&resources.Directory{Path: "~/.config/terminator"}, repo)

	if viaduct.Attribute.Platform.ID == "manjaro" {
		r.Add(&resources.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.manjaro"}, repo, termdir)
	}

	if viaduct.IsUbuntu() {
		if viaduct.Attribute.Hostname == "laura-hub" {
			r.Add(&resources.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.desktop"}, repo, termdir)
		} else {
			r.Add(&resources.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.laptop"}, repo, termdir)
		}
	}

	// Ensure CoC is set up correctly
	vim := r.Add(&resources.Directory{Path: "~/.vim"})
	r.Add(&resources.Link{Path: "~/.vim/coc-settings.json", Source: "~/.dotfiles/coc-settings.json"}, repo, vim)
}

func tools() {
	r.Add(&resources.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git"})

	var vim, git *viaduct.Resource
	if viaduct.IsUbuntu() {
		// vim ppa
		vim = r.Add(&resources.Apt{
			Name: "vim",
			URI:  "https://ppa.launchpadcontent.net/jonathonf/vim/ubuntu",
		})

		git = r.Add(&resources.Apt{
			Name: "git",
			URI:  "https://ppa.launchpadcontent.net/git-core/ppa/ubuntu",
		})
	}

	var pkgs []string
	switch viaduct.Attribute.Platform.ID {
	case "manjaro":
		pkgs = archPackages
		r.Add(resources.Pkgs(pkgs...))
	default:
		pkgs = ubuntuPackages
		r.Add(resources.Pkgs(pkgs...), vim, git)
	}

	if viaduct.IsUbuntu() {
		// Install delta
		if viaduct.CommandOutput("dpkg -l | awk '/git-delta/ {print $3}'") != deltaVersion {
			deltaSource := fmt.Sprintf("https://github.com/dandavison/delta/releases/download/%s/git-delta_%s_amd64.deb", deltaVersion, deltaVersion)
			deltaPkg := viaduct.TmpFile("delta.deb")

			delta := r.Add(resources.Wget(deltaSource, viaduct.TmpFile("delta.deb")))
			r.WithLock(r.Add(resources.Exec("sudo dpkg -i "+deltaPkg), delta))
		} else {
			viaduct.Log("Delta up to date")
		}
	}
}

func tmux() {
	r.Add(&resources.Git{
		Path:      "~/.tmux/plugins/tpm",
		URL:       "https://github.com/tmux-plugins/tpm",
		Reference: "refs/heads/master",
		Ensure:    true,
	})
}

func slack() {
	if viaduct.IsUbuntu() {
		currentVersion := viaduct.CommandOutput("dpkg -l | awk '/slack-desktop/ {print $3}'")
		if currentVersion != slackVersion {
			viaduct.Log(currentVersion)
			slackSource := fmt.Sprintf("https://downloads.slack-edge.com/releases/linux/%s/prod/x64/slack-desktop-%s-amd64.deb", slackVersion, slackVersion)
			slackPkg := viaduct.TmpFile("slack.deb")

			slack := r.Add(resources.Wget(slackSource, slackPkg))
			r.WithLock(r.Add(resources.Exec("sudo dpkg -i "+slackPkg), slack))
		} else {
			viaduct.Log("Slack up to date")
		}
	}
}

func asdf() {
	repo := r.Add(&resources.Git{
		Path:      "~/.asdf",
		URL:       "https://github.com/asdf-vm/asdf",
		Reference: "refs/tags/v0.10.2",
	})

	dir := r.Add(&resources.Directory{Path: "~/.asdf/plugins"}, repo)

	for plugin, url := range map[string]string{
		"golang": "https://github.com/kennyp/asdf-golang",
		"nodejs": "https://github.com/asdf-vm/asdf-nodejs",
		"python": "https://github.com/danhper/asdf-python",
		"ruby":   "https://github.com/asdf-vm/asdf-ruby",
	} {
		r.Add(&resources.Git{
			Path:      fmt.Sprintf("~/.asdf/plugins/%s", plugin),
			URL:       url,
			Reference: "refs/heads/master",
			Ensure:    true,
		}, dir)
	}
}

func docker() {
	if viaduct.IsUbuntu() {
		apt := r.Add(&resources.Apt{
			Name:       "docker",
			URI:        "https://download.docker.com/linux/ubuntu",
			Parameters: map[string]string{"arch": viaduct.Attribute.Arch},
			Source:     "stable",
		})

		install := r.Add(resources.Pkg("docker-ce"), r.Add(resources.AptUpdate(), apt))

		// We need to add a User resource here to manage users, so we can
		// add the docker group to the user
		r.Add(&resources.Execute{
			Command: fmt.Sprintf("usermod -G docker %s", viaduct.Attribute.User.Username),
			Unless:  fmt.Sprintf("grep %s /etc/group | grep -q docker", viaduct.Attribute.User.Username),
		}, install)
	}
}

func nodejs() {
	if viaduct.IsUbuntu() {
		key := r.Add(&resources.Execute{
			Command: "curl -s https://deb.nodesource.com/gpgkey/nodesource.gpg.key | gpg --dearmor | sudo tee /usr/share/keyrings/nodesource.gpg >/dev/null",
			Unless:  "dpkg -l | grep -q nodejs",
		})

		apt := r.Add(&resources.Apt{
			Name: "nodesource",
			URI:  "https://deb.nodesource.com/node_18.x",
			Parameters: map[string]string{
				"signed-by": "/usr/share/keyrings/nodesource.gpg",
			},
		}, key)

		update := r.Add(resources.AptUpdate(), apt)
		r.Add(resources.Pkg("nodejs"), update)
	}
}
