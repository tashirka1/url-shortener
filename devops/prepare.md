# apt update
```
sudo apt update && sudo apt upgrade -y
sudo apt install curl wget htop tree ufw git nano build-essential -y
```

# install docker
```
curl -fsSL https://get.docker.com | sh
```

# add user
```
sudo adduser tashirka
sudo echo "tashirka ALL=(ALL)  ALL" >> /etc/sudoers
sudo su tashirka
```

# add deploy key
```
mkdir ~/.ssh
echo "<public key>" >> ~/.ssh/authorized_keys
```

# docker group
```
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker
```

# create ssh key
```
ssh-keygen -t ed25519 -C "rishatsharafiev@ya.ru"
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
```

# clone project
```
git clone git@github.com:tashirka1/url-shortener.git
```

# firewall
```
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow https
sudo ufw allow http
sudo ufw enable
```

### swap
```
sudo swapoff -a
sudo dd if=/dev/zero of=/swapfile bs=1M count=2048
sudo chmod 0600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

sudo nano /etc/fstab # add "/swapfile none swap sw 0 0"
sudo swapon --all
sudo swapon --show
```

### swappiness
```
sudo nano /etc/sysctl.conf
vm.swappiness = 20
sudo sysctl -p
```

### fail2ban
```
sudo apt install fail2ban -y
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo systemctl restart fail2ban
sudo systemctl enable fail2ban
sudo fail2ban-client status
```

### ssh config
```
sudo nano /etc/ssh/sshd_config

PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
```

### install go and goose
```
wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export PATH=$PATH:~/go/bin' >> ~/.profile
go install -v -tags='no_postgres no_mysql no_mssql no_vertica no_clickhouse no_libsql no_ydb' github.com/pressly/goose/v3/cmd/goose@latest
```
