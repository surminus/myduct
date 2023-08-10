# myduct

My personal development environment, setup using
[viaduct](https://github.com/surminus/viaduct).

Currently used as a way to understand what features need adding to Viaduct
itself.

## My quickstart

### Install Git

```
sudo apt install git curl -y
```

### Generate and add SSH key to GitHub

Create a password and copy into paste buffer:

```
openssl rand -hex 32 > ssh.pw
cat ssh.pw | xclip -selection clipboard
```

Generate a new key:

```
ssh-keygen -t ed25519
```

Login and add it to GitHub.

### Setup sudo

```
echo "Defaults env_keep+=SSH_AUTH_SOCK" | sudo tee -a /etc/sudoers
```

Log out and log back in

### Install and run

Allows running without Go already installed.

Set latest version, example:
```
export MYDUCT_VERSION=v11
```

Install the binary:

```
cd /tmp
wget https://github.com/surminus/myduct/releases/download/${MYDUCT_VERSION}/myduct_${MYDUCT_VERSION}_$(uname -s)_$(uname -p).tar.gz
tar zxvf myduct_${MYDUCT_VERSION}_$(uname -s)_$(uname -p).tar.gz
```

Run it:
```
sudo /tmp/myduct
```
