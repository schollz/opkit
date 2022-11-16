import os

kits=["Kick","Snare","Bass","Hat","808","Beep","Clap","Bongo"]
kits+=["Kick","Snare","Bass","Hat","808","Beep","Clap","Bongo"]
kits+=["Kick","Snare","Bass","Hat","808","Beep","Clap","Bongo"]
notes=["c","db","d","eb","e","f","gb","g","ab","a","bb","b"]
for j,kit in enumerate(kits):
	for i,note in enumerate(notes):
		fname="kits/{}/{}_{}_{}.aif".format(note,kit,j,note)
		os.system("./postsolarpunk --kit '{}' --tune {} --out {}".format(kit,i+24,fname))
		os.system("sox {} 1.wav".format(fname))
		os.system("audiowaveform -i 1.wav -o {}.png -s 0 -e 11.5 --background-color ffffff".format(fname))
