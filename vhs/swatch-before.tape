# VHS tape for the swatch example (before)
# Run from bubblepicker dir: vhs vhs/swatch-before.tape

# Output "screenshots/swatch-before.gif"

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
Sleep 5s

Screenshot "screenshots/swatch-before.png"

Type "q"
Sleep 1s
