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

Type "enventory env show"
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
