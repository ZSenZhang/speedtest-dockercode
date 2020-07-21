#!/bin/bash
#used to install tools for running speedtest exps from cloud servers
#usage: chmod 777 ./setupspeedtest.bash; yes | ./setupspeedtest.bash 

# echo "Downloading chromium..."
# sudo apt-get update;
# sudo apt-get install libpcap-dev;
# sudo apt install -y chromium-browser;
LNAME=`echo $USER`
EXPRPROFILE="/home/$LNAME/exprprofile"
cd /home/$LNAME;
rm $EXPRPROFILE
#echo "Install golang..."
#wget -q -O - https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash
#export PATH=$PATH:/usr/local/go/bin
#echo "Downloading ndt command line tool..."
#go get -v github.com/m-lab/ndt7-client-go/cmd/ndt7-client
#go get ./cmd/ndt7-client

sudo mv outdir/softwares/ndt7-client /usr/local/bin/
sudo setcap cap_net_raw,cap_net_admin=eip /usr/sbin/tcpdump


echo "Downloading someta..."
#go version someta
# go get github.com/jsommers/someta# cd $GOPATH/src/github.com/jsommers/someta;
# go build;

# python version someta
# cd /scratch/cloudspeedtest/outdir/softwares/speedtest
# git clone https://github.com/jsommers/metameasurement
# mv  -v ./metameasurement/* ~/scratch/cloudspeedtest/outdir/softwares/speedtest/
# mv ss_monitor.py ./monitors/ss_monitor.py
#sudo apt install python3-pip
sudo apt-get install -y python3 -m venv xenv
source xenv/bin/activate
pip3 install -r /home/$LNAME/outdir/softwares/speedtest/requirements.txt
echo "install python libs"
pip3 install asyncio



echo "Downloading ookla command line tool..."
export INSTALL_KEY=379CE192D401AB61;
export DEB_DISTRO=$(lsb_release -sc);
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys $INSTALL_KEY;
echo "deb https://ookla.bintray.com/debian ${DEB_DISTRO} main" | sudo tee  /etc/apt/sources.list.d/speedtest.list;
sudo apt-get update;
sudo apt-get install speedtest -y;

echo "Installing node and npm..."
wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.35.3/install.sh | bash
echo 'export NVM_DIR=~/.nvm' >> $EXPRPROFILE
echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"'>> $EXPRPROFILE
echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"'>> $EXPRPROFILE
echo 'source ~/.nvm/nvm.sh' >> $EXPRPROFILE
source $EXPRPROFILE

nvm install node
NODEPATH=`ls -1 /home/$LNAME/.nvm/versions/node |head -n1`
echo "export PATH=$PATH:/home/$LNAME/.nvm/versions/node/$NODEPATH/bin" >> $EXPRPROFILE
source $EXPRPROFILE

npm install puppeteer
npm install commander
sudo sysctl -w kernel.unprivileged_userns_clone=1

EXPEXIST=`grep "exprprofile" /home/$LNAME/.profile |wc -l`
if [[ $EXPEXIST == "0" ]]; then
    echo "bash $EXPRPROFILE" >> ~/.profile
fi

