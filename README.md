# myduct

My personal development environment, setup using
[viaduct](https://github.com/surminus/viaduct).

Currently used as a way to understand what features need adding to Viaduct
itself.

## My quickstart

### Install Git

```
sudo apt install git curl xclip -y
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

Download a binary from GitHub releases.

Configure known hosts:
```
sudo mkdir -p /root/.ssh && sudo chmod 0600 /root/.ssh
ssh-keyscan github.com | tee ~/.ssh/known_hosts | sudo tee /root/.ssh/known_hosts
```

Add identity:
```
ssh-add -k
```

Configure for home install (optional):
```
touch ~/.myducthome
```

Configure system:
```
sudo ./myduct
```
