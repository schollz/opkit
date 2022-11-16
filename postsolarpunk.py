import os

kits=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
kits+=["Kick","Snare","Bass","Hat"]
notes=["c","db","d","eb","e","f","gb","g","ab","a","bb","b"]
for j,kit in enumerate(kits):
    for i,note in enumerate(notes):
        fname="kits/{}/{}_{}_{}.aif".format(note,kit,j,note)
        jj=j
        while os.path.isfile(fname):
            jj=jj+1
            fname="kits/{}/{}_{}_{}.aif".format(note,kit,jj,note)
        os.system("./opkit --kit '{}' --tune {} --out {}".format(kit,i+24,fname))
        os.system("sox {} 1.wav".format(fname))
        os.system("audiowaveform -i 1.wav -o {}.png -s 0 -e 11.5 --background-color ffffff".format(fname))
