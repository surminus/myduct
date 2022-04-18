package main

import (
	"path/filepath"

	v "github.com/surminus/viaduct"
)

var archPackages = []string{
	"git-delta",
	"noto-fonts-emoji", // https://chrpaul.de/2019/07/Enable-colour-emoji-support-on-Manjaro-Linux.html we should add this config here
}

var ubuntuPackages = []string{}

func main() {
	v.Directory{Path: filepath.Join(v.Attribute.User.HomeDir, "bin")}.Create()

	zsh()
	vim()
	dotfiles()
	runtimeEnvs()
	tools()
}

func zsh() {
	v.Package{Name: "zsh", Sudo: true}.Install()
	v.Git{Path: "~/.oh-my-zsh", URL: "https://github.com/ohmyzsh/ohmyzsh.git"}.Create()
	v.Git{Path: "~/.oh-my-zsh/custom/plugins/zsh-autosuggestions", URL: "https://github.com/zsh-users/zsh-autosuggestions"}.Create()
}

func vim() {
	v.Git{
		Path: filepath.Join(v.Attribute.User.HomeDir, ".cache", "dein", "repos", "github.com", "Shougo", "dein.vim"),
		URL:  "https://github.com/Shougo/dein.vim",
	}.Create()

	var pkgs []string
	switch v.Attribute.Platform.ID {
	case "manjaro":
		pkgs = []string{"python", "python-pip"}
	default:
		pkgs = []string{"python3", "python3-pip"}
	}

	v.Packages{Names: pkgs, Sudo: true}.Install()
	v.Execute{Command: "pip install --user pynvim"}.Run()

	// Allow recursive creates
	v.Directory{Path: "~/.vim/swapfiles"}.Create()
}

func dotfiles() {
	v.Git{
		Path: "~/.dotfiles",
		URL:  "git@github.com:surminus/dotfiles.git",
	}.Create()

	files := []string{
		"colordiffrc",
		"gitconfig",
		"terraformrc",
		"tmux.conf",
		"vimrc",
		"zshrc",
	}

	for _, file := range files {
		// I opted against forcibly removing files, but I should JFDI
		// v.File{Path: "~/." + file}.Delete()

		v.Link{
			Path:   "~/." + file,
			Source: filepath.Join(v.Attribute.User.HomeDir, ".dotfiles", file), // This should also expand tildes
		}.Create()
	}

	v.Link{Path: "~/.oh-my-zsh/custom/themes/surminus.zsh-theme", Source: "~/.dotfiles/surminus.zsh-theme"}.Create()

	if v.Attribute.Platform.ID == "manjaro" {
		v.Directory{Path: "~/.config/terminator"}.Create()
		v.Link{Path: "~/.config/terminator/config", Source: "~/.dotfiles/terminator.manjaro"}.Create()
	}
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
			Path: path,
			URL:  url,
		}.Create()
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

	v.Packages{Names: pkgs, Sudo: true}.Install()
}
