package main

import (
	"fmt"
	"path/filepath"

	v "github.com/surminus/viaduct"
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
	"git-delta",
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

func main() {
	v.Directory{Path: filepath.Join(v.Attribute.User.HomeDir, "bin")}.Create()

	myduct()
	aptUpdate()
	zsh()
	vim()
	dotfiles()
	runtimeEnvs()
	tools()
	tmux()
	asdf()
	docker()
}

func aptUpdate() {
	v.Execute{Command: "apt-get update", Sudo: true, Quiet: true}.Run()
}

func zsh() {
	v.Package{Name: "zsh", Sudo: true}.Install()
	v.Git{Path: "~/.oh-my-zsh", URL: "https://github.com/ohmyzsh/ohmyzsh.git"}.Create()
	v.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-autosuggestions", URL: "https://github.com/zsh-users/zsh-autosuggestions"}.Create()
}

func vim() {
	v.Directory{Path: "~/.vim/swapfiles"}.Create()
}

func dotfiles() {
	v.Git{
		Path:   "~/.dotfiles",
		URL:    "git@github.com:surminus/dotfiles.git",
		Ensure: true,
	}.Create()

	files := []string{
		"colordiffrc",
		"gitconfig",
		"terraformrc",
		"tmux.conf",
		"tool-versions",
		"vimrc",
		"zshrc",
	}

	for _, file := range files {
		// I opted against forcibly removing files, but I should JFDI
		// v.File{Path: "~/." + file}.Delete()

		v.Link{
			Path:   "~/." + file,
			Source: filepath.Join("~/.dotfiles", file),
		}.Create()
	}

	v.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}.Create()

	// Add terminator configuration
	v.Directory{Path: "~/.config/terminator"}.Create()
	if v.Attribute.Platform.ID == "manjaro" {
		v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.manjaro"}.Create()
	}

	if v.Attribute.Platform.IDLike == "ubuntu" {
		if v.Attribute.Hostname == "laura-hub" {
			v.Directory{Path: "~/.config/terminator"}.Create()
			v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.desktop"}.Create()
		} else {
			v.Directory{Path: "~/.config/terminator"}.Create()
			v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.laptop"}.Create()
		}
	}

	// Ensure CoC is set up correctly
	v.Directory{Path: "~/.vim"}.Create()
	v.Link{Path: "~/.vim/coc-settings.json", Source: "~/.dotfiles/coc-settings.json"}.Create()
}

func runtimeEnvs() {
	envs := map[string]string{
		"https://github.com/kamatama41/tfenv.git": "~/.tfenv",
		"https://github.com/pyenv/pyenv.git":      "~/.pyenv",
		"https://github.com/rbenv/rbenv.git":      "~/.rbenv",
		"https://github.com/syndbg/goenv.git":     "~/.goenv",
	}

	for url, path := range envs {
		v.Git{
			Path:      path,
			URL:       url,
			Reference: "refs/heads/master",
			Ensure:    true,
		}.Delete()
	}
}

func tools() {
	v.Git{Path: "~/.fzf", URL: "https://github.com/junegunn/fzf.git"}.Create()

	var pkgs []string
	switch v.Attribute.Platform.ID {
	case "manjaro":
		pkgs = archPackages
	default:
		pkgs = ubuntuPackages
	}

	v.Package{Names: pkgs, Sudo: true}.Install()
}

func tmux() {
	v.Git{
		Path:      "~/.tmux/plugins/tpm",
		URL:       "https://github.com/tmux-plugins/tpm",
		Reference: "refs/heads/master",
		Ensure:    true,
	}.Create()
}

func asdf() {
	v.Git{
		Path:      "~/.asdf",
		URL:       "https://github.com/asdf-vm/asdf",
		Reference: "refs/tags/v0.10.2",
	}.Create()

	v.Directory{Path: "~/.asdf/plugins"}.Create()

	for plugin, url := range map[string]string{
		"golang": "https://github.com/kennyp/asdf-golang",
		"nodejs": "https://github.com/asdf-vm/asdf-nodejs",
		"python": "https://github.com/danhper/asdf-python",
		"ruby":   "https://github.com/asdf-vm/asdf-ruby",
	} {
		v.Git{
			Path:      fmt.Sprintf("~/.asdf/plugins/%s", plugin),
			URL:       url,
			Reference: "refs/heads/master",
			Ensure:    true,
		}.Create()
	}
}

func docker() {
	if v.Attribute.Platform.IDLike == "ubuntu" {
		// We should have an apt resource that allows adding repositories
		// using sudo, because the File resource doesn't support writing
		// files as sudo. Instead, we should run myduct using sudo, and
		// allow setting default attributes; or use `cp` instead of
		// using os.WriteFile.
		v.Execute{
			Command: fmt.Sprintf(
				"echo \"deb [arch=%s] https://download.docker.com/linux/ubuntu %s stable\" | sudo tee /etc/apt/sources.list.d/docker.list",
				v.Attribute.Arch,
				v.Attribute.Platform.UbuntuCodename,
			),
			Unless: "test -f /etc/apt/sources.list.d/docker.list",
		}.Run()

		aptUpdate()

		v.Package{Name: "docker-ce", Sudo: true}.Install()
	}

	// We need to add a User resource here to manage users, so we can
	// add the docker group to the user
}

func myduct() {
	v.Git{
		Path:   "~/.myduct",
		URL:    "https://github.com/surminus/myduct",
		Ensure: true,
	}.Create()

	v.Link{Path: "~/bin/myduct", Source: "~/.myduct/build/myduct"}.Create()
}
