# Records demo.gif for the README. Run with: vhs demo.tape
Output demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 900
Set Height 1100
Set TypingSpeed 120ms

Hide
Type "export XDG_CONFIG_HOME=$(mktemp -d) && go build -o /tmp/sudo-ku-demo ./cmd/sudo-ku && clear"
Enter
Sleep 3s
Show

Type "/tmp/sudo-ku-demo"
Enter
Sleep 2s

# Tutorial: rules, move to the yellow cell, place the 5.
Enter
Sleep 1s
Down 4
Right 4
Sleep 500ms
Type "5"
Sleep 1s

# Notes on the cell two to the left of the top row's blank.
Up 4
Left 2
Type "n14"
Sleep 1s
Type "n"

# A conflict, then undo.
Down
Right 4
Type "6"
Sleep 1s
Type "u"
Sleep 1.5s

Escape
Sleep 1.5s
