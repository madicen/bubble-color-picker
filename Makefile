.PHONY: screenshots clean

screenshots: screenshots/simple.png screenshots/swatch-before.png screenshots/swatch-after.png screenshots/modal.png

screenshots/simple.png: vhs/simple.tape
	vhs vhs/simple.tape

screenshots/swatch-before.png: vhs/swatch-before.tape
	vhs vhs/swatch-before.tape

screenshots/swatch-after.png: vhs/swatch-after.tape
	vhs vhs/swatch-after.tape

screenshots/modal.png: vhs/modal.tape
	vhs vhs/modal.tape

clean:
	rm -f screenshots/*.png screenshots/swatch-before.png screenshots/swatch-after.png
