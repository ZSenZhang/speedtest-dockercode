MONTH=$(date -d "$D" '+%m')
YEAR=$(date -d "$D" '+%Y')

FIRSTDIR="$YEAR""$MONTH""01"
echo $FIRSTDIR

ASRANKRIB="/data/external/as-rank-ribs/$FIRSTDIR"
