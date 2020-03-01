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

## Important Notes

### Deploy Function To Cloud

- cd into auth folder and deploy from there
  `gcloud functions deploy <FunctionName> --entry-point <ActualFunctionFileName> --runtime go113 --trigger-http --allow-unauthenticated`

- when testing locally, cd into cmd folder, run go build and then run ./cmd
- `export GOOGLE_APPLICATION_CREDENTIALS="../../config/adminSdkSecret.json"`
