#!/usr/bin/env bash


CHECK_LIST=(
  'black hat'
  'blacklist'
  'blackout'
  'brownout'
  'cakewalk'
  'disable'
  'female'
  'grandfathered'
  'handicap'
  'he'
  'him'
  'his'
  'kill'
  'male'
  'rule of thumb'
  'sanity test'
  'sanity check'
  'segregate'
  'segregation'
  'she'
  'her'
  'hers'
  'slave'
  'suffer'
  'war room'
  'white hat'
  'whitelist'
)

for check in "${CHECK_LIST[@]}"
do
  if grep -riwI "$check" --exclude ./hack/inclusive-terminology.sh; then
    echo found "$check"
  fi
done


if grep -riwI "master" --exclude ./hack/inclusive-terminology.sh --exclude ./.git/hooks/pre-rebase.sample | grep -wv "blob/master"; then
  echo found "master"
fi




