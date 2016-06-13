#!/usr/bin/python3

import http.server
import os.path
import subprocess
import urllib.parse


CURRENT_DIR = os.path.dirname(os.path.realpath(__file__))
ARCHIVE_FILE = os.path.join(CURRENT_DIR, 'archive')
SERVER_ADDRESS = ('', 0)


def main():
    http.server.HTTPServer(SERVER_ADDRESS, YTDownloadHTTPRequestHandler).serve_forever()


class YTDownloadHTTPRequestHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        """ Serve a GET request: download YT song """
        download_cmd = [
            'youtube-dl',
            '--no-playlist',
            '--download-archive', ARCHIVE_FILE,
            '--extract-audio',
            '--audio-quality', '0',
            '--output', '/home/loamhoof/Downloads/%(title)s.%(ext)s',
            urllib.parse.unquote(self.path[1:]),
        ]
        subprocess.Popen(download_cmd)



if __name__ == '__main__':
    main()
