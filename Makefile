.PHONY: screenshots clean

screenshots: screenshots/simple.png screenshots/swatch.gif screenshots/modal.gif

screenshots/simple.png: vhs/simple.tape
	vhs vhs/simple.tape

screenshots/swatch.gif: vhs/swatch.tape
	vhs vhs/swatch.tape

screenshots/modal.gif: vhs/modal.tape
	vhs vhs/modal.tape

clean:
	rm -f screenshots/*.png screenshots/*.gif
