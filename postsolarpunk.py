import os

kits=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
kits+=["Combo","Snares","Hat","Noise"]
# notes=["c","db","d","eb","e","f","gb","g","ab","a","bb","b"]
for j,kit in enumerate(kits):
    # for i,note in enumerate(notes):
    note="g"
    i=7
    fname="kits/{}/{}_{}_{}.aif".format(note,kit,j,note)
    jj=j
    while os.path.isfile(fname):
        jj=jj+1
        fname="kits/{}/{}_{}_{}.aif".format(note,kit,jj,note)
    maxx=2
    if kit=="Hat":
        maxx=0.75
    os.system("./opkit --min 0.03 --max {} --kit '{}' --tune {} --out {}".format(maxx,kit,i+24,fname))
    os.system("sox {} 1.wav".format(fname))
    os.system("audiowaveform -i 1.wav -o {}.png -s 0 -e 11.5 --background-color ffffff".format(fname))
