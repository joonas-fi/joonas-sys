#!/bin/bash -eu

cd /sysroot/apps/deployer

# deployments/joonas.fi-blog/ => joonas.fi-blog
# joonas.fi-blog => (keep as-is)
# service="$(basename $1)"

service="$(ls -1 deployments/ | rofi -dmenu -p 'Service')"

source env-prod.env

# need this because we're not in a terminal
i3-sensible-terminal --command rel/deployer_linux-amd64 deploy "$service" ""

# rel/deployer_linux-amd64 deploy "$service" ""

notify-send "Deployed: $service"
