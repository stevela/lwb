#!/usr/bin/env python2.7

# Copyright 2011 Steve Lacey
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Create a skeleton post

import argparse
import json
from datetime import datetime
import uuid

class Entry:
  """Represents an entry."""

  def __init__(self, title, body, post_type):
    dt = datetime.now()
    date_str = dt.strftime("%a %b %d %H:%M:%S PST %Y")
    
    self._data = {
      'basename': 'a_base_name',
      'body': body,
      'categories': [ 'Misc' ],
      'format': 'textile',
      'isOldEntry': False,
      'lastModifiedDate': date_str,
      'publishedDate': date_str,
      'status': 'draft',
      'tags': [ 'Misc' ],
      'title': title,
      'type': post_type,
      'uuid': str(uuid.uuid4()),
    }

  def dump(self):
    if False:
      try:
        foo = json.dumps(self._data, sort_keys=True, indent=4)
      except UnicodeDecodeError:
        print self._data['body']
    else:
      out = open("%s.%s" % (self._data['uuid'], self._data['type']), "w")
      out.write(json.dumps(self._data, sort_keys=True, indent=4))
      out.close()


def main():
  parser = argparse.ArgumentParser(description='Create a skeleton post')
  parser.add_argument('-b', '--body', metavar='BODY', dest='body',
                      default='', help='the input body')
  parser.add_argument('-t', '--title', metavar='TITLE', dest='title',
                      default='', help='the title')
  parser.add_argument('-y', '--type', metavar='TYPE', dest='post_type',
                      choices=['post', 'page'], default='post',
                      help='the post type')
  args = parser.parse_args()

  entry = Entry(args.title, args.body, args.post_type)
  entry.dump()


if __name__ == '__main__':
  main()
