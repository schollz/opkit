kits: postsolarpunk
	mkdir -p kits/c
	mkdir -p kits/db
	mkdir -p kits/d
	mkdir -p kits/eb
	mkdir -p kits/e
	mkdir -p kits/f
	mkdir -p kits/gb
	mkdir -p kits/g
	mkdir -p kits/ab
	mkdir -p kits/a
	python3 run.py

postsolarpunk:
	go build -v

clean:
	rm -rf postsolarpunk
	rm -rf kits
	rm -rf *.wav
	rm -rf *.aif
	rm -rf *.png
	rm -rf concat*
