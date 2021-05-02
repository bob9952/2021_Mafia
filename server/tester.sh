#!/bin/bash

for i in $(seq "$1")
do
	#gnome-terminal -- /bin/bash -c "(echo '$i' && sleep 1 && echo '/join soba' && cat) | go run ../client/gui_client.go"
	gnome-terminal -- /bin/bash -c "(echo '$i' && sleep 1 && echo '/join soba' && cat) | nc localhost 8888"
done
