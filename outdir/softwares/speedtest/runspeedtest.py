# Adapted from control_nossh.pl
# Used to obtain IP addresses of speed test websites

import re, sys, os, time, subprocess
from subprocess import check_output

# predefined parameters
repeat = 20
rate_control = 1
eth0 = "ens5"
username = "rainayangg" 
my_ip = "172.16.134.195"

comcast_server = "portland"
ookla_server = "san_diego"
ookla_net = "scalematrix"

sometabin = "./someta"

try:
  repeat = int(sys.argv[1])
  ss_platform = sys.argv[2]
  server_hint = sys.argv[3]
  eth0 = sys.argv[4]
except:
  pass

# dlbw = ["10000", "25000", "50000","100000", "300000", "500000", "unlimited"]
# upbw = ["1000", "5000", "10000", "10000", "100000","30000", "500000", "unlimited"]

# speedtest = ["ndt","comcast", "ookla"]
display_filter = {'ndt':"dns.resp.name contains measurement-lab.org", 'comcast':"dns.resp.name contains sys.comcast.net", 'ookla':"tcp.port == 8080"}

# queue = ["ethernet-default", "queue-200"]
# queue = ["ethernet-default"]
# shape the bandwidth using the max Rx rate of the interface in kbps
dleth = "ether5-local-slave";
uleth = "ether4-local-slave";

# not using it on macOS
#my $tcpdumppath = "/usr/sbin/tcpdump";
#outgoing interface

print("create output dir\n")
subprocess.run(["mkdir", ss_platform])
subprocess.run(["chmod", "777", ss_platform])

# bwlen = len(dlbw)
# qlen = len(queue)

# if rate_control != 1:
  # no rate control
  # qlen = 1

# initialize the data structure to store IP addresses
ip_all = dict()
ip_all[ss_platform] = set()

# loop over each network setting
# for bw in range(bwlen):
ip_res = set()
for r in range(repeat):
  ctime = time.time()
  exprname = ss_platform + "_"  + str(ctime)
  if rate_control != 1:
    # no rate control, just need the timestamp
    exprname = ss_platform + "_" + str(ctime)
  print(exprname)
  pcapname = ss_platform + "/" + exprname + ".pcap"
  pid = os.fork()
  if pid == 0:
    subprocess.run(["tcpdump", "-i", eth0, "-n", "-s", "100", "-w", pcapname])
    os._exit(0)
  time.sleep(1)
  nodename = ss_platform + ".js"
  outputname = ss_platform + "/" + exprname
  metaname = ss_platform+"/"+exprname+".meta"
  # block at node until if finishes
  if (ss_platform == 'comcast'):
    comcast_server = server_hint.split(',')[0]
    nodecmd = "node "+nodename+" "+outputname+" -host "+comcast_server 
    print(nodecmd)

  elif (ss_platform == 'ookla'):
    sys_user = check_output(['whoami'], shell=True).decode().split('\n')[0]
    is_eu = (sys_user == 'caida')
    ookla_server = server_hint.split(' - ')[0].split(',')[0].replace(" ","_")
    ookla_net = server_hint.split(' - ')[1].replace(" ","_")
    if not is_eu:
      nodecmd = "node "+nodename+" "+outputname+" -city "+ookla_server+" -net "+ookla_net
    else:
      nodecmd = "node "+"ookla_eu.js"+" "+outputname+" -city "+ookla_server+" -net "+ookla_net
    print(nodecmd)

  else:
    tmp = server_hint.split('.measurement-lab.org')
    ndt_server = tmp[0].replace('.','-') + '.measurement-lab.org'
    nodecmd = "ndt7-client " + "-hostname " + ndt_server + " > " + outputname + ".web.csv"
    print(nodecmd)
  subprocess.run([sometabin, "-M=cpu", "-M=mem", "-M=ss:interval=0.1s","-f", metaname, "-c", nodecmd])
        # `sudo $sometabin -M=cpu -M=mem -M=ss:interval=0.1s -f $metaname -c "sudo -u $username node $nodename $outputname $testparam"`;
  subprocess.run(["pkill", "tcpdump"])
  # # display = subprocess.run(["tshark", "-r", pcapname, display_filter[ss_platform]], encoding='utf-8', stdout=subprocess.PIPE)
  # display_in_list = display.stdout.split()
  # #print(display.stdout.split())
  # if (ss_platform == "ookla"):
  #   for i in range(len(display_in_list)):
  #     if display_in_list[i] == "TCP" and display_in_list[i + 4] == "8080":
  #       ip_all[ss_platform].add(display_in_list[i - 1])
  # else:
  #   for i in range(len(display_in_list)):
  #     #if display_in_list[i] == "A":
  #       #print("There is an A")
  #     if display_in_list[i] == "A" and (not re.search('[a-zA-Z]', display_in_list[i + 1])):
  #       ip_all[ss_platform].add(display_in_list[i + 1])
# sleep for 3 seconds before each run
time.sleep(3)

# # conduct traceroute to collected IP addresses
# for sp in range(splen):
#   # ensure own IP address in not falsely included
#   try:
#     ip_all[ss_platform].remove(my_ip)
#   except:
#     pass
#   for addr in list(ip_all[ss_platform]):
#     print("traceroute to "+addr)
#     ip_output_name = speedtest[sp] + "_" + addr + ".txt"
#     with open(ip_output_name, "w") as f:
#       tr_res = subprocess.run(["traceroute", "-I","-n", "-m", "32", addr], encoding='utf-8', stdout=subprocess.PIPE)
#       for hop in tr_res.stdout:
#         f.write(hop)
 
 
def disableinf(link):	
  return "/interface ethernet set " + link + " disabled=yes"
	

def enableinf(link):
  return "/interface ethernet set " + link + " disabled=no"
	

def setbwcmd(bw, link):
  return "/interface ethernet set " + link + " bandwidth=" + bw + "/unlimited"


def setqueuecmd(inf, qt):
  return "/queue interface set " + inf + " queue=" + qt
