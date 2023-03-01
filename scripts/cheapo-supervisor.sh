#!/bin/bash

source $HOME/.bashrc

echodate() {
    echo `date +%Y/%m/%d-%H:%M:%S` $*
}

# Run on cron schedule every minute
echodate "Checking if bot is running..."
ps aux | fgrep "./bitheroes-guide-bot" | grep -v grep >/dev/null && echodate "Bot is already running, exiting..." && exit 0

echodate "Bot is not running, running bot..."
cd $HOME/app/bitheroes-guide-bot
nohup bash -c "./bitheroes-guide-bot >> $HOME/user-logs/bitheroes-guide-bot 2>&1 &"

exit 0
