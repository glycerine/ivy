#!/bin/bash

deactivate # exit any old venv first! (if any)
python3 -m venv goivy-venv
source goivy-venv/bin/activate
pip3 install -r requirements.txt
pip3 install 'setuptools<70.0.0' ## depends on older version

cd ivy
pip3 uninstall ms_ivy 
pip3 install -e .
