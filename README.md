<h1 align="center">Grüvee-Backend</h1>
<p align="center">
  <strong>Let's get Grüvee with a new social, collaborative playlist for iPhone and Android</strong>
</p>

<p align="center">
    <a href="https://discordapp.com/invite/8NFtvp5">
        <img src="https://img.shields.io/discord/391635862959554561?label=Discord" alt="Discord members online" />
    </a>
    <a href="https://github.com/pixelogicdev/gruvee">
        <img alt="GitHub issues" src="https://img.shields.io/github/issues/pixelogicdev/gruveebackend">
    </a>
    <a href="#-how-to-contribute">
        <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg" alt="PRs welcome!" />
    </a>
</p>

<h3 align="center">
 <a href="#getting-started">Getting Started</a>
  <span> · </span>
  <a href="#running-grüvee-backend-locally">Running Grüvee Backend Locally</a>
  <span> · </span>
  <a href="README-Support/CONTRIBUTING.md#-how-to-contribute">How to Contribute</a>
  <span> · </span>
  <a href="README-Support/FAQ.md">FAQ</a>
  <span> · </span>
  <a href="README-Support/CONTRIBUTING.md#-current-contributors">Current Contributors</a>
</h3>

Grüvee is an open source social, collabortive playlist made by the [PixelogicDev Twitch Community](https://twitch.tv/pixelogicdev). This project was entirely made live, on Twitch while receiving help from chat and contributing to Pull Requests here!

This is currently the Backend portion of the platform. We are currently building a Mobile client that can be found [here](https://github.com/PixelogicDev/Gruvee-Mobile)!

If you are interested in becoming a member of the team check out the **[PixelogicDev Twitch](https://twitch.tv/pixelogicdev)**, the **[PixelogicDev Discord](https://discord.gg/ubgX6T8)** and **[contribute](#-how-to-contribute)** to this awesome project!

---

# Getting Started

## Tech Stack

| Stack    | Tech                                                                                       |                                                                |
| -------- | :----------------------------------------------------------------------------------------- | :------------------------------------------------------------- |
| IDE      | [Visual Studio Code](https://code.visualstudio.com/)                                       | You can use your preferred IDE but this is the one we like 🙃  |
| Backend  | [Firebase (Repo)](https://github.com/PixelogicDev/Gruvee-Backend)                          | Serverless Functions in Firebase using GoLang                  |
| Frontend | [React Native 0.60](<[LinkToReactNative0.60](https://www.npmjs.com/package/react-native)>) | Utilising Javascript to develop this cross platform mobile app |

> ALL of these sections are open for contributions and are highly encouraged!

## Running Grüvee Backend Locally

### Golang Setup

We found [this guide](https://www.digitalocean.com/community/tutorials/understanding-the-gopath) pretty helpful in understanding the Golang file structure and how it should be setup. We have tweaked this file structure to fit our Firebase Function setup.

1. Make sure to [install Golang](https://golang.org/dl/) we are using v1.13
2. Find your `GOPATH`. ALWAYS located in this path (`$HOME/go`) unless put otherwise
   1. Example GOPATH: `Users/YourComputerName/go`
   2. Example GOPATH/BIN:`Users/YourComputerName/go/bin`
3. Add GOPATH env variable

```
export GOPATH="$HOME/go"
export GOBIN="$HOME/go/bin"
export PATH=$PATH:$GOBIN:$GOPATH
```

### Golang and VSCode Extensions

We have found some awesome workflows and tools to get us up and running with Golang. This repo includes a [`.vscode/settings.json`](.vscode/settings.json) & [`.vscode/extensions.json`](.vscode/extensions.json) which will allow you to automatically download the recommended extenstions and keeps your repo styling in sync with all the contributors. If you find more extenstions please join the [PixelogicDev Discord](https://discordapp.com/invite/8NFtvp5) and share with us so we can add it to the project!

### Grüvee Backend File Path

Make sure `gruveebackend` is in your GOPATH (This helps a lot. We promise.)

1.  Clone repo from Github
2.  Open up `GOPATH/src/github.com/`
3.  Create folder called `pixelogicdev`
4.  `cd pixelogicdev` and move/clone `gruveebackend` into this folder

### Running Functions Locally

When you want to run all the functions locally, all you need to do is run `scripts/gorun.sh`. This will work out of the box on any macOS/Linux/Ubuntu system. If you are running on Windows 10 you will need to follow a [guide like this](https://www.howtogeek.com/249966/how-to-install-and-use-the-linux-bash-shell-on-windows-10/) to get the bash system running on your machine.

This script does the following:

```
Adds all the replace lines in all the go mod files
Builds and runs the `main.go` file in the root
```

# ❓ FAQ

To understand how certain aspects of this project work please see the [FAQ Documentation](README-Support/FAQ.md)!

# 🤘 Contributing Changes

See [CONTRIBUTING.md](README-Support/CONTRIBUTING.md)
