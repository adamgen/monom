#!/usr/bin/env python3
# Wraps, in another language: runs the rest of the chain and passes its status on.
import subprocess
import sys

print("inner: before", flush=True)
status = subprocess.call(sys.argv[1:])
print(f"inner: after (exit {status})", flush=True)
sys.exit(status)
