# VHS tape for the swatch example (2×2 grid; key 1 opens modal)
# Run from bubblepicker dir: vhs vhs/swatch.tape
# Key "1" opens the first swatch so the picker is visible in the screenshot.

Output "screenshots/swatch.png"

Set FontSize 14
Set Width 900
Set Height 500
Set Shell "bash"
Set Theme "Builtin Dark"

Type "go run ./examples/swatch"
Enter
Sleep 3s
Type "1"
Sleep 2s

Screenshot "screenshots/swatch.png"

Type "q"
Enter
Sleep 500ms
