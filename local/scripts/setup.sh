#!/bin/bash

# setup up venv and configures deps
if [[ ! -d "script-venv" ]]; then
    echo "initializing venv"
    python -m venv script-venv
    source script-venv/bin/activate
    pip install -r requirements.txt
    deactivate
else
    echo "venv already exists"
fi
