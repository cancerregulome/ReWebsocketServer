N=$1
HOST=$2

seq 0 1 $N | xargs -P $N -I I ./testclient -hostname $HOST  -username userI > output/outputI.out
