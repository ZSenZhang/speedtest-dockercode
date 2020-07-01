#class for retreiving data from Amazon, Google cloud and Azure

import boto
from boto.s3.key import Key
import boto.s3.connection

class retreive(interval, vmtype, hostalias):

	def __init__():
		self.vmtype = vmtype
		self.hostalias = hostalias
		self.interval = interval

		


	def retreive(local_dir, remote_dir):
		if self.vmtype == "amazon":
			# to retreive data from aws s3 and then archive retreived data to the old folder
			for key in bucket.list():
    			key.name.encode('utf-8')

		elif self.vmtype == "azure":

		elif self.vmtype == "google":


	def archive(flist,dir):

	def analyze_data(flist,dir):

	def rm_local(flist,dir):

	