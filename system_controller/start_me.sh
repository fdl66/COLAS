#!/bin/bash
if [ $1 = "up" ];then
    docker-compose -f me0.yml up -d
    docker-compose -f me1.yml up -d
    docker-compose -f me2.yml up -d
    docker-compose -f me3.yml up -d
elif [ $1 = "stop" ];then
    docker-compose -f me0.yml stop
    docker-compose -f me1.yml stop
    docker-compose -f me2.yml stop
    docker-compose -f me3.yml stop
elif [ $1 = "rm" ];then
    docker-compose -f me0.yml rm -f
    docker-compose -f me1.yml rm -f
    docker-compose -f me2.yml rm -f
    docker-compose -f me3.yml rm -f
fi
