#!/bin/bash

echo $1
for i in $(seq "$1")
do
	gnome-terminal -- /bin/sh -c "(echo '$i' && sleep 1 && echo '/join soba' && cat) | nc localhost 8888"
done
