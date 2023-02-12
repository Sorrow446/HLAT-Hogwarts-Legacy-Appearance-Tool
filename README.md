# HLAT---Hogwarts-Legacy-Appearance-Tool
Tool for appearance manipulation for Hogwarts Legacy written in Go.    
[Windows binaries](https://github.com/Sorrow446/HLAT-Hogwarts-Legacy-Appearance-Tool/releases)

## Why?
- Allows users to create presets from their own save files and share them for others to use in their saves.
- Allows you to edit parts of characters the game wouldn't allow you to (gender, head etc.).

## Usage
Create JSON appearance preset from save file:   
`hlat_x64.exe export -i HL-00-00.sav -o eve.json`

Inject JSON appearance preset into save file:   
`hlat_x64.exe import -i eve.json -o HL-00-00.sav`

```
Usage: main.exe --inpath INPATH [--outpath OUTPATH] [--orig-name] COMMAND

Positional arguments:
  COMMAND                import or export

Options:
  --inpath INPATH, -i INPATH
                         Path of input file. JSON appearance if import, save file if export.
  --outpath OUTPATH, -o OUTPATH
                         Path of output file. Save file if import, JSON appearance if export.
  --orig-name            Keep original character name.
  --help, -h             display this help and exit
```

## Disclaimer
- I will not be responsible for any possibility of save corruption.    
- Hogwarts Legacy brand and name is the registered trademark of its respective owner.    
- HLAT has no partnership, sponsorship or endorsement with Avalanche Software or Warner Bros. Games.
