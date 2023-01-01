# Installation

### Pre-requisites
* Go
* Git
* AWS CLI
* AWS CDK
* Node.js
* VS Code
* TypeScript
* AWS Account
* Github Account

#### Linux Installation
##### Go
```bash
dev@dev:~$ sudo apt install software-properties-common apt-transport-https wget
dev@dev:~$ wget -c https://go.dev/dl/go1.xx.xx.linux-amd64.tar.gz
dev@dev:~$ sudo tar -C /usr/local -xzf go1.xx.xx.linux-amd64.tar.gz
dev@dev:~$ sudo vim .profile
```

NOTE: Please remember to change the xx.xx to your desired version of go.

Add this inside the .profile
```bash
## GO configuration ##
PATH="$HOME/bin:$HOME/.local/bin:$PATH"
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

**NOTE**: The changes made to the `.profile` will not be applied until the next time you log into your machine. Please restart your machine.

To check if it's already installed:
```bash
dev@dev:~$ go version
go version go1.xx.xx linux/amd64
```

##### GIT
```bash
dev@dev:~$ sudo apt update
dev@dev:~$ sudo apt-get install git
# Git configuration
dev@dev:~$ git config --global user.name "username"
dev@dev:~$ git config --global user.email "email@email.com"
```

##### AWS CLI
```bash
dev@dev:~$ curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
dev@dev:~$ unzip awscliv2.zip
dev@dev:~$ sudo ./aws/install
```

To check if it's already installed:
```bash
dev@dev:~$ aws --version
aws-cli/2.4.25 Python/3.8.8 Linux/5.13.0-52-generic exe/x86_64.ubuntu.20 prompt/off
```

##### Node.js
```bash
dev@dev:~$ sudo apt update
dev@dev:~$ sudo apt install nodejs
dev@dev:~$ node -v
v10.19.0
```

##### VS Code
```bash
dev@dev:~$ sudo apt update
dev@dev:~$ sudo apt install software-properties-common apt-transport-https
dev@dev:~$ wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
dev@dev:~$ sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
dev@dev:~$ sudo sh -c 'echo "deb [arch=amd64 signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/vscode stable main" > /etc/apt/sources.list.d/vscode.list'
dev@dev:~$ sudo apt update
dev@dev:~$ sudo apt install code
```

**NOTE:** Due to its size, the installation takes approximately 5 minutes.

##### TypeScript
```bash
dev@dev:~$ sudo npm install -g typescript
dev@dev:~$ tsc
Version 4.6.3
tsc: The TypeScript Compiler - Version 4.6.3
```

## Reference
* [Go Installation](https://rmarasigan.github.io/notes/notes/go-lang/Installation.html)
* [TypeScript Installation](https://rmarasigan.github.io/notes/notes/typescript/introduction.html#linux-installation)
* [Installing or updating the latest version of the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)