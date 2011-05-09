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

# Convert a MoveableType export to json files.

import argparse
import json
import re
import string
import time
import uuid

TIMEZONE = "PST"

class Entry:
  """Represents an entry."""
  next_id = 1

  def __init__(self):
    self._id = Entry.next_id
    Entry.next_id += 1
    self._data = {
      'parsed_num': self._id,
      'uuid': str(uuid.uuid4()),
      'isOldEntry': True,
      'type': 'post'
    }

  def set_attribute(self, key, value):
    self._data[key] = value

  def append_attribute(self, key, value):
    if (key in self._data):
      self._data[key] += value
    else:
      self._data[key] = value

  def append_array_attribute(self, key, value):
    if (key in self._data):
      self._data[key].append(value)
    else:
      self._data[key] = [value]

  def dump(self, output_dir):
    if False:
      try:
        foo = json.dumps(self._data, sort_keys=True, indent=4)
      except UnicodeDecodeError:
        print self._data['body']
    else:
      out = open("%s/%s.post" % (output_dir, self._data['uuid']), "w")
      out.write(json.dumps(self._data, sort_keys=True, indent=4))
      out.close()


def main():

  parser = argparse.ArgumentParser(description='Parse a MoveableType export')
  parser.add_argument('-i', '--input', metavar='INPUT', dest='input',
                      help='the input file')
  parser.add_argument('-o', '--output', metavar='OUTPUT', dest='output',
                      help='the output directory')
  args = parser.parse_args()

  file = open(args.input, 'r')

  entry = Entry()
  in_body = False
  stop_reading = False
  for unstripped_line in file:
    line = string.strip(unstripped_line)
    if line == 'BODY:':
      in_body = True
      continue

    if in_body:
      if line == '-----' or line == '--------':
        in_body = False
        stop_reading = True
        continue
      entry.append_attribute('body', unstripped_line)

    if not in_body and line == '--------':
      entry.dump(args.output)
      entry = Entry()
      stop_reading = False

    if stop_reading:
      continue

    for key in ('title', 'basename', 'category', 'status', 'convert breaks',
                'date', 'tags'):
      search_for = string.upper(key) + ': (?P<value>.*)$'
      m = re.match(search_for, line)
      if m:
        value = string.strip(m.group('value'))

        if key == 'status':
          entry.set_attribute('status', string.lower(value))
        elif key == 'convert breaks':
          if value == '0':
            value = 'none'
          elif value == 'textile_2':
            value = 'textile'
          else:
            value = 'convertbreaks'
          entry.set_attribute('format', value)

        elif key == 'tags':
          tags = []
          for tag in string.split(value, ','):
            tags.append(string.strip(string.strip(tag, '\"')))
          entry.set_attribute('tags', sorted(tags))

        elif key == 'category':
          entry.append_array_attribute('categories', string.lower(value))

        elif key == 'date':
          t = time.strptime(value, "%m/%d/%Y %I:%M:%S %p")
          date_str = time.strftime("%a %b %d %H:%M:%S " + TIMEZONE + " %Y", t)
          entry.set_attribute('publishedDate', date_str)
          entry.set_attribute('lastModifiedDate', date_str)
          
        else:
          entry.set_attribute(key, value)

  file.close()


if __name__ == '__main__':
  main()
