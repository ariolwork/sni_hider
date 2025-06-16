#!/bin/bash

# Проверка наличия аргументов
if [[ -z "$1" ]]; then
    echo "Usage: $0 <file_name>"
    exit 1
fi

CONFIG_FILE="$1"

current_folder="$PWD"
while read -r test_name testfolder; do
    # Игнорируем пустые строки и комментарии
    [[ -z "$test_name" || "$test_name" =~ ^#.* ]] && continue
    cd $testfolder
    tput setaf 1 && echo "----------------------------------------------------------------------------" && tput sgr0
    tput setaf 1 && echo "test  : $test_name" && tput sgr0

    # Запуск Go-тестов
    tput setaf 4 && go test -bench=. -benchmem -count=1 && tput sgr0
    tput setaf 1 && echo "----------------------------------------------------------------------------" && tput sgr0
    cd "$current_folder"
done  < "$CONFIG_FILE" 

if [[ !(-z "$test_name" || "$test_name" =~ ^#.*) ]]; then
    cd $testfolder
    tput setaf 1 && echo "----------------------------------------------------------------------------" && tput sgr0
    tput setaf 1 && echo "test  : $test_name" && tput sgr0

    # Запуск Go-тестов
    tput setaf 4 && go test -bench=. -benchmem -count=1 && tput sgr0
    tput setaf 1 && echo "----------------------------------------------------------------------------" && tput sgr0
    cd "$current_folder"
fi