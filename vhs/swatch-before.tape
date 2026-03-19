# VHS tape for the swatch example (before)
# Run from bubblepicker dir: vhs vhs/swatch-before.tape

Output "screenshots/swatch-before.png"

Set FontSize 14
Set Width 450
Set Height 500
Set Shell "bash"
Set Theme "Builtin Dark"
Set TypingSpeed 75ms
Env COLORTERM truecolor
Env TERM xterm-256color

Type "go run ./examples/swatch"
Enter
Sleep 3s

Screenshot "screenshots/swatch-before.png"
