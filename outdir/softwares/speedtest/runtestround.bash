#!/usr/bin/bash
LNAME=`echo $USER`
HNAME=`cat /home/$LNAME/hostalias`
HOSTTYPE=`cat /home/$LNAME/vmtype`
LISTFILE="/home/$LNAME/outdir/datafiles/$HOSTTYPE-serverlist.csv"
PRIETH=`ls -1 /sys/class/net/ |head -n1`
STARTTS=`date +%s`
OUTPUTDIR="results/speedtest"
SPTESTDIR="outdir/softwares/speedtest"
#export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/ubuntu/.nvm/versions/node/v14.0.0/bin
#detect Node version

NODEPATH=`ls -1 /home/$LNAME/.nvm/versions/node |head -n1`
if [[ $NODEPATH == "" ]]; then
    NODEPATH="v14.2.0"
fi

export PATH=/home/$LNAME/.local/bin:/home/$LNAME/.nvm/versions/node/$NODEPATH/bin:/home/$LNAME/go/bin:/home/$LNAME/.go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin

#start the rate limit
echo "Expr round start\n"
sudo bash /home/$LNAME/outdir/softwares/ratelimit.bash
sleep 1
shuf $LISTFILE --output $LISTFILE
IFS=$(echo -en "\n\b")
for servers in `grep "$HNAME|" $LISTFILE`; do
    TESTMETHOD=`echo "$servers" | cut -d'|' -f4`
    TESTSERVER=`echo "$servers" | cut -d'|' -f5`
    FARIP=`echo "$servers" | cut -d'|' -f2`
    FILEPREFIX=$HNAME.$STARTTS.$FARIP.$TESTMETHOD.tar.bz2
    echo "$FILEPREFIX\n";
    cd $SPTESTDIR;
    timeout 90 python3 runspeedtest.py 1 $TESTMETHOD "$TESTSERVER" $PRIETH
    #compress the results
    cd /home/$LNAME;
    sudo tar cfvj $OUTPUTDIR/$FILEPREFIX $SPTESTDIR/$TESTMETHOD/* --remove-files
done
echo "Expr round ends\n"
sudo bash /home/$LNAME/outdir/softwares/removeratelimit.bash
bash /home/$LNAME/outdir/softwares/runtestservertr.bash outdir/datafiles results/trace
#transfer results after one round
sudo bash /home/$LNAME/outdir/softwares/syncstorage.bash
echo "Sync data ends\n"
