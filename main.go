package main

import (
	"path/filepath"

	v "github.com/surminus/viaduct"
)

const home = "/home/laura"

func main() {
	v.Directory{Path: filepath.Join(home, "bin")}.Create()

	vim()
	dotfiles()
}

func vim() {
	v.Git{
		Path: filepath.Join(home, ".cache", "dein", "repos", "github.com", "Shougo", "dein.vim"),
		URL:  "https://github.com/Shougo/dein.vim",
	}.Create()

	// Need to run this as sudo
	// v.Packages{Packages: []string{"python", "python-pip"}}.Install()

	v.Execute{Command: "pip install --user pynvim"}.Run()

	// Allow recursive creates
	vimDir := v.Directory{Path: filepath.Join(home, ".vim")}.Create()
	v.Directory{Path: filepath.Join(vimDir.Path, "swapfiles")}.Create()
}

func dotfiles() {
	v.Git{
		Path: home + "/.dotfiles",
		URL:  "git@github.com:surminus/dotfiles.git",
	}.Create()

	// This should be a new v.Link resource, since it fails after the first
	// run because the file already exists
	// v.Execute{Command: fmt.Sprintf("ln -s %s/.dotfiles/vimrc %s/.vimrc", home, home)}.Run()

}
