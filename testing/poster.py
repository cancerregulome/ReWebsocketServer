import sys
import json
import urllib
import urllib2


def main():
	fo = open(sys.argv[1])
	targetUrl = sys.argv[2]
	
	data = urllib.urlencode({"results":fo.read()})
	urllib2.urlopen(targetUrl,data)


if __name__ == '__main__':
	main()