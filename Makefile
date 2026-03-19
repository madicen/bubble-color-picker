.PHONY: screenshots clean

screenshots: screenshots/simple.png screenshots/swatch.png screenshots/modal.png

screenshots/simple.png: vhs/simple.tape
	vhs vhs/simple.tape

screenshots/swatch-before.png: vhs/swatch-before.tape
	mkdir -p screenshots
	vhs vhs/swatch-before.tape
	@[ -f $@ ] || (echo "Error: $@ was not created by vhs"; exit 1)

screenshots/swatch-after.png: vhs/swatch-after.tape
	mkdir -p screenshots
	vhs vhs/swatch-after.tape
	@[ -f $@ ] || (echo "Error: $@ was not created by vhs"; exit 1)

screenshots/swatch.png: screenshots/swatch-before.png screenshots/swatch-after.png
	@if command -v convert >/dev/null 2>&1; then \
		convert +append screenshots/swatch-before.png screenshots/swatch-after.png screenshots/swatch.png; \
	elif command -v magick >/dev/null 2>&1; then \
		magick +append screenshots/swatch-before.png screenshots/swatch-after.png screenshots/swatch.png; \
	else \
		echo "Error: ImageMagick not found. Please install 'imagemagick' (providing 'convert' or 'magick')."; \
		exit 1; \
	fi

screenshots/modal.png: vhs/modal.tape
	vhs vhs/modal.tape

clean:
	rm -f screenshots/*.png screenshots/*.gif
