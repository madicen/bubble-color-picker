# VHS tape for the swatch example (2×2 grid; key 1 opens modal)
# Run from bubblepicker dir: vhs vhs/swatch.tape
# Key "4" opens the fourth swatch so the picker is visible in the screenshot.

Output "screenshots/swatch.png"

Set FontSize 14
Set Width 1300
Set Height 650
Set Shell "bash"
Set Theme "Builtin Dark"
Set TypingSpeed 75ms
Env COLORTERM truecolor
Env TERM xterm-256color

Type "go run ./examples/swatch"
Enter
Sleep 3s
Type "4"
Sleep 2s

Screenshot "screenshots/swatch.png"

Type "q"
Enter
Sleep 500ms
