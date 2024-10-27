#!/usr/bin/env bash

for clfg in {30..37} ; do \
  echo -en "\e[${clfg}m\t^[${clfg}m \e[0m"; \
done;
echo
for clfg in {90..97} ; do \
  echo -en "\e[${clfg}m\t^[${clfg}m \e[0m"; \
done;
echo
for clbg in {40..47} ; do \
  echo -en "\e[${clbg}m\t^[${clbg}m \e[0m"; \
done
echo
for clbg in {100..107} ; do \
  echo -en "\e[${clbg}m\t^[${clbg}m \e[0m"; \
done
echo
for attr in 1 2 4 5 7 ; do \
  echo -en "\e[31;${attr}m\t^[31;${attr}m \e[0m"; \
done
echo
