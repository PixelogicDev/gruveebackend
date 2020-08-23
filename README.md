<!-- markdownlint-disable -->
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
  <span> . </span>
  <a href="README-Support/DEPLOYMENT_STATUS.md">Deployment Status</a>
</h3>
<!-- markdownlint-enable -->
<!-- markdownlint-disable MD013 MD041 -->

Grüvee is an open source social, collaborative playlist made by the [PixelogicDev Twitch Community](https://twitch.tv/pixelogicdev). This project was entirely made live, on Twitch while receiving help from chat and contributing to Pull Requests here!

This is currently the Backend portion of the platform. We are currently building a Mobile client that can be found [here](https://github.com/PixelogicDev/Gruvee-Mobile)!

If you are interested in becoming a member of the team check out the **[PixelogicDev Twitch](https://twitch.tv/pixelogicdev)**, the **[PixelogicDev Discord](https://discord.gg/ubgX6T8)** and **[contribute](#-how-to-contribute)** to this awesome project!
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-5-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

---

# Getting Started

## Tech Stack

| Stack    | Tech                                                                                       |                                                                |
| -------- | :----------------------------------------------------------------------------------------- | :------------------------------------------------------------- |
| IDE      | [Visual Studio Code](https://code.visualstudio.com/)                                       | You can use your preferred IDE but this is the one we like 🙃  |
| Backend  | [Firebase (Repo)](https://github.com/PixelogicDev/Gruvee-Backend)                          | Serverless Functions in Firebase using GoLang                  |
| Frontend | [React Native 0.60](<[LinkToReactNative0.60](https://www.npmjs.com/package/react-native)>) | Utilizing Javascript to develop this cross platform mobile app |

> ALL of these sections are open for contributions and are highly encouraged!

## Running Grüvee Backend Locally

### Golang Setup

We found [this guide](https://www.digitalocean.com/community/tutorials/understanding-the-gopath) pretty helpful in understanding the Golang file structure and how it should be setup. We have tweaked this file structure to fit our Firebase Function setup.

1. Make sure to [install Golang](https://golang.org/dl/) we are using v1.13
1. Find your `GOPATH`. ALWAYS located in this path (`$HOME/go`) unless put otherwise
    1. Example GOPATH: `Users/YourComputerName/go`
    1. Example GOPATH/BIN:`Users/YourComputerName/go/bin`
1. Add GOPATH env variable

```shell
export GOPATH="$HOME/go"
export GOBIN="$HOME/go/bin"
export PATH=$PATH:$GOBIN:$GOPATH
```

### Golang and VSCode Extensions

We have found some awesome workflows and tools to get us up and running with Golang. This repo includes a [`.vscode/settings.json`](.vscode/settings.json) & [`.vscode/extensions.json`](.vscode/extensions.json) which will allow you to automatically download the recommended extensions and keeps your repo styling in sync with all the contributors. If you find more extensions please join the [PixelogicDev Discord](https://discordapp.com/invite/8NFtvp5) and share with us so we can add it to the project!

### Grüvee Backend File Path

Make sure `gruveebackend` is in your GOPATH (This helps a lot. We promise.)

1. Clone repo from Github
1. Open up `GOPATH/src/github.com/`
1. Create folder called `pixelogicdev`
1. `cd pixelogicdev` and move/clone `gruveebackend` into this folder

### Running Functions Locally

When you want to run all the functions locally, all you need to do is run `scripts/gorun.sh`. This will work out of the box on any macOS/Linux/Ubuntu system. If you are running on Windows 10 you will need to follow a [guide like this](https://www.howtogeek.com/249966/how-to-install-and-use-the-linux-bash-shell-on-windows-10/) to get the bash system running on your machine.

This script does the following:

- Adds all the replace lines in all the go mod files
- Builds and runs the `main.go` file in the root

## Auto Tagging and Deploy with Github Actions

GitHub [Actions](https://docs.github.com/en/actions) is used for tagging functions when updated and pushing them to gcp. This is an automated process which has the following requirements:

- Any change made in a function's folder structure (e.g. `/cmd/appleauth/**`) must be accompanied with a new version (maintained in `.version` file).
  - If a change is made to a function and the same tag is used again, the deploy will **fail**!
- All code destined for `master` should/must go through a pull request (pr). No pushes should go directly to master, let's keep good habits!
- GitHub Actions yaml files are kept at `.github/workflows/`
- There is **one** Actions yaml file per **function** per **trigger**. This allows multiple functions to be updated in a push and each will be tagged and deployed as needed.
  - There is currently only a 'push to master' trigger file. Other useful triggers are on PR, which can run linting, tests, and even pushed to staging, vetting the code being staged for production deployment.
- The tag written is taken from the first line of the `.version` file (e.g. - v1.0.0-beta.3: Tweaking...). The function used splits on the colon, uses the first half grabbing everything from the `v` forward.

### Basic Workflow

There is a separate Actions trigger file for each function. The Action will trigger whenever there is a change in the functions directory (e.g. `cmd/tokengen/*`). The following occurs:

1. The code is checked out
1. The version is extracted from the first line of the `.version` file
1. A tag is written to the merge's SHA using the version extracted from the file
1. The config.yml file is written to disk
1. gcloud action is loaded
1. The deploy is run using the `.deployment` file from the function's directory.

### Required Secrets

The Actions configuration requires four [secrets](https://github.com/PixelogicDev/gruveebackend/settings/secrets) to be configured in GitHub repo:

- **PROD_CONFIG_YAML** --  The configuration file used by GCP for the function's variables
- **PROD_CONFIG_YAML_64** -- A base64 encoded version of `PROD_CONFIG_YAML`
- **PROD_CLOUD_AUTH** -- GCP service account which can deploy to the GCP project's function. `JSON` format with no carriage returns or line feeds
- **PROD_GCP_PROJECT_ID** -- The GCP project id being deployed to

You may be wondering why there are both the yaml config file and the yaml config file base64 encoded. The base64 version is used to write the file used by the deploy process. The regular yaml version is used to redact all the values from the log. It's a work-around to ensure values aren't leaked in the logs. IT IS VERY IMPORTANT TO UPDATE BOTH OF THESE FILES WHENEVER THERE IS A CONFIGURATION CHANGE!

### Setting up a new function for tagging and deploy on push to master

Each function has it's own GitHub Actions file under `.github/workflows`. At the time of writing all of these files are triggered on **push master**. The steps for creating a file for a function is _very_ straight forward. There are only **three** lines to update (1, 8, 17).

1. Copy `template_pushMaster.yml` file from the workflows template directory, `.github/workflow_templates/`, into the workflows directory, `.github/workflows`.
1. Rename the file copied with the function's name. E.g. function `tokengen` would be renamed: `tokengen_pushMaster.yml`
1. There are three lines to update in the file with the function's name: 1, 8, & 17
    - **line 1**: Replace [function name] with the function's name (e.g. `tokengen`)
    - **line 8**: Replace [function name] with the function's name (e.g. `tokengen`)
    - **line 17**: Replace [function name] with the function's name (e.g. `tokengen`)
1. That is all, GitHub Actions is now configured for the function. (after the files are pushed to master that is 😃)
1. Add a new badge to the [Deployment Status](./README-Support/DEPLOYMENT_STATUS.md) page. You can get the badge from the Actions page or create from one of the existing ones by changing the function's name in the label and url.

### Deployment Status

See [Deployment Status Page](./README-Support/DEPLOYMENT_STATUS.md) to see the GitHub Actions deployment status of all the functions.

# ❓ FAQ

To understand how certain aspects of this project work please see the [FAQ Documentation](README-Support/FAQ.md)!

# 🤘 Contributing Changes

See [CONTRIBUTING.md](README-Support/CONTRIBUTING.md)

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/adilanchian"><img src="https://avatars0.githubusercontent.com/u/13204620?v=4" width="100px;" alt=""/><br /><sub><b>Alec Dilanchian</b></sub></a><br /><a href="https://github.com/PixelogicDev/gruveebackend/commits?author=adilanchian" title="Code">💻</a> <a href="#maintenance-adilanchian" title="Maintenance">🚧</a> <a href="https://github.com/PixelogicDev/gruveebackend/commits?author=adilanchian" title="Documentation">📖</a></td>
    <td align="center"><a href="https://edburtnieks.me"><img src="https://avatars0.githubusercontent.com/u/47947787?v=4" width="100px;" alt=""/><br /><sub><b>Edgar Burtnieks</b></sub></a><br /><a href="https://github.com/PixelogicDev/gruveebackend/commits?author=edburtnieks" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/isabellabrookes"><img src="https://avatars1.githubusercontent.com/u/12928252?v=4" width="100px;" alt=""/><br /><sub><b>Isabella Brookes</b></sub></a><br /><a href="https://github.com/PixelogicDev/gruveebackend/commits?author=isabellabrookes" title="Documentation">📖</a> <a href="#maintenance-isabellabrookes" title="Maintenance">🚧</a></td>
    <td align="center"><a href="https://github.com/LeviHarrison"><img src="https://avatars3.githubusercontent.com/u/54278938?v=4" width="100px;" alt=""/><br /><sub><b>Levi Harrison</b></sub></a><br /><a href="https://github.com/PixelogicDev/gruveebackend/commits?author=LeviHarrison" title="Code">💻</a></td>
    <td align="center"><a href="http://blog.brettski.com"><img src="https://avatars3.githubusercontent.com/u/473633?v=4" width="100px;" alt=""/><br /><sub><b>Brett Slaski</b></sub></a><br /><a href="https://github.com/PixelogicDev/gruveebackend/commits?author=brettski" title="Code">💻</a> <a href="https://github.com/PixelogicDev/gruveebackend/commits?author=brettski" title="Documentation">📖</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
