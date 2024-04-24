#!/usr/bin/env python3
import os
import sys
import json
import re

# Install the default signal handler.
from signal import signal, SIGPIPE, SIG_DFL
signal(SIGPIPE, SIG_DFL)

def main():
    if len(sys.argv) < 4:
        print("Usage: " + sys.argv[0] + " input.pdb template_domain_names.json stoichiometry", file=sys.stderr)
        sys.exit(1)

    pdb = sys.argv[1]
    # use stdin
    if pdb == "-":
        pdb = 0

    stoic_str = sys.argv[3]
    if stoic_str == "":
        stoic_str = "A1"
    stoic = []
    search_stoic = re.compile('([A-Za-z]+)([0-9]+)')
    for match in search_stoic.finditer(stoic_str):
        for _ in range(0, int(match.group(2))):
            stoic.append(match.group(1))

    names_file = sys.argv[2]
    try:
        with open(names_file) as f:
            names = json.load(f)
        names = [ [key, *value] for (key, value) in names.items()]
    except FileNotFoundError:
        names = []
        for s in stoic:
            names.append([s[0]])

    with open(pdb) as f:
        in_record = False
        prev = ""
        record = 0
        for line in f:
            fields = line.split()
            if fields[0] == "ATOM":
                chain = fields[4]

            if in_record == False and fields[0] == "ATOM":
                cur_stoic = stoic[record]
                cur_names = names[record]
                record = record + 1
                in_record = True
                if cur_stoic != prev:
                    if len(cur_names) > 1:
                        print("PARENT " + " ".join(cur_names[1:6]))
                    else:
                        print("PARENT N/A")
                    prev = cur_stoic

            if fields[0] == "TER" or fields[0] == "ENDMDL":
                in_record = False
   
            print(line, end="")

main()
