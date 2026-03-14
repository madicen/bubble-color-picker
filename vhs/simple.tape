# VHS tape for the simple example (picker is the whole app)
# Run from bubblepicker dir: vhs vhs/simple.tape
# Picker is always visible; Tab + arrows show it in action.

Output "screenshots/simple.png"

Set FontSize 14
Set Width 900
Set Height 500
Set Shell "bash"
Set Theme "Builtin Dark"

Type "go run ./examples/simple"
Enter
Sleep 3s
Tab
Sleep 300ms
Right
Right
Down
Sleep 500ms

Screenshot "screenshots/simple.png"

Type "q"
Enter
Sleep 500ms
