#!/usr/bin/bash
#push results to cloud storage
HOSTTYPE=`cat /home/ubuntu/vmtype`
HNAME=`cat /home/ubuntu/hostalias`
RESULTDIR="/home/ubuntu/results"

if [[ $HOSTTYPE == "amazon" ]]; then
    aws s3 cp results/ "s3://cloudspeedtest/$HNAME/" --recursive --exclude "*.warts"
fi
if [[ $HOSTTYPE == "azure" ]]; then
    azcopy login --identity;
    azcopy copy "results/*" "https://cloudspeedtestblob.blob.core.windows.net/cloudspeedtestcontainer/$HNAME/" --recursive=true --exclude-pattern "*.warts";
fi
if [[ $HOSTTYPE == "google" ]]; then
    #only transfer compressed files
    find results/bdrmap -type f -name *.tar.bz2 | gsutil -m cp -I gs://cloudspeedtest/$HNAME/results/bdrmap/
    find results/speedtest -type f -name *.tar.bz2 | gsutil -m cp -I gs://cloudspeedtest/$HNAME/results/speedtest/
    find results/trace -type f -name *.tar.bz2 | gsutil -m cp -I gs://cloudspeedtest/$HNAME/results/trace/
#    gsutil cp -r results gs://cloudspeedtest/$HNAME
fi
if [ $? -eq 0 ]; then
    sudo rm -rf results/bdrmap/*.tar.bz2 results/speedtest/*.tar.bz2 results/trace/*.tar.bz2
else
    echo "cloud cp failed";
fi

