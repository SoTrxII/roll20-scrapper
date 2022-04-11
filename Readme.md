# Serverless Roll20 Scrapper

[![codecov](https://codecov.io/gh/SoTrxII/roll20-scrapper/branch/master/graph/badge.svg?token=YI8X1HA6I7)](https://codecov.io/gh/SoTrxII/roll20-scrapper)

This project is a serverless (OpenFaas flavored) implementation of a [Roll20](https://roll20.net/welcome) scrapper.
Although all function share a single core, each of them is distributed as its own container to leverage scalability.

Current functionnalities includes :

- Retrieving players from a game
- Make the bot account join the game as a player (necessary for other functions)

## Configure

All the functions needs the following environment variables:

- **ROLL20_USERNAME**: Username of the bot account to use
- **ROLL20_PASSWORD**: Password of the bot account to use
- **ROLL20_BASE_URL**: Roll20 base URL. Value should be "https://app.roll20.net/". This is a variable for future
  proofing and testing purposes.

## Deploying

## Local development

This scrapper uses an account to retrieve data from a roll20 game. 

