#!/bin/bash

if command -v monom >/dev/null 2>&1; then
    echo "monom is already installed, try running it by typing \"monom\""
else
    if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
        echo "Please source this script insteaf of calling it:"
        echo "source $0"
    else
        DIR="$(cd "$(dirname "$0")" && pwd)"
        alias_string="$(printf "source \"%s/src/monom\"\n" "$DIR")"

        if [ "$SHELL" = "/bin/zsh" ]; then
            echo "Installing monom..."
            echo "$alias_string" >> ~/.zshrc
            # shellcheck disable=SC1090
            . ~/.zshrc
        elif [ "$SHELL" = "/bin/bash" ]; then
            echo "Installing monom..."
            echo "$alias_string" >> ~/.bashrc
            # shellcheck disable=SC1090
            . ~/.bashrc
        else
            echo "Add this to your rc or profile file:"
            echo "$alias_string"
        fi
    fi
fi
