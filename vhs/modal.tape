# VHS tape for the modal example (two color boxes; Tab+Enter opens picker)
# Run from bubblepicker dir: vhs vhs/modal.tape
# Tab then Enter opens the picker so it is visible in the screenshot.

Output "screenshots/modal.gif"

Set FontSize 14
Set Width 650
Set Height 650
Set Shell "bash"
Set Theme "Builtin Dark"
Set TypingSpeed 75ms
Env COLORTERM truecolor
Env TERM xterm-256color

Hide
Type "go run ./examples/modal"
Enter
Sleep 1s
Show
Sleep 3s
Tab
Sleep 1s
Enter
Sleep 2s

Type "q"
Enter
Sleep 500ms
