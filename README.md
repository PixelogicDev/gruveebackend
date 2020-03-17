<h1 align="center">Grüvee-Backend</h1>
<p align="center">
  <strong>Let's get Grüvee with a new social, collaborative playlist for iPhone and Android</strong>
</p>

<p align="center">
    <a href="https://discordapp.com/invite/ubgX6T8">
        <img src="https://img.shields.io/discord/391635862959554561?label=Discord" alt="Discord members online" />
    </a>
    <a href="https://github.com/pixelogicdev/gruvee">
        <img alt="GitHub issues" src="https://img.shields.io/github/issues/pixelogicdev/gruvee-backend">
    </a>
    <a href="#-how-to-contribute">
        <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg" alt="PRs welcome!" />
    </a>
</p>

<h3 align="center">
 <a href="#getting-started">Getting Started (Coming #soon)</a>
  <span> · </span>
  <a href="#running-grüvee-locally">Running Grüvee Locally (Coming #soon)</a>
  <span> · </span>
  <a href="#-how-to-contribute">How to Contribute (Coming #soon)</a>
  <span> · </span>
  <a href="#current-contributors">Current Contributors (Coming #soon)</a>
  
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

### Golang Setup (https://www.digitalocean.com/community/tutorials/understanding-the-gopath)

1. Make sure to install Golang (we are using v1.13)
2. Find your GOPATH. ALWAYS located in this path (`$HOME/go`) unless put otherwise
   1. Example GOPATH: `Users/alecdilanchian/go`
   2. Example GOPATH/BIN:`Users/alecdilanchian/go/bin`
3. Add GOPATH env variable
   Example:

```
export GOPATH="$HOME/go"
export GOBIN="$HOME/go/bin"
export PATH=$PATH:$GOBIN:$GOPATH
```

4. Make sure `gruveebackend` is in your GOPATH (This helps a lot. I promise)

   1. Clone repo from Github
   2. Open up `GOPATH/src/github.com/
   3. Create folder called `pixelogicdev`
   4. `cd pixelogicdev` and move `gruveebackend` into the folder

5. We deploy to Github so our internal packages will be good to go

   1. They need to be in master to be picked up properly
   2. We need to `git tag` them
   3. We can then deploy to Firebase

6. Since packages are being pulled from our actual Github Repo, we will need to use the `replace` command in order to actually see our changes go through. These are located in mod files that need them.

### Golang and VSCode Extensions

1. Will make sure to have a vscode settings file included in the repo for consistent settings, but for informational purposes download extension from here: https://code.visualstudio.com/docs/languages/go
2. This repo includes a [`.vscode/settings.json`](.vscode/settings.json) & [`.vscode/extensions.json`](.vscode/extensions.json)
3. These should get you started pretty much right away when you open up the repo

### Running Functions Locally
When you want to run all the functions locally, all you need to do is run `scripts/gorun.sh`. (Currently there is no windows equivalent, but should make one #SOON.) Essentially this will:
- Add all the replace lines in all the go mod files
- Build and run the `main.go` file in the root

When adding new functions to this project you will need to also update the build script as follows:
1. Add a variable for the function replace path. For example if your new function is called `addCoolPerson` create a replace variable that looks like this: `addCoolPerson=github.com/pixelogicdev/gruveebackend/cmd/addcoolperson=../cmd/addcoolperson`
2. Add a new if statement for `addcoolperon/` directory in the if logic
3. Make sure to add the new `go mod edit -replace $addCoolPerson` to all other functions (for now it's easier this way to make sure we aren't missing anything)

### Deploy Function To Cloud

- Change `config.yaml` file `ENIRONMENT: PROD`
- cd into cmd/function folder and deploy from there using an example like so:
  `gcloud functions deploy authorizeWithSpotify --entry-point AuthorizeWithSpotify --runtime go113 --trigger-http --env-vars-file internal/config.yaml --allow-unauthenticated`
