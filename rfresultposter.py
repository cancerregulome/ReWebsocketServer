import sys
import json
import urllib
import urllib2


def main():
	fo = open(sys.argv[1])
	targetUrl = sys.argv[2]
	
	urllib2.urlopen(targetUrl,fo.read())


if __name__ == '__main__':
	main()