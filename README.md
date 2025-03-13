# Overview
**pdx-deepl** is a tool to do incremental auto translation of mods for Paradox games using DeepL.

## How does it work?
The tool will use a base language and check whether a localization key
from the base language is either missing or outdated in other languages.
If a key is missing or outdated, it will translate the base language localization
and add the translation to the same file for another language.
Also, the translation is marked with a checksum of the base language key so that in
a second run pdx-deepl can check whether a key was updated.

## Supported Games
- Victoria 3
- Crusader Kings 3
- Others may work but are untested

## Usage
Help dialog (`pdx-deepl -h`):
```
Usage of pdx-deepl:
  -api-token string
    	Required: Deepl API Token
  -api-type string
    	Optional: Whether to use free or paid Deepl API (default "free")
  -config string
    	Optional: Path to translation config file (default "translation-config.json")
  -localization string
    	Optional: Path to localization directory of your mod (default ".")
```