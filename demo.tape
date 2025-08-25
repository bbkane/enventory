# https://github.com/charmbracelet/vhs

Require enventory

Output demo.gif

Set Shell zsh
Set PlaybackSpeed 0.5
Set TypingSpeed 75ms

Set FontSize 22
Set Height 650
Set Width 950
Set Padding 25

Hide

Type 'setopt interactive_comments'
Enter

Type 'source ~/Git-GH/F-Sy-H/F-Sy-H.plugin.zsh'
Enter

Type 'enventory env create --name shared-env'
Enter

Type 'enventory var create --name SHAREDVAR --value "Hello from shared env" --env shared-env'
Enter

Type 'clear'
Enter

Show

Type 'eval "$(enventory shell zsh init)"  # add this to ~/.zshrc'
Sleep 200ms
Enter
Sleep 1s

Type 'enventory env create'
Sleep 200ms
Enter
Sleep 1s

Type "enventory var create --name MYENVVAR"
Sleep 200ms
Enter
Sleep 1s

Type "Hello enventory"
Sleep 200ms
Enter
Sleep 1s

Type "# reference a variable from another environment"
Enter

Type "enventory var ref create --name SHAREDVAR --ref-env shared-env --ref-var SHAREDVAR"
Sleep 200ms
Enter
Sleep 1s

Type "enventory env show --mask false"
Sleep 200ms
Enter
Sleep 1s

Type 'cd .  # simulate directory change'
Sleep 200ms
Enter
Sleep 1s

Type "echo $MYENVVAR"
Sleep 200ms
Enter
Sleep 1s

Sleep 3s
