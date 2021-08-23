#!/usr/bin/python3
# ./generate_from_libwebp.py /path/to/clean/libwebp/src/
import glob
import os
import re
import sys

FILE_KEYS = {}

def fix_headers(libwebp_dir, code_text):
    out = ''
    for line in code_text:
        # Special config.h
        if 'src/webp/config.h' in line:
            out += '#include "config.h"\n'
            continue
        line = line.replace('#include "src/', '#include "' + libwebp_dir + '/')
        header_file = re.match('#include\s+["]([^"]+)["].*', line)
        # regex to search for <, > too
        #header_file = re.match('#include\s+[<"]([^>"]+)[>"].*', line)
        if header_file:
            header = header_file.groups()[0]
            abs_header = os.path.abspath(header)
            header_exists = os.path.exists(abs_header)
            if header_exists and abs_header in FILE_KEYS:
                out += '#include "' + FILE_KEYS[abs_header] + '"\n'
            else:
                out += line + '\n'
        else:
            out += line + '\n'
    return out

if len(sys.argv) != 2:
    print('usage: ./generate_from_libwebp.py /path/to/clean/libwebp/src/')
    os._exit(1)

code = ['.c', '.s', '.S', '.sx', 'cc', 'cpp', 'cpp' ]
header = ['.h', '.hh', '.hpp', '.hxx' ]

# Remove old files
files = os.listdir('.')
for file in files:
    if file.endswith(tuple(code)) or file.endswith(tuple(header)):
        os.remove(os.path.join('.', file))

path = sys.argv[1]

for file in glob.iglob(path + '/**', recursive=True):
    if file.endswith(tuple(code)) or file.endswith(tuple(header)):
        key = os.path.abspath(file)
        val = file.replace(path, '').replace('/', '_')
        FILE_KEYS[key] = val

root_dir = os.path.abspath('.')
libwebp_dir = os.path.abspath(path)
for full_path, local in FILE_KEYS.items():
    os.chdir(os.path.dirname(full_path))
    with open(full_path) as code:
        code_text = code.read().splitlines() 
    code.close()
    fixed = fix_headers(libwebp_dir, code_text)
    os.chdir(root_dir)
    local_file = open(local, "w")
    local_file.write(fixed)
    local_file.close()

# Write config.h
config = '#ifndef WEBANIMATION_HPP\n#define WEBANIMATION_HPP\n'
config += '#define WEBP_USE_THREAD\n'
config += '#endif\n'
config_file = open('config.h', "w")
config_file.write(config)
config_file.close()
