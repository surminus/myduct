package main

import (
	"fmt"
	"log"
	"path/filepath"

	v "github.com/surminus/viaduct"
)

const (
	deltaVersion = "0.13.0"
	slackVersion = "4.27.156"
)

var archPackages = []string{
	"bat",
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

var r = v.New()

func main() {
	if v.Attribute.User.Username != "root" {
		log.Fatal("Must run as root")
	}

	v.Attribute.SetUser("laura")

	r.Create(v.Directory{Path: filepath.Join(v.Attribute.User.HomeDir, "bin")})

	var deps []*v.Resource
	if v.IsUbuntu() {
		deps = append(deps, r.Update(v.Apt{}))
	}

	if v.Attribute.Platform.IDLike == "arch" {
		r.Run(v.E("sudo pacman -Syy --needed"))
	}

	z := zsh(deps)
	vim()
	dotfiles()
	runtimeEnvs()
	t := tools(z)
	tmux()
	asdf()
	d := docker(t)
	s := slack(d)
	nodejs(s)

	r.Start()
}

func zsh(deps []*v.Resource) (p []*v.Resource) {
	p = append(p, r.Create(v.P("zsh"), deps...))
	zsh := r.Create(v.Git{Path: "~/.oh-my-zsh", URL: "https://github.com/ohmyzsh/ohmyzsh.git"})
	r.Create(v.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-autosuggestions", URL: "https://github.com/zsh-users/zsh-autosuggestions"}, zsh)

	return p
}

func vim() {
	r.Create(v.D("~/.vim/swapfiles"))
}

func dotfiles() {
	repo := r.Create(v.Git{
		Path:   "~/.dotfiles",
		URL:    "git@github.com:surminus/dotfiles.git",
		Ensure: true,
	})

	files := []string{
		"colordiffrc",
		"gemrc",
		"gitconfig",
		"gitignore_global",
		"ripgreprc",
		"terraformrc",
		"tmux.conf",
		"tool-versions",
		"vimrc",
		"zshrc",
	}

	for _, file := range files {
		r.Create(v.Link{
			Path:   "~/." + file,
			Source: filepath.Join("~/.dotfiles", file),
		}, repo)
	}

	r.Create(v.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}, repo)

	// Add terminator configuration
	termdir := r.Create(v.Directory{Path: "~/.config/terminator"}, repo)

	if v.Attribute.Platform.ID == "manjaro" {
		r.Create(v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.manjaro"}, repo, termdir)
	}

	if v.IsUbuntu() {
		if v.Attribute.Hostname == "laura-hub" {
			r.Create(v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.desktop"}, repo, termdir)
		} else {
			r.Create(v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.laptop"}, repo, termdir)
		}
	}

	// Ensure CoC is set up correctly
	vim := r.Create(v.Directory{Path: "~/.vim"})
	r.Create(v.Link{Path: "~/.vim/coc-settings.json", Source: "~/.dotfiles/coc-settings.json"}, repo, vim)
}

func runtimeEnvs() {
	envs := map[string]string{
		"https://github.com/kamatama41/tfenv.git": "~/.tfenv",
		"https://github.com/pyenv/pyenv.git":      "~/.pyenv",
		"https://github.com/rbenv/rbenv.git":      "~/.rbenv",
		"https://github.com/syndbg/goenv.git":     "~/.goenv",
	}

	for url, path := range envs {
		r.Delete(v.Git{
			Path:      path,
			URL:       url,
			Reference: "refs/heads/master",
			Ensure:    true,
		})
	}
}

func tools(deps []*v.Resource) (p []*v.Resource) {
	r.Create(v.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git"})

	if v.IsUbuntu() {
		// vim ppa
		vim := r.Create(v.Apt{
			Name: "vim",
			URI:  "https://ppa.launchpadcontent.net/jonathonf/vim/ubuntu",
		})

		git := r.Create(v.Apt{
			Name: "git",
			URI:  "https://ppa.launchpadcontent.net/git-core/ppa/ubuntu",
		})

		deps = append(deps, r.Update(v.Apt{}, append(deps, []*v.Resource{vim, git}...)...))
	}

	var pkgs []string
	switch v.Attribute.Platform.ID {
	case "manjaro":
		pkgs = archPackages
	default:
		pkgs = ubuntuPackages
	}

	p = append(p, r.Create(v.Ps(pkgs...), deps...))

	if v.IsUbuntu() {
		// Install delta
		deltaSource := fmt.Sprintf("https://github.com/dandavison/delta/releases/download/%s/git-delta_%s_amd64.deb", deltaVersion, deltaVersion)
		deltaPkg := filepath.Join(v.Attribute.TmpDir, "delta.deb")

		delta := r.Run(v.Execute{
			Command: fmt.Sprintf("wget -q %s -O %s", deltaSource, deltaPkg),
			Unless:  "dpkg -l | grep -q git-delta",
		})

		p = append(p, r.Run(v.Execute{
			Command: "sudo dpkg -i " + deltaPkg,
			Unless:  "dpkg -l | grep -q git-delta",
		}, append(deps, delta)...))
	}

	return p
}

func tmux() {
	r.Create(v.Git{
		Path:      "~/.tmux/plugins/tpm",
		URL:       "https://github.com/tmux-plugins/tpm",
		Reference: "refs/heads/master",
		Ensure:    true,
	})
}

func slack(deps []*v.Resource) (p []*v.Resource) {
	if v.IsUbuntu() {
		slackSource := fmt.Sprintf("https://downloads.slack-edge.com/releases/linux/%s/prod/x64/slack-desktop-%s-amd64.deb", slackVersion, slackVersion)
		slackPkg := filepath.Join(v.Attribute.TmpDir, "slack.deb")

		slack := r.Run(v.Execute{
			Command: fmt.Sprintf("wget -q %s -O %s", slackSource, slackPkg),
			Unless:  "dpkg -l | grep -q slack-desktop",
		})

		p = append(p, r.Run(v.Execute{
			Command: "sudo dpkg -i " + slackPkg,
			Unless:  "dpkg -l | grep -q slack-desktop",
		}, append(deps, slack)...))
	}

	return p
}

func asdf() {
	repo := r.Create(v.Git{
		Path:      "~/.asdf",
		URL:       "https://github.com/asdf-vm/asdf",
		Reference: "refs/tags/v0.10.2",
	})

	dir := r.Create(v.Directory{Path: "~/.asdf/plugins"}, repo)

	for plugin, url := range map[string]string{
		"golang": "https://github.com/kennyp/asdf-golang",
		"nodejs": "https://github.com/asdf-vm/asdf-nodejs",
		"python": "https://github.com/danhper/asdf-python",
		"ruby":   "https://github.com/asdf-vm/asdf-ruby",
	} {
		r.Create(v.Git{
			Path:      fmt.Sprintf("~/.asdf/plugins/%s", plugin),
			URL:       url,
			Reference: "refs/heads/master",
			Ensure:    true,
		}, dir)
	}
}

func docker(deps []*v.Resource) (p []*v.Resource) {
	if v.IsUbuntu() {
		apt := r.Create(v.Apt{
			Name:       "docker",
			URI:        "https://download.docker.com/linux/ubuntu",
			Parameters: map[string]string{"arch": v.Attribute.Arch},
			Source:     "stable",
		})

		update := r.Update(v.Apt{}, append(deps, apt)...)
		install := r.Create(v.P("docker-ce"), append(deps, []*v.Resource{update, apt}...)...)

		// We need to add a User resource here to manage users, so we can
		// add the docker group to the user
		p = append(p, r.Run(v.Execute{
			Command: fmt.Sprintf("usermod -G docker %s", v.Attribute.User.Username),
			Unless:  fmt.Sprintf("grep %s /etc/group | grep -q docker", v.Attribute.User.Username),
		}, install))
	}

	return p
}

func nodejs(deps []*v.Resource) (p []*v.Resource) {
	if v.IsUbuntu() {
		key := r.Run(v.Execute{
			Command: "curl -s https://deb.nodesource.com/gpgkey/nodesource.gpg.key | gpg --dearmor | sudo tee /usr/share/keyrings/nodesource.gpg >/dev/null",
			Unless:  "dpkg -l | grep -q nodejs",
		})

		apt := r.Create(v.Apt{
			Name: "nodesource",
			URI:  "https://deb.nodesource.com/node_18.x",
			Parameters: map[string]string{
				"signed-by": "/usr/share/keyrings/nodesource.gpg",
			},
		}, key)

		update := r.Update(v.Apt{}, append(deps, apt)...)
		p = append(p, r.Create(v.P("nodejs"), append(deps, update)...))
	}

	return p
}
