# VHS tape for the swatch example
# Run from bubblepicker dir: vhs vhs/swatch-after.tape (used for swatch.gif)
# Key "4" opens the fourth swatch so the picker is visible in the screenshot.

Output "screenshots/swatch.gif"

Set FontSize 14
Set Width 650
Set Height 650
Set Shell "bash"
Set Theme "Builtin Dark"
Set TypingSpeed 75ms
Env COLORTERM truecolor
Env TERM xterm-256color

Hide
Type "go run ./examples/swatch"
Enter
Sleep 1s
Show
Sleep 2s
Type "4"
Sleep 2s
Left 6
Sleep 1s
Tab
Sleep 1s
Left 2
Sleep 1s
Down 2
Sleep 1s
Left 2
Sleep 1s
Enter
Sleep 2s

Type "q"
Sleep 1s
